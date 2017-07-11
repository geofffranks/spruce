package patch

import (
	"fmt"
)

type FindOp struct {
	Path Pointer
}

func (op FindOp) Apply(doc interface{}) (interface{}, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return doc, nil
	}

	obj := doc

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2

		switch typedToken := token.(type) {
		case IndexToken:
			idx := typedToken.Index

			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			if idx >= len(typedObj) {
				return nil, opMissingIndexErr{idx, typedObj}
			}

			if isLast {
				return typedObj[idx], nil
			} else {
				obj = typedObj[idx]
			}

		case AfterLastIndexToken:
			errMsg := "Expected not to find after last index token in path '%s' (not supported in find operations)"
			return nil, fmt.Errorf(errMsg, op.Path)

		case MatchingIndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			var idxs []int

			for itemIdx, item := range typedObj {
				typedItem, ok := item.(map[interface{}]interface{})
				if ok {
					if typedItem[typedToken.Key] == typedToken.Value {
						idxs = append(idxs, itemIdx)
					}
				}
			}

			if typedToken.Optional && len(idxs) == 0 {
				obj = map[interface{}]interface{}{typedToken.Key: typedToken.Value}

				if isLast {
					return obj, nil
				}
			} else {
				if len(idxs) != 1 {
					return nil, opMultipleMatchingIndexErr{NewPointer(tokens[:i+2]), idxs}
				}

				idx := idxs[0]

				if isLast {
					return typedObj[idx], nil
				} else {
					obj = typedObj[idx]
				}
			}

		case KeyToken:
			typedObj, ok := obj.(map[interface{}]interface{})
			if !ok {
				return nil, newOpMapMismatchTypeErr(tokens[:i+2], obj)
			}

			var found bool

			obj, found = typedObj[typedToken.Key]
			if !found && !typedToken.Optional {
				return nil, opMissingMapKeyErr{typedToken.Key, NewPointer(tokens[:i+2]), typedObj}
			}

			if isLast {
				return typedObj[typedToken.Key], nil
			} else {
				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case MatchingIndexToken:
						obj = []interface{}{}
					case KeyToken:
						obj = map[interface{}]interface{}{}
					default:
						errMsg := "Expected to find key or matching index token at path '%s'"
						return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+3]))
					}
				}
			}

		default:
			return nil, opUnexpectedTokenErr{token, NewPointer(tokens[:i+2])}
		}
	}

	return doc, nil
}
