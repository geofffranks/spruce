package main

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/geofffranks/spruce/resolve"
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
	ttl  int
}

// Action returns the Action string for the Concatenator
func (s Concatenator) Action() string {
	return "concatenator"
}

// parseConcatOp - determine if an object is a (( concat ... )) call
func parseConcatOp(o interface{}) (bool, string) {
	if o != nil && reflect.TypeOf(o).Kind() == reflect.String {
		re := regexp.MustCompile(`^\Q((\E\s*concat\s+(.+)\s*\Q))\E$`)
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			return true, keys[1]
		}
	}
	return false, ""
}

// resolve - resolves a set of tokens (literals or references), co-recursively with resolveKey()
func (s Concatenator) resolve(node string, tokens []Token) (string, error) {
	str := ""

	for _, token := range tokens {
		if token.Type == TokenLiteral {
			str += token.Value
		} else {
			DEBUG("%s: resolving from %s", node, token.Value)
			val, err := s.resolveKey(token.Value)
			if err != nil {
				return "", err
			}
			str += val
		}
	}
	return str, nil
}

// resolveKey - resolves a single key reference, co-recursively with resolve()
func (s Concatenator) resolveKey(key string) (string, error) {
	val, err := resolve.ResolveNode(key, s.root)
	if err != nil {
		return "", fmt.Errorf("Unable to resolve `%s`: `%s", key, err)
	}

	if should, args := parseConcatOp(val); should {
		if s.ttl -= 1; s.ttl <= 0 {
			return "", fmt.Errorf("possible recursion detected in call to (( concat ))")
		}
		str, err := s.resolve(key, parseWords(args))
		s.ttl += 1
		if err != nil {
			return "", err
		}
		return str, nil
	}
	return fmt.Sprintf("%v", val), nil
}

// PostProcess - resolves a value by seeing if it matches (( concat me.data )) and retrieves me.data's value
func (s Concatenator) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if should, args := parseConcatOp(o); should {
		s.ttl = 64
		str, err := s.resolve(node, parseWords(args))
		if err != nil {
			return nil, "error", fmt.Errorf("%s: %s", node, err.Error())
		}
		return str, "replace", nil
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
