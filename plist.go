package plist

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"io"
	"reflect"
	"strconv"
	"time"
	//"fmt"
)

func Unmarshal(b []byte, v interface{}) error {
	d := NewDecoder(bytes.NewReader(b))
	return d.Decode(v)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	d := xml.NewDecoder(r)
	d.Strict = true
	return &Decoder{
		d:   xml.NewDecoder(r),
		err: nil,
	}
}

// Decoder could be used to parse plist into pointer of required type
// Advantage Unmarshal is that any io.Reader can be initialized
type Decoder struct {
	// internal data struct - d embeeded decoder
	d *xml.Decoder

	// here we store the first error that appears
	err error
}

// Decode is something like Unmarshal, but we have an object store status information
func (d *Decoder) Decode(v interface{}) error {
	// TODO: recover from panic
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		d.setError(ErrMustBePointer)
		return d.err
	}
	d.decode(reflect.Indirect(val))
	return d.err
}

func (d *Decoder) DecodeElement(v interface{}, start *xml.StartElement) {
	if d.err == nil {
		err := d.d.DecodeElement(v, start)
		if err != nil {
			d.setError(err)
		}
	}
}

// Return next Token
func (d *Decoder) Token() xml.Token {
	if d.err != nil {
		return io.EOF
	}
	t, err := d.d.Token()
	if err != nil {
		d.setError(err)
	}
	return t
}

func (d *Decoder) firstNotEmptyToken() (xml.Token, error) {
	for {
		t, err := d.d.Token()
		if err != nil {
			return t, d.setError(err)
		}
		switch t := t.(type) {
		case xml.Comment:
			continue
		case xml.CharData:
			if len(bytes.TrimSpace([]byte(t))) == 0 {
				continue
			}
		default:
			return t, nil
		}
	}
}

func (d *Decoder) decode(v reflect.Value) error {
	for {
		t, err := d.firstNotEmptyToken()
		if err != nil {
			return d.setError(err)
		}
		switch se := t.(type) {
		case xml.StartElement:
			return d.decodeElement(v, se)
		default:
			return d.setError(NewUnexpectedTokenError("<any token>", t))
		}
	}
}

func (d *Decoder) decodeElement(v reflect.Value, se xml.StartElement) error {
	switch v.Kind() {
	default:
		return d.setError(&CannotParseTypeError{v})
	case reflect.String:
		if se.Name.Local != "string" {
			return d.setError(NewUnexpectedTokenError("<string>", se))
		}
		var s string
		d.DecodeElement(&s, &se)
		v.SetString(s)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if se.Name.Local != "integer" {
			return d.setError(NewUnexpectedTokenError("<integer>", se))
		}
		var s string
		d.DecodeElement(&s, &se)
		num, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return d.setError(err)
		}
		v.SetInt(num)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if se.Name.Local != "integer" {
			return d.setError(NewUnexpectedTokenError("<integer>", se))
		}
		var s string
		d.DecodeElement(&s, &se)
		num, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return d.setError(err)
		}
		v.SetUint(num)
		return nil
	case reflect.Float32, reflect.Float64:
		if se.Name.Local != "real" {
			return d.setError(NewUnexpectedTokenError("<real>", se))
		}
		var s string
		d.DecodeElement(&s, &se)
		num, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return d.setError(err)
		}
		v.SetFloat(num)
		return nil
	case reflect.Bool:
		if se.Name.Local != "true" && se.Name.Local != "false" {
			return d.setError(NewUnexpectedTokenError("<true> or <false>", se))
		}
		v.SetBool(se.Name.Local == "true")
		return nil
	case reflect.Slice:
		if se.Name.Local != "array" {
			return d.setError(NewUnexpectedTokenError("<array>", se))
		}
		v.SetLen(0)
		for {
			t, err := d.d.Token()
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
			case xml.EndElement:
				return nil
			default:
				continue
			}
		}
	case reflect.Struct:
		var t time.Time
		writerType := reflect.TypeOf((*io.Writer)(nil)).Elem()
		if v.Type() == reflect.TypeOf(t) {
			// parse it like date
			if se.Name.Local != "date" {
				return d.setError(NewUnexpectedTokenError("<date>", se))
			}
			var s string
			d.DecodeElement(&s, &se)
			tm, err := time.Parse("2006-01-02T15:04:05Z", string(s))
			if err != nil {
				return d.setError(err)
			}
			v.Set(reflect.ValueOf(tm))
			return nil
		} else if v.Addr().Type().Implements(writerType) {
			if se.Name.Local != "data" {
				return d.setError(NewUnexpectedTokenError("<data>", t))
			}
			var data []byte
			d.DecodeElement(&data, &se)
			buf := v.Addr().Interface().(io.Writer)
			decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data))
			_, err := io.Copy(buf, decoder)
			if err != nil {
				return d.setError(err)
			}
			return nil
		}
		return d.setError(&CannotParseTypeError{v})
	}
}

// setError set d.err but only if is empty
func (d *Decoder) setError(e error) error {
	if d.err == nil {
		d.err = e
	}
	return d.err
}

// verify, if it next token really is endElement if given name
/*func (d *Decoder) endElement() {
	t := d.Token()
	switch t := t.(type) {
	default:
		d.setError(NewUnexpectedTokenError("</"+name+">", t))
	case xml.EndElement:
		if t.Name.Local != name {
		d.setError(NewUnexpectedTokenError("</"+name+">", t))
		}
	}
}*/

// verify, if it next token really charData and return that data
/*func (d *Decoder) charData() []byte {
	t := d.Token()
	switch t := t.(type) {
	default:
		d.setError(NewUnexpectedTokenError("CharData", t))
		return nil
	case xml.CharData:
		return t
	}
}*/
