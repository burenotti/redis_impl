package resp

import (
	"bufio"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected interface{}
		Error    error
	}{
		{Name: "simple string", Input: "+abacaba\r\n", Expected: "abacaba"},
		{Name: "empty_string", Input: "+\r\n", Expected: ""},
		{Name: "positive", Input: ":67890\r\n", Expected: int64(67890)},
		{Name: "negative", Input: ":-123456\r\n", Expected: int64(-123456)},
		{Name: "unary_plus", Input: ":+123\r\n", Expected: int64(123)},
		{Name: "error", Input: "-abacaba\r\n", Expected: errors.New("abacaba")},
		{Name: "empty_bulk_string", Input: "$0\r\n\r\n", Expected: []byte{}},
		{Name: "nil_bulk_string", Input: "$-1\r\n", Expected: []byte(nil)},
		{Name: "bulk_string", Input: "$7\r\na\rb\nc\r\n\r\n", Expected: []byte("a\rb\nc\r\n")},
		{Name: "nil_array", Input: "*-1\r\n", Expected: []interface{}(nil)},
		{Name: "empty_array", Input: "*0\r\n", Expected: []interface{}{}},
		{Name: "empty buffer", Input: "", Error: io.EOF},
		{Name: "lkasjdaksdjasgdjh", Input: "", Error: io.EOF},
		{
			Name:  "Array",
			Input: "*4\r\n+abacaba\r\n:32\r\n:42\r\n$3\r\nabc\r\n",
			Expected: []interface{}{"abacaba",
				int64(32),
				int64(42),
				[]byte("abc"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(c.Input))
			actual, err := Unmarshal(r)
			if err != nil {
				require.ErrorIs(t, err, c.Error)
			}
			assert.Equal(t, c.Expected, actual)
		})
	}
}

func TestMarshal(t *testing.T) {
	cases := []struct {
		Name     string
		Input    interface{}
		Expected string
		Error    error
	}{
		{Name: "simple string", Expected: "+abacaba\r\n", Input: "abacaba"},
		{Name: "empty_string", Expected: "+\r\n", Input: ""},
		{Name: "positive", Expected: ":67890\r\n", Input: int64(67890)},
		{Name: "negative", Expected: ":-123456\r\n", Input: int64(-123456)},
		{Name: "error", Expected: "-abacaba\r\n", Input: errors.New("abacaba")},
		{Name: "empty_bulk_string", Expected: "$0\r\n\r\n", Input: []byte{}},
		{Name: "nil_bulk_string", Expected: "$-1\r\n", Input: []byte(nil)},
		{Name: "bulk_string", Expected: "$7\r\na\rb\nc\r\n\r\n", Input: []byte("a\rb\nc\r\n")},
		{Name: "nil_array", Expected: "*-1\r\n", Input: []interface{}(nil)},
		{Name: "empty_array", Expected: "*0\r\n", Input: []interface{}{}},
		{
			Name:     "Array",
			Expected: "*4\r\n+abacaba\r\n:32\r\n:42\r\n$3\r\nabc\r\n",
			Input: []interface{}{"abacaba",
				int64(32),
				int64(42),
				[]byte("abc"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			w := &strings.Builder{}
			err := Marshal(w, c.Input)
			if err != nil {
				require.ErrorIs(t, err, c.Error)
			}
			assert.Equal(t, c.Expected, w.String())
		})
	}
}
