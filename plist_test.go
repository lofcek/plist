package plist

import (
	"reflect"
	"testing"
	"strconv"
)

type TestCase interface {
	TestUnmarshal(t *testing.T, xml_text string, v interface{})
}

type expect_ok struct {
	expected_val interface{}
}


func (res expect_ok) TestUnmarshal(t *testing.T, xml_text string, v interface{}) {
	err := Unmarshal([]byte(xml_text), v)
	if err != nil {
		t.Errorf("Unmarshaling of %q into %T unexpectedly failed: %#s",xml_text, v, err)
	}

	val := reflect.Indirect(reflect.ValueOf(v)).Interface()
	if !reflect.DeepEqual(val, res.expected_val) {
		t.Errorf("Unmarshaling of %q into %T should return %#v not %#v", xml_text, v, res.expected_val, val)
	}
}

type expect_error struct {
	err error
}

func (exp expect_error) TestUnmarshal(t *testing.T, xml_text string, v interface{}) {
	err := Unmarshal([]byte(xml_text), v)
	if reflect.TypeOf(err) != reflect.TypeOf(exp.err) {
		t.Errorf("Unmarshaling of %q into %T expected error %#v, but got %#v", xml_text, v, exp.err, err)
	}
}

func TestOk(t *testing.T) {
	var str string
	var integer int
	var integer8 int8
	var integer64 int64
	type Test struct {
		xml          string
		variable     interface{}
		test_case TestCase
	}

	test_cases := []Test{
		{`<string>a</string>`, &str, expect_ok{"a"}},
		{`<string>b</string>`, &str, expect_ok{"b"}},
		{`<integer>42</integer>`, &integer, expect_ok{int(42)}},
		{`<integer>42</integer>`, &integer64, expect_ok{int64(42)}},
		{`<integer>256</integer>`, &integer8, expect_error{&strconv.NumError{}}},
	}

	for _, c := range test_cases {
		c.test_case.TestUnmarshal(t, c.xml, c.variable)
	}
}
