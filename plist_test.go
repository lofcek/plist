package plist

import (
	"bytes"
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

	type S1 struct {	// structure without tags
		I int
		B bool
	}
	var s1 S1

	type TestUnmarshal struct {
		run  bool
		xml  string
		pvar interface{}
		test TestUnmarshaler
	}

	test_cases := []TestUnmarshal{
		{true, `<string>a</string>`, &s, UnmarshalExpectsEq{"a"}},
		{true, `<string>&lt;&gt;</string>`, &s, UnmarshalExpectsEq{"<>"}},
		{true, `<integer>42</integer>`, &i, UnmarshalExpectsEq{int(42)}},
		{true, `<integer>42</integer>`, &i64, UnmarshalExpectsEq{int64(42)}},
		{true, `<integer>256</integer>`, &i8, UnmarshalExpectsError{&strconv.NumError{}}},
		{true, `<integer>256</integer>`, &s, UnmarshalExpectsError{&UnexpectedTokenError{}}},
		{true, `<integer>10</integer>`, &u16, UnmarshalExpectsEq{uint16(10)}},
		{true, `<integer>10</integer>`, &up, UnmarshalExpectsEq{uintptr(10)}},
		{true, `<integer>10</integer>`, new(chan int), UnmarshalExpectsError{&CannotParseTypeError{}}},
		{true, `<real>3.14</real>`, &f32, UnmarshalExpectsEq{float32(3.14)}},
		{true, `<false/>`, &b, UnmarshalExpectsEq{false}},
		{true, `<true/>`, &b, UnmarshalExpectsEq{true}},
		{true, `<date>2016-05-04T03:02:01Z</date>`, &tm, UnmarshalExpectsEq{time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)}},
		{true, `<data>aGVsbG8=</data>`, &buf, UnmarshalExpectsEq{*bytes.NewBuffer([]byte("hello"))}},
		{true, `<array><integer>4</integer><integer>2</integer></array>`, &ai, UnmarshalExpectsEq{[]int{4, 2}}},
		{true, ` <!-- use spaces and comments inside--> <array><!-- --><real>4</real> <real>2</real><!-- --> </array> <!-- -->`, &af32, UnmarshalExpectsEq{[]float32{4, 2}}},
		{true, `<any><key>B</key><true/><key>I</key><integer>42</integer></any>`, &s1, UnmarshalExpectsEq{S1{42, true}}},
	}

	for _, c := range test_cases {
		if c.run {
			// set c.pvar to zero before test starts
			v := reflect.Indirect(reflect.ValueOf(c.pvar))
			v.Set(reflect.Zero(reflect.TypeOf(v.Interface())))

			c.test.TestUnmarshal(t, c.xml, c.pvar)
		}
	}
}

type TestUnmarshaler interface {
	TestUnmarshal(t *testing.T, xml_text string, v interface{})
}

type UnmarshalExpectsEq struct {
	expected_val interface{}
}

func (res UnmarshalExpectsEq) TestUnmarshal(t *testing.T, xml_text string, v interface{}) {
	err := Unmarshal([]byte(xml_text), v)
	if err != nil {
		t.Errorf("Unmarshaling of %q into %T unexpectedly failed: %#s", xml_text, v, err)
		return
	}

	val := reflect.Indirect(reflect.ValueOf(v)).Interface()
	if !reflect.DeepEqual(val, res.expected_val) {
		t.Errorf("Unmarshaling of %q into %T should return %#v not %#v", xml_text, v, res.expected_val, val)
		return
	}
}

type UnmarshalExpectsError struct {
	errType error
}

func (exp UnmarshalExpectsError) TestUnmarshal(t *testing.T, xml_text string, v interface{}) {
	err := Unmarshal([]byte(xml_text), v)
	if reflect.TypeOf(err) != reflect.TypeOf(exp.errType) {
		t.Errorf("Unmarshaling of %q into %T expected error %#v, but got %#v", xml_text, v, exp.errType, err)
	}
}
