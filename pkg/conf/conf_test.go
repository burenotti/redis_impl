package conf

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
)

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

type Size int

func (s *Size) SetValue(raw []string) error {
	i, err := strconv.Atoi(raw[0])
	if err != nil {
		return err
	}
	*s = Size(i)
	return nil
}

func TestBind(t *testing.T) {
	data := `#
bind 0.0.0.0
use-tls true
use-ssl false
port 6379
timeout_idle 5
timeout_read 6
timeout_write 7
a_b_c_d 100
`
	r := strings.NewReader(data)

	cfg := struct {
		Host           string `redis:"bind" redis-default:"127.0.0.1"`
		Port           int    `redis:"port" redis-default:"6379"`
		MaxConnections int    `redis:"max_connections" redis-default:"10"`
		UseTLS         bool   `redis:"use-tls" redis-default:"false"`
		UseSSL         bool   `redis:"use-ssl" redis-default:"false"`
		MaxMemory      Size   `redis:"max_memory" redis-default:"300"`
		A              struct {
			B struct {
				C struct {
					D int `redis:"d" redis-default:"1"`
				} `redis-prefix:"c_"`
			} `redis-prefix:"b_"`
		} `redis-prefix:"a_"`
		Timeout struct {
			Idle  int `redis:"idle" redis-default:"1"`
			Read  int `redis:"read" redis-default:"2"`
			Write int `redis:"write" redis-default:"3"`
		} `redis-prefix:"timeout_"`
	}{}
	err := Bind(&cfg, r)
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 6379, cfg.Port)
	assert.Equal(t, 10, cfg.MaxConnections)
	assert.Equal(t, 5, cfg.Timeout.Idle)
	assert.Equal(t, 6, cfg.Timeout.Read)
	assert.Equal(t, 7, cfg.Timeout.Write)
	assert.Equal(t, Size(300), cfg.MaxMemory)
	assert.Equal(t, true, cfg.UseTLS)
	assert.Equal(t, false, cfg.UseSSL)
	assert.Equal(t, 100, cfg.A.B.C.D)

}

func TestBind_requiredFields(t *testing.T) {

	r := strings.NewReader(``)

	cfg1 := struct {
		B string `redis:"b" redis-default:"default-b"`
		C string `redis:"c" redis-default:"default-c"`
		A string `redis:"a" redis-required:""`
	}{}
	err := Bind(&cfg1, r)
	assert.ErrorIs(t, err, ErrRequired)
}

func TestBind_badCases(t *testing.T) {
	cfg := struct {
		A complex64 `redis:"a"`
	}{}

	err := Bind(&cfg, strings.NewReader("a 123i+2"))
	require.ErrorIs(t, err, ErrTypeNotSupported)
}
