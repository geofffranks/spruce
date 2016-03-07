package tree

import "fmt"
import "reflect"

// Attemts to find the value at `path` inside data structure `o`.
// If found, attempts to cast it as a string. Errors will be
// returned for data of invalid type, or nonexistent paths.
func FindString(o interface{}, path string) (string, error) {
	obj, err := Find(o, path)
	if err != nil {
		return "", err
	}
	if s, ok := obj.(string); ok {
		return s, nil
	} else {
		return "", fmt.Errorf("Invalid data type - wanted string, got %s", reflect.TypeOf(obj))
	}
}

// Attemts to find the value at `path` inside data structure `o`.
// If found, attempts to cast it as a Number. Errors will be
// returned for data of invalid type, or nonexistent paths.
func FindNum(o interface{}, path string) (Number, error) {
	var num Number
	obj, err := Find(o, path)
	if err != nil {
		return num, err
	}
	switch obj.(type) {
	case float64:
		num = Number(obj.(float64))
	case int:
		num = Number(float64(obj.(int)))
	default:
		return num, fmt.Errorf("Invalid data type - wanted number, got %s", reflect.TypeOf(obj))
	}
	return num, nil
}

// Attemts to find the value at `path` inside data structure `o`.
// If found, attempts to cast it as a bool. Errors will be
// returned for data of invalid type, or nonexistent paths.
func FindBool(o interface{}, path string) (bool, error) {
	obj, err := Find(o, path)
	if err != nil {
		return false, err
	}
	if b, ok := obj.(bool); ok {
		return b, nil
	} else {
		return false, fmt.Errorf("Invalid data type - wanted bool, got %s", reflect.TypeOf(obj))
	}
}

// Attemts to find the value at `path` inside data structure `o`.
// If found, attempts to cast it as a map[string]interface{}. Errors will be
// returned for data of invalid type, or nonexistent paths.
func FindMap(o interface{}, path string) (map[string]interface{}, error) {
	obj, err := Find(o, path)
	if err != nil {
		return map[string]interface{}{}, err
	}
	if m, ok := obj.(map[string]interface{}); ok {
		return m, nil
	} else {
		return map[string]interface{}{}, fmt.Errorf("Invalid data type - wanted map, got %s", reflect.TypeOf(obj))
	}
}

// Attemts to find the value at `path` inside data structure `o`.
// If found, attempts to cast it as an interface{} slice. Errors will be
// returned for data of invalid type, or nonexistent paths.
func FindArray(o interface{}, path string) ([]interface{}, error) {
	obj, err := Find(o, path)
	if err != nil {
		return []interface{}{}, err
	}
	if arr, ok := obj.([]interface{}); ok {
		return arr, nil
	} else {
		return []interface{}{}, fmt.Errorf("Invalid data type - wanted array, got %s", reflect.TypeOf(obj))
	}
}

// Attemts to find the value at `path` inside data structure `o`.
// If found, returns it as a plain interface{} type, for you to
// typecheck + cast as you see fit. Errors will be
// returned for data of invalid type, or nonexistent paths.
func Find(o interface{}, path string) (interface{}, error) {
	c, err := ParseCursor(path)
	if err != nil {
		return nil, err
	}
	return c.Resolve(o)
}

type Number float64

// Returns an `int64` representation of the Number. If the value
// is not an integer, returns an error, so you do not accidentally
// lose precision while trying to see if it is an integer value.
func (n Number) Int64() (int64, error) {
	i := int64(n)
	if Number(i) != n {
		return 0, fmt.Errorf("%f does not represent an integer, cannot auto-convert", float64(n))
	}
	return i, nil
}

// Returns a `float64` representation of the Number.
func (n Number) Float64() float64 {
	return float64(n)
}

// Returns a string representation of the Number
func (n Number) String() string {
	intVal, err := n.Int64()
	if err == nil {
		return fmt.Sprintf("%d", intVal)
	}

	return fmt.Sprintf("%f", n.Float64())
}
