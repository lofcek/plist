package plist

import (
	"bytes"
	"encoding/xml"
	"io"
	"reflect"
	"strconv"
)

func Unmarshal(b []byte, v interface{}) error {
	d:=NewDecoder(bytes.NewReader(b))
	return d.Decode(v)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	d := xml.NewDecoder(r)
	d.Strict = true
	return &Decoder {
		d: xml.NewDecoder(r),
		err: nil,
	}
}

// Decoder could be used to parse plist into pointer of required type
type Decoder struct {
	d *xml.Decoder
	err error
}


// Decode is something like Unmarshal, but we have an object store status information
func (d* Decoder) Decode(v interface{}) error {
	// TODO: recover from panic
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		d.setError(ErrMustBePointer)
		return d.err
	}
	d.decodeValue(reflect.Indirect(val))
	return d.err
}


func (d* Decoder) decodeValue(v reflect.Value) {
	switch(v.Kind()) {
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
	}
}

// setError set d.err but only if is empty
func (d *Decoder) setError(e error) {
	if d.err == nil {
		d.err = e
	}
}


func (d *Decoder) startElement(name string) {
	t, err := d.d.Token()
	if err != nil {
		d.setError(err)
		return
	}
	switch t := t.(type) {
		default:
			d.setError(NewUnexpectedTokenError("<" + name + ">", t))
		case xml.StartElement:
			if t.Name.Local != name {
				d.setError(NewUnexpectedTokenError("<" + name + ">", t))
			}
	}
}

func (d *Decoder) endElement(name string) {
	t, err := d.d.Token()
	if err != nil {
		d.setError(err)
		return
	}
	switch t := t.(type) {
		default:
			d.setError(NewUnexpectedTokenError("</" + name + ">", t))
		case xml.EndElement:
			if t.Name.Local != name {
				d.setError(NewUnexpectedTokenError("</" + name + ">", t))
			}
	}
}

func (d *Decoder) charData() []byte {
	t, err := d.d.Token()
	if err != nil {
		d.setError(err)
		return nil
	}
	switch t := t.(type) {
		default:
			d.setError(NewUnexpectedTokenError("CharData", t))
			return nil
		case xml.CharData:
			return t
	}
}