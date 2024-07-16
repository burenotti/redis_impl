package conf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrNotSupported = errors.New("feature not supported")
	ErrSyntax       = errors.New("syntax error")
)

func loadModule(_ map[string][]string, _ []string) error {
	return fmt.Errorf("%w: loading modules is not yet supported", ErrNotSupported)
}

func include(data map[string][]string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: include requires exactly one argument", ErrSyntax)
	}
	file, err := os.Open(args[0])
	if err != nil {
		return err
	}

	return parse(data, file)
}

func parse(data map[string][]string, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if scanner.Err() != nil {
			return scanner.Err()
		}

		if err := parseLine(data, scanner.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func parseLine(data map[string][]string, line []byte) error {
	// Comment line
	if r, _ := utf8.DecodeRune(line); r == '#' {
		return nil
	}

	tokens, err := tokenizeLine(bytes.Runes(line))
	if err != nil {
		return err
	}

	// Empty strings
	if len(tokens) == 0 {
		return nil
	}

	keywords := map[string]func(data map[string][]string, args []string) error{
		"include":    include,
		"loadmodule": loadModule,
	}

	// Has a keyword (loadmodule, include, etc...)
	if handler, ok := keywords[tokens[0]]; ok {
		return handler(data, tokens[1:])
	}

	data[tokens[0]] = tokens[1:]
	return nil
}

func tokenizeLine(line []rune) (result []string, err error) {
	word := strings.Builder{}
	quoteBegin := -1
	var quoteType rune

	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch ch {
		case '\'', '"':
			if quoteBegin == -1 {
				quoteType = ch
				quoteBegin = i + 1
			} else {
				if quoteType == ch {
					quoteType = 0
					quoteBegin = -1
					result = append(result, word.String())
					word.Reset()
				} else {
					word.WriteRune(ch)
				}
			}
		case '\\':
			if i >= len(line)-1 {
				return result, fmt.Errorf("%w: backslash at the end of line", ErrSyntax)
			}
			if line[i+1] == '\'' || line[i+1] == '"' {
				word.WriteRune(line[i+1])
				i++
			} else {
				return result, fmt.Errorf("%w: unknown escape sequence", ErrSyntax)
			}

		default:
			if unicode.IsSpace(ch) && quoteBegin == -1 && word.Len() > 0 {
				result = append(result, word.String())
				word.Reset()
			} else if word.Len() != 0 || !unicode.IsSpace(ch) {
				word.WriteRune(ch)
			}
		}
	}
	if quoteBegin != -1 {
		err = fmt.Errorf("%w: unclosed quote started at %d", ErrSyntax, quoteBegin)
	}
	if word.Len() > 0 {
		result = append(result, word.String())
	}
	return result, err
}
