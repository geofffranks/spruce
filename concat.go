package main

import (
	"fmt"
	"reflect"
	"regexp"
)

type TokenType int

const (
	TokenLiteral = iota
	TokenReference
)

type Token struct {
	Type  TokenType
	Value string
}

// Concatenator is an implementation of PostProcessor to de-reference (( grab me.data )) calls
type Concatenator struct {
	root map[interface{}]interface{}
}

// Action returns the Action string for the Concatenator
func (s Concatenator) Action() string {
	return "concatenator"
}

// PostProcess - resolves a value by seeing if it matches (( grab me.data )) and retrieves me.data's value
func (s Concatenator) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if o != nil && reflect.TypeOf(o).Kind() == reflect.String {
		re := regexp.MustCompile(`^\Q((\E\s*concat\s+(.+)\s*\Q))\E$`)
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))

			tokens := parseWords(keys[1])
			str := ""
			for _, token := range tokens {
				if token.Type == TokenLiteral {
					str += token.Value
				} else {
					DEBUG("%s: resolving from %s", node, token.Value)
					val, err := resolveNode(token.Value, s.root)
					if err != nil {
						return nil, "error", fmt.Errorf("%s: Unable to resolve `%s`: `%s", node, token.Value, err.Error())
					}
					// error if val is not a string
					str += val.(string)
				}
			}

			return str, "replace", nil
		}
	}

	return nil, "ignore", nil
}

func splitQuoted(src string) []string {
	list := make([]string, 0, 0)

	buf := ""
	escaped := false
	quoted := false

	for _, c := range src {
		if escaped {
			buf += string(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == ' ' || c == '\t' {
			if quoted {
				buf += string(c)
				continue
			} else if buf != "" {
				list = append(list, buf)
				buf = ""
			}
			continue
		}

		if c == '"' {
			buf += string(c)
			quoted = !quoted
			continue
		}

		buf += string(c)
	}

	if buf != "" {
		list = append(list, buf)
	}

	return list
}

func parseWords(src string) []Token {
	raw := splitQuoted(src)
	re := regexp.MustCompile(`^"(.*)"$`)
	tokens := make([]Token, len(raw), len(raw))
	for i, s := range raw {
		if re.MatchString(s) {
			keys := re.FindStringSubmatch(s)
			tokens[i] = Token{Type: TokenLiteral, Value: keys[1]}
		} else {
			tokens[i] = Token{Type: TokenReference, Value: s}
		}
	}

	return tokens
}
