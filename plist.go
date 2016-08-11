package plist

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	// "fmt"
)

func Unmarshal(b []byte, v interface{}) error {
	return UnmarshalWithErrCallback(b, v, nil)
}

func UnmarshalWithErrCallback(b []byte, v interface{}, errCallback func(error)) error {
	d := NewDecoder(bytes.NewReader(b))
	d.OnError = errCallback
	err := d.Decode(v)
	if err != nil {
		return err
	}
	t, err := d.firstNotEmptyToken()
	if err != io.EOF {
		if err != nil {
			return err
		}
		return d.callOnError(NewUnexpectedTokenError("EOF", t, d.d.InputOffset()))
	}
	return nil
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	d := xml.NewDecoder(r)
	d.Strict = true
	return &Decoder{
		d:       xml.NewDecoder(r),
		OnError: nil,
	}
}

// Decoder could be used to parse plist into pointer of required type
// Advantage Unmarshal is that any io.Reader can be initialized
type Decoder struct {
	// internal data struct - d embeeded decoder
	d *xml.Decoder

	// function, that should not be called, if everything is decoded without error.
	// If problem appears it should be called just once on place nearest to place, where error appear.
	// Policy how to do that is:
	// - if error appears in method called outside of Decoder, call d.callOnError(err)
	// - if error appears in Decoder's method, return error, but don't call d.callOnError second time
	OnError func(error)
}

// Decode is something like Unmarshal, but we have an object store status information
func (d *Decoder) Decode(v interface{}) error {
	// TODO: recover from panic
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return d.callOnError(ErrMustBePointer)
	}
	return d.decode(reflect.Indirect(val))
}

func (d *Decoder) DecodeElement(v interface{}, start *xml.StartElement) error {
	return d.callOnError(d.d.DecodeElement(v, start))
}

func (d *Decoder) firstNotEmptyToken() (xml.Token, error) {
	for {
		t, err := d.d.Token()
		if err == io.EOF {
			// for io.EOF we should not call OnError callback
			return t, err
		}
		if err != nil {
			return t, d.callOnError(err)
		}
		switch t := t.(type) {
		case xml.Comment, xml.ProcInst, xml.Directive:
			continue
		case xml.CharData:
			if len(bytes.TrimSpace([]byte(t))) == 0 {
				continue
			}
			return t, nil
		default:
			return t, nil
		}
	}
}

func (d *Decoder) decode(v reflect.Value) error {
	for {
		t, err := d.firstNotEmptyToken()
		if err != nil {
			return err
		}
		switch se := t.(type) {
		case xml.StartElement:
			return d.decodeElement(v, se)
		default:
			return d.callOnError(NewUnexpectedTokenError("<any token>", t, d.d.InputOffset()))
		}
	}
}

func (d *Decoder) decodeElement(v reflect.Value, se xml.StartElement) error {
	switch v.Kind() {
	default:
		return d.callOnError(&CannotParseTypeError{v})
	case reflect.String:
		if se.Name.Local != "string" {
			return d.callOnError(NewUnexpectedTokenError("<string>", se, d.d.InputOffset()))
		}
		var s string
		err := d.DecodeElement(&s, &se)
		if err != nil {
			return err
		}
		v.SetString(s)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if se.Name.Local != "integer" {
			return d.callOnError(NewUnexpectedTokenError("<integer>", se, d.d.InputOffset()))
		}
		var s string
		err := d.DecodeElement(&s, &se)
		if err != nil {
			return err
		}
		num, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return d.callOnError(err)
		}
		v.SetInt(num)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if se.Name.Local != "integer" {
			return d.callOnError(NewUnexpectedTokenError("<integer>", se, d.d.InputOffset()))
		}
		var s string
		err := d.DecodeElement(&s, &se)
		if err != nil {
			return err
		}
		num, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return d.callOnError(err)
		}
		v.SetUint(num)
		return nil
	case reflect.Float32, reflect.Float64:
		if se.Name.Local != "real" {
			return d.callOnError(NewUnexpectedTokenError("<real>", se, d.d.InputOffset()))
		}
		var s string
		err := d.DecodeElement(&s, &se)
		if err != nil {
			return err
		}
		num, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return d.callOnError(err)
		}
		v.SetFloat(num)
		return nil
	case reflect.Bool:
		if se.Name.Local != "true" && se.Name.Local != "false" {
			return d.callOnError(NewUnexpectedTokenError("<true> or <false>", se, d.d.InputOffset()))
		}
		v.SetBool(se.Name.Local == "true")
		var s struct{}
		err := d.d.DecodeElement(&s, &se)
		if err != nil {
			return d.callOnError(err)
		}
		return nil
	case reflect.Slice:
		if se.Name.Local != "array" {
			return d.callOnError(NewUnexpectedTokenError("<array>", se, d.d.InputOffset()))
		}
		v.SetLen(0)
		for {
			t, err := d.firstNotEmptyToken()
			if err != nil {
				return err
			}
			switch se := t.(type) {
			case xml.StartElement:
				newVal := reflect.Zero(v.Type().Elem())
				v.Set(reflect.Append(v, newVal))
				err := d.decodeElement(v.Index(v.Len()-1), se)
				if err != nil {
					return err
				}
				continue
			case xml.EndElement:
				// todo testing wheather endElement is really </array>
				return nil
			default:
				return d.callOnError(NewUnexpectedTokenError("</array>", se, d.d.InputOffset()))
			}
		}
	case reflect.Struct:
		var t time.Time
		writerType := reflect.TypeOf((*io.Writer)(nil)).Elem()
		if v.Type() == reflect.TypeOf(t) {
			// parse it like date
			if se.Name.Local != "date" {
				return d.callOnError(NewUnexpectedTokenError("<date>", se, d.d.InputOffset()))
			}
			var s string
			err := d.d.DecodeElement(&s, &se)
			if err != nil {
				return d.callOnError(err)
			}
			tm, err := time.Parse("2006-01-02T15:04:05Z", string(s))
			if err != nil {
				return d.callOnError(err)
			}
			v.Set(reflect.ValueOf(tm))
			return nil
		} else if v.Addr().Type().Implements(writerType) {
			if se.Name.Local != "data" {
				return d.callOnError(NewUnexpectedTokenError("<data>", t, d.d.InputOffset()))
			}
			var data []byte
			err := d.DecodeElement(&data, &se)
			if err != nil {
				return err
			}
			buf := v.Addr().Interface().(io.Writer)
			decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data))
			_, err = io.Copy(buf, decoder)
			if err != nil {
				return d.callOnError(err)
			}
			return nil
		} else {
			if se.Name.Local != "dict" {
				return d.callOnError(NewUnexpectedTokenError("<dict>", se, d.d.InputOffset()))
			}
			// this struct have to decoded to members
			for {
				is_key, key, err := d.decodeKey(se)
				if err != nil {
					return err
				}
				if !is_key {
					return nil
				}
				f, ok := getFieldByName(key, v)
				//fmt.Printf("getFieldByName %s returns %#v, %t\n", key, f, ok)
				if ok {
					err = d.decode(f)
					if err != nil {
						return err
					}
				} else {
					t, err := d.firstNotEmptyToken()
					if err != nil {
						return err
					}
					if _, ok := t.(xml.StartElement); !ok {
						return d.callOnError(NewUnexpectedTokenError("<any key>", &t, d.d.InputOffset()))
					}
					err = d.d.Skip()
					if err != nil {
						return d.callOnError(err)
					}
				}
			}
			return nil
		}
		return d.callOnError(&CannotParseTypeError{v})
	case reflect.Ptr:
		v.Set(reflect.New(v.Type().Elem()))
		return d.decodeElement(reflect.Indirect(v), se)
	case reflect.Map:
		v2 := reflect.MakeMap(v.Type())
		for {
			is_key, key, err := d.decodeKey(se)
			if err != nil {
				return err
			}
			if !is_key {
				v.Set(v2)
				return nil
			}
			p := reflect.New(v.Type().Elem())
			err = d.decode(reflect.Indirect(p))
			if err != nil {
				return err
			}
			v2.SetMapIndex(reflect.ValueOf(key), reflect.Indirect(p))
		}
	case reflect.Interface:
		switch se.Name.Local {
		default:
			return d.callOnError(&CannotParseTypeError{v})
		case "true", "false":
			var b bool
			return d.decodeInterface(reflect.ValueOf(&b), se, v)
		case "integer":
			var i int64
			return d.decodeInterface(reflect.ValueOf(&i), se, v)
		case "real":
			var f float64
			return d.decodeInterface(reflect.ValueOf(&f), se, v)
		case "string":
			var s string
			return d.decodeInterface(reflect.ValueOf(&s), se, v)
		case "date":
			var t time.Time
			return d.decodeInterface(reflect.ValueOf(&t), se, v)
		case "data":
			var buf bytes.Buffer
			return d.decodeInterface(reflect.ValueOf(&buf), se, v)
		case "array":
			var arr []interface{}
			return d.decodeInterface(reflect.ValueOf(&arr), se, v)
		case "dict":
			var m map[string]interface{}
			return d.decodeInterface(reflect.ValueOf(&m), se, v)
		}
	}
}

func (d *Decoder) decodeInterface(i reflect.Value, se xml.StartElement, v reflect.Value) error {
	err := d.decodeElement(reflect.Indirect(i), se)
	if err != nil {
		return err
	}
	v.Set(reflect.Indirect(i))
	return nil
}

func (d *Decoder) decodeKey(se xml.StartElement) (is_key bool, key string, err error) {
	t, err := d.firstNotEmptyToken()
	if err != nil {
		return false, "", err
	}

	switch t := t.(type) {
	default:
		return false, "", d.callOnError(NewUnexpectedTokenError("</"+se.Name.Local+">", t, d.d.InputOffset()))
	case xml.EndElement:
		return false, key, nil // everything was parsed fine
	case xml.StartElement:
		if t.Name.Local != "key" {
			return false, "", d.callOnError(NewUnexpectedTokenError("<key>", &t, d.d.InputOffset()))
		}
		var key string
		err = d.d.DecodeElement(&key, &t)
		if err != nil {
			return false, "", d.callOnError(err)
		}
		return true, key, nil
	}
}

func ScanCommaFields(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func getFieldByName(name string, v reflect.Value) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("plist")
		arr := strings.Split(tag, ",")
		if len(arr) > 0 && arr[0] == name {
			return v.Field(i), true
		}
	}
	s, ok := v.Type().FieldByName(name)
	if ok && strings.HasPrefix(s.Tag.Get("plist"), "-") {
		return v, false
	}
	return v.FieldByName(name), ok
}

func (d *Decoder) callOnError(err error) error {
	if err != nil && d.OnError != nil {
		d.OnError(err)
	}
	return err
}
