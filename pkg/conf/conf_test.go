package conf

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestConfig_Get(t *testing.T) {
	t.Parallel()
	config := &Config{}
	config.data["key"] = []string{"val"}
	config.data["key_arr"] = []string{"val1", "val2", "val3"}
	config.data["int"] = []string{"a", "123"}

	cases := []struct {
		Name     string
		Expected interface{}
		Error    error
		Get      func(c *Config) (interface{}, error)
	}{
		{
			Name:     "string",
			Expected: "val",
			Error:    nil,
			Get: func(c *Config) (interface{}, error) {
				return c.Get("key").String()
			},
		},
		{
			Name:     "string at 0",
			Expected: "val1",
			Error:    nil,
			Get: func(c *Config) (interface{}, error) {
				return c.Get("key_arr").At(0).String()
			},
		},
		{
			Name:     "string at 1",
			Expected: "val2",
			Error:    nil,
			Get: func(c *Config) (interface{}, error) {
				return c.Get("key_arr").At(1).String()
			},
		},
		{
			Name:     "int",
			Expected: 123,
			Error:    nil,
			Get: func(c *Config) (interface{}, error) {
				return c.Get("int").At(1).Int()
			},
		},
	}

	for _, c := range cases {
		actual, err := c.Get(config)
		if c.Error != nil {
			assert.ErrorIs(t, err, c.Error)
		} else {
			require.NoError(t, err)
			assert.EqualValues(t, c.Expected, actual)
		}
	}
}

func TestConfig_parse(t *testing.T) {
	cases := []struct {
		Name     string
		Input    []string
		Expected map[string][]string
		Error    error
	}{
		{
			Name: "simple case",
			Input: []string{
				"bind 0.0.0.0 127.0.0.1",
				"listen 80",
			},
			Expected: map[string][]string{
				"bind":   {"0.0.0.0", "127.0.0.1"},
				"listen": {"80"},
			},
			Error: nil,
		},
		{
			Name: "quotes",
			Input: []string{
				`name "artem burenin"`,
				`name1 'artem burenin' "val 1 val 2"`,
			},
			Expected: map[string][]string{
				"name":  {"artem burenin"},
				"name1": {"artem burenin", "val 1 val 2"},
			},
			Error: nil,
		},
		{
			Name: "quotes with escapes",
			Input: []string{
				`first_name 'ar\'tem'`,
				`last_name "\"burenin\""`,
			},
			Expected: map[string][]string{
				"first_name": {"ar'tem"},
				"last_name":  {`"burenin"`},
			},
			Error: nil,
		},
		{
			Name: "quotes with escapes",
			Input: []string{
				`first_name 'ar\'tem'`,
				`last_name "\"burenin\""`,
			},
			Expected: map[string][]string{
				"first_name": {"ar'tem"},
				"last_name":  {`"burenin"`},
			},
			Error: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			input := strings.Join(c.Input, "\r\n")
			r := strings.NewReader(input)
			actual := make(map[string][]string)
			err := parse(actual, r)
			if c.Error != nil {
				assert.ErrorIs(t, err, c.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.Expected, actual)
			}
		})
	}
}
