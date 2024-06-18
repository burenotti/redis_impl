package resp

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestValue_Unmarshal(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected interface{}
		Value    Value
	}{
		{Name: "string", Input: "+abacaba\r\n", Expected: "abacaba", Value: SimpleString("")},
		{Name: "empty_string", Input: "+\r\n", Expected: "", Value: SimpleString("")},
		{Name: "positive", Input: ":67890\r\n", Expected: int64(67890), Value: Int(0)},
		{Name: "negative", Input: ":-123456\r\n", Expected: int64(-123456), Value: Int(0)},
		{Name: "unary_plus", Input: ":+123\r\n", Expected: int64(123), Value: Int(0)},
		{Name: "error", Input: "-abacaba\r\n", Expected: "abacaba", Value: Error("")},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(c.Input))
			err := c.Value.Unmarshal(r)
			require.NoError(t, err, "valid input should not produce an error")
			assert.Equal(t, c.Expected, c.Value.Value())
		})
	}
}

func TestBulkString_Unmarshal(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		r := bufio.NewReader(strings.NewReader("$0\r\n\r\n"))
		str := NullBulkString()
		err := str.Unmarshal(r)
		require.NoError(t, err)
		bytes, ok := str.Bytes()
		assert.True(t, ok)
		assert.Equal(t, 0, len(bytes))
		assert.False(t, str.IsNull())
	})

	t.Run("null string", func(t *testing.T) {
		r := bufio.NewReader(strings.NewReader("$-1\r\n"))
		str := NullBulkString()
		err := str.Unmarshal(r)
		require.NoError(t, err)
		bytes, ok := str.Bytes()
		assert.True(t, ok)
		assert.Equal(t, 0, len(bytes))
		assert.True(t, str.IsNull())
	})

	t.Run("string with line endings", func(t *testing.T) {
		r := bufio.NewReader(strings.NewReader("$7\r\na\rb\nc\r\n\r\n"))
		str := NullBulkString()
		err := str.Unmarshal(r)
		require.NoError(t, err)
		bytes, ok := str.Bytes()
		assert.True(t, ok)
		assert.Equal(t, []byte("a\rb\nc\r\n"), bytes)
		assert.False(t, str.IsNull())
	})

}

func TestArray_Unmarshal(t *testing.T) {
	cases := []struct {
		Name          string
		Input         string
		ExpectedSize  int
		ExpectedNull  bool
		ExpectedValue []Value
	}{
		{
			Name:          "null array",
			Input:         "*-1\r\n",
			ExpectedSize:  0,
			ExpectedNull:  true,
			ExpectedValue: []Value(nil),
		},
		{
			Name:          "empty array",
			Input:         "*0\r\n",
			ExpectedSize:  0,
			ExpectedNull:  false,
			ExpectedValue: []Value{},
		},
		{
			Name:         "simple values",
			Input:        "*4\r\n+abacaba\r\n:32\r\n:42\r\n$3\r\nabc\r\n",
			ExpectedSize: 4,
			ExpectedNull: false,
			ExpectedValue: []Value{
				SimpleString("abacaba"),
				Int(32),
				Int(42),
				BulkString([]byte("abc")),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(c.Input))
			arr := NullArray()
			err := arr.Unmarshal(r)
			require.NoError(t, err)
			value, ok := arr.Array()
			assert.True(t, ok)
			assert.Equal(t, c.ExpectedNull, arr.IsNull())
			assert.Equal(t, c.ExpectedSize, len(value))
			assert.Equal(t, value, c.ExpectedValue)
		})
	}
}
