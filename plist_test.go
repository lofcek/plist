package plist

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"strconv"
	"testing"
	"time"
	//"fmt"
)

func TestUnmarshalPlist(t *testing.T) {
	var s string
	var i int
	var i8 int8
	var i64 int64
	var u16 uint16
	var f32 float32
	var up uintptr
	var b bool
	var tm time.Time
	var buf bytes.Buffer
	var af32 []float32
	var ai []int
	var pi *int
	var i4 int = 4

	type S1 struct { // structure without tags
		I int
		B bool
	}
	var s1 S1
	var ps1 *S1

	type S2 struct { // struct with and without tags
		// plist shold swap names A and X
		A int `plist:"C"`
		B int
		C int `plist:"A"`
		D int `plist:"-"`
	}
	var s2 S2
	var iface interface{}
	//var piface *interface{}
	var m1 map[string]interface{}
	var m2 map[string]bool
	var pm1 *map[string]interface{}

	type TestUnmarshal struct {
		name string
		xml  string
		pvar interface{}
		test TestUnmarshaler
	}

	test_cases := [...]TestUnmarshal{
		// decode primitive types
		{"true", `<true/>`, &b, UnmarshalExpectsEq{true}},
		{"false", `<false/>`, &b, UnmarshalExpectsEq{false}},
		{"string", `<string>a</string>`, &s, UnmarshalExpectsEq{"a"}},
		{"string_escape", `<string>&lt;&gt;</string>`, &s, UnmarshalExpectsEq{"<>"}},
		{"unmarshalInt", `<integer>42</integer>`, &i, UnmarshalExpectsEq{int(42)}},
		{"unmarshalInt64", `<integer>42</integer>`, &i64, UnmarshalExpectsEq{int64(42)}},
		{"unmarshalOverflow", `<integer>256</integer>`, &i8, UnmarshalExpectsError{(*strconv.NumError)(nil)}},
		{"wrong type", `<integer>256</integer>`, &s, UnmarshalExpectsError{(*UnexpectedTokenError)(nil)}},
		{"unmarshalUInt16", `<integer>10</integer>`, &u16, UnmarshalExpectsEq{uint16(10)}},
		{"unmarshalUIntPtr", `<integer>10</integer>`, &up, UnmarshalExpectsEq{uintptr(10)}},
		{"unmarshalToChan", `<integer>10</integer>`, new(chan int), UnmarshalExpectsError{(*CannotParseTypeError)(nil)}},
		{"unmarshalFloat", `<real>3.14</real>`, &f32, UnmarshalExpectsEq{float32(3.14)}},
		// spaces could be skipped, any not empty text should cause panic
		{"true again", `<true/>  `, &b, UnmarshalExpectsEq{true}},
		{"not empty space", `<true/>  aa`, &b, UnmarshalExpectsError{(*UnexpectedTokenError)(nil)}},

		// special types like date or data
		{"date", `<date>2016-05-04T03:02:01Z</date>`, &tm, UnmarshalExpectsEq{time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)}},
		{"base64data", `<data>aGVsbG8=</data>`, &buf, UnmarshalExpectsEq{*bytes.NewBuffer([]byte("hello"))}},

		// arrays
		{"array", `<array><integer>4</integer><integer>2</integer></array>`, &ai, UnmarshalExpectsEq{[]int{4, 2}}},
		{"wrong xml", `<array></integer>`, &ai, UnmarshalExpectsError{(*xml.SyntaxError)(nil)}},
		{"comments", ` <!-- use spaces and comments inside--> <array><!-- --><real>4</real> <real>2</real><!-- --> </array> <!-- -->`, &af32, UnmarshalExpectsEq{[]float32{4, 2}}},
		{"pointer", `<integer>4</integer>`, &pi, UnmarshalExpectsEq{&i4}},

		// dictionaries
		{"dicts", `<dict><key>B</key><true/><key>I</key><integer>42</integer></dict>`, &s1, UnmarshalExpectsEq{S1{42, true}}},
		{"dict_ptr", `<dict><key>B</key><true/><key>I</key><integer>42</integer></dict>`, &ps1, UnmarshalExpectsEq{&S1{42, true}}},
		{"tags", `<dict><key>B</key><integer>1</integer><key>A</key><integer>2</integer><key>C</key><integer>3</integer></dict>`, &s2, UnmarshalExpectsEq{S2{B: 1, C: 2, A: 3, D: 0}}},
		{"dict, omit field", `<dict><key>B</key><integer>1</integer><key>A</key><integer>2</integer><key>C</key><integer>3</integer><key>D</key><integer>5</integer></dict>`, &s2, UnmarshalExpectsEq{S2{B: 1, C: 2, A: 3, D: 0}}},
		{"empty dict", `<dict></dict>`, &s1, UnmarshalExpectsEq{S1{0, false}}},
		{"wrong tag", `<not-dict></not-dict>`, &s1, UnmarshalExpectsError{(*UnexpectedTokenError)(nil)}},
		{"xml directive", `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist SYSTEM "file://localhost/System/Library/DTDs/PropertyList.dtd"><true/>`, &b, UnmarshalExpectsEq{true}},

		// decode into interface{}
		{"trueToIface", `<true/>`, &iface, UnmarshalExpectsEq{true}},
		{"intToiface", `<integer>42</integer>`, &iface, UnmarshalExpectsEq{int64(42)}},
		{"realToIface", `<real>42</real>`, &iface, UnmarshalExpectsEq{float64(42)}},
		{"stringToIface", `<string>hello</string>`, &iface, UnmarshalExpectsEq{"hello"}},
		{"dateToIface", `<date>2016-05-04T03:02:01Z</date>`, &iface, UnmarshalExpectsEq{time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)}},
		{"b64ToIface", `<data>aGVsbG8=</data>`, &iface, UnmarshalExpectsEq{*bytes.NewBuffer([]byte("hello"))}},
		{"iface_map_iface", `<dict><key>x</key><true/><key>y</key><false/></dict>`, &m1, UnmarshalExpectsEq{map[string]interface{}{"x": true, "y": false}}},
		{"iface_map_bool", `<dict><key>x</key><true/><key>y</key><false/></dict>`, &m2, UnmarshalExpectsEq{map[string]bool{"x": true, "y": false}}},
		{"dictToIface", `<dict><key>x</key><true/><key>y</key><false/></dict>`, &iface, UnmarshalExpectsEq{map[string]interface{}{"x": true, "y": false}}},
		//{"dictToPIface", `<dict><key>x</key><true/><key>y</key><false/></dict>`, &piface, UnmarshalExpectsEq{&map[string]interface{}{"x": true, "y": false}}},
		{"pmap_iface", `<dict><key>x</key><true/><key>y</key><false/></dict>`, &pm1, UnmarshalExpectsEq{&map[string]interface{}{"x": true, "y": false}}},
	}

	for _, c := range test_cases {
		// set c.pvar to zero before test starts
		v := reflect.Indirect(reflect.ValueOf(c.pvar))
		v.Set(reflect.Zero(v.Type()))
		//c.test.TestUnmarshal(t, c.xml, c.pvar)
		runTest(c.name, func(t *testing.T) {c.test.TestUnmarshal(t, c.xml, c.pvar)},t)
	}
}

type TestUnmarshaler interface {
	TestUnmarshal(t *testing.T, xml_text string, v interface{})
}

type UnmarshalExpectsEq struct {
	expected_val interface{}
}

func (res UnmarshalExpectsEq) TestUnmarshal(t *testing.T, xml_text string, v interface{}) {
	var errors []error
	err := UnmarshalWithErrCallback([]byte(xml_text), v, func(err error) { errors = append(errors, err) })
	if err != nil {
		t.Errorf("Unmarshaling of %q into %T unexpectedly failed: %#s", xml_text, v, err)
		return
	}

	val := reflect.Indirect(reflect.ValueOf(v)).Interface()
	if !reflect.DeepEqual(val, res.expected_val) {
		t.Errorf("Unmarshaling of %q into %T should return %#v not %#v", xml_text, v, res.expected_val, val)
		return
	}

	if len(errors) != 0 {
		t.Errorf("Unmarshaling of %q into %T should not call any OnError callback, but was called %d time(s) with arguments %#v", xml_text, v, len(errors), errors)
		return
	}
}

type UnmarshalExpectsError struct {
	errType interface{
		error
	}
}

func (exp UnmarshalExpectsError) TestUnmarshal(t *testing.T, xml_text string, v interface{}) {
	var errors []error
	err := UnmarshalWithErrCallback([]byte(xml_text), v, func(err error) { errors = append(errors, err) })
	if reflect.TypeOf(err) != reflect.TypeOf(exp.errType) {
		t.Errorf("Unmarshaling of %q into %T expected error %#v, but got %#v", xml_text, v, exp.errType, err)
	}
	if len(errors) != 1 || errors[0] != err {
		t.Errorf("Unmarshaling of %q into %T should call OnError callback just one with %#v, but it was called %d times (%#v)", xml_text, v, err, len(errors), errors)
	}
}
