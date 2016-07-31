package plist

import (
	"bytes"
	"encoding/xml"
	"io"
	"reflect"
	"strconv"
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

	// first error that appear
	err error

	// token obtain by peekToken
	peekToken interface{}
}

// Decode is something like Unmarshal, but we have an object store status information
func (d *Decoder) Decode(v interface{}) error {
	// TODO: recover from panic
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		d.setError(ErrMustBePointer)
		return d.err
	}
	d.decodeValue(reflect.Indirect(val))
	return d.err
}

// Look to next token, but don't move ahead
func (d *Decoder) PeekNextToken() xml.Token {
	var err error
	if d.peekToken == nil {
		d.peekToken, err = d.d.Token()
		if err != nil {
			d.setError(err)
		}
	}
	return d.peekToken
}

// Return nextToken
func (d *Decoder) NextToken() xml.Token {
	if d.peekToken != nil {
		tmp := d.peekToken
		d.peekToken = nil
		return tmp
	}
	t, err := d.d.Token()
	if err != nil {
		d.setError(err)
	}
	return t
}

func (d *Decoder) decodeValue(v reflect.Value) {
	switch v.Kind() {
	default:
		d.setError(&CannotParseTypeError{v})
	case reflect.String:
		d.startElement("string")
		v.SetString(string(d.charData()))
		d.endElement("string")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.startElement("integer")
		num, err := strconv.ParseInt(string(d.charData()), 10, v.Type().Bits())
		if err != nil {
			d.setError(err)
		} else {
			v.SetInt(num)
		}
		d.endElement("integer")
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		d.startElement("integer")
		num, err := strconv.ParseUint(string(d.charData()), 10, v.Type().Bits())
		if err != nil {
			d.setError(err)
		} else {
			v.SetUint(num)
		}
		d.endElement("integer")
	case reflect.Float32, reflect.Float64:
		d.startElement("real")
		num, err := strconv.ParseFloat(string(d.charData()), v.Type().Bits())
		if err != nil {
			d.setError(err)
		} else {
			v.SetFloat(num)
		}
		d.endElement("real")
	case reflect.Bool:
		t := d.PeekNextToken()
		switch t := t.(type) {
			case xml.StartElement:
				if t.Name.Local == "true" {
					v.SetBool(true)
					d.startElement("true")
					d.endElement("true")
					return
				}
			}
		v.SetBool(false)
		d.startElement("false")
		d.endElement("false")
	}
}

// setError set d.err but only if is empty
func (d *Decoder) setError(e error) {
	if d.err == nil {
		d.err = e
	}
}

// verify, if it next token really is startElement if given name
func (d *Decoder) startElement(name string) {
	t := d.NextToken()
	switch t := t.(type) {
	default:
		d.setError(NewUnexpectedTokenError("<"+name+">", t))
	case xml.StartElement:
		if t.Name.Local != name {
			d.setError(NewUnexpectedTokenError("<"+name+">", t))
		}
	}
}

// verify, if it next token really is endElement if given name
func (d *Decoder) endElement(name string) {
	t := d.NextToken()
	switch t := t.(type) {
	default:
		d.setError(NewUnexpectedTokenError("</"+name+">", t))
	case xml.EndElement:
		if t.Name.Local != name {
			d.setError(NewUnexpectedTokenError("</"+name+">", t))
		}
	}
}

// verify, if it next token really charData and return that data
func (d *Decoder) charData() []byte {
	t := d.NextToken()
	switch t := t.(type) {
	default:
		d.setError(NewUnexpectedTokenError("CharData", t))
		return nil
	case xml.CharData:
		return t
	}
}
