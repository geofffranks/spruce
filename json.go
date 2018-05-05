package spruce

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/starkandwayne/goutils/ansi"

	"github.com/geofffranks/simpleyaml"
	. "github.com/geofffranks/spruce/log"
)

func jsonifyData(data []byte, strict bool) (string, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return "", err
	}

	doc, err := y.Map()
	if err != nil {
		return "", ansi.Errorf("@R{Root of YAML document is not a hash/map}: %s\n", err.Error())
	}

	doc_, err := deinterface(doc, strict)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(doc_)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func JSONifyIO(in io.Reader, strict bool) (string, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return "", ansi.Errorf("@R{Error reading input}: %s", err)
	}
	return jsonifyData(data, strict)
}

func JSONifyFiles(paths []string, strict bool) ([]string, error) {
	l := make([]string, len(paths))

	for i, path := range paths {
		DEBUG("Processing file '%s'", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, ansi.Errorf("@R{Error reading file} @m{%s}: %s", path, err)
		}

		if l[i], err = jsonifyData(data, strict); err != nil {
			return nil, ansi.Errorf("%s: %s", path, err)
		}
	}

	return l, nil
}

func deinterface(o interface{}, strict bool) (interface{}, error) {
	switch o.(type) {
	case map[interface{}]interface{}:
		return deinterfaceMap(o.(map[interface{}]interface{}), strict)
	case []interface{}:
		return deinterfaceList(o.([]interface{}), strict)
	default:
		return o, nil
	}
}

func addKeyToMap(m map[string]interface{}, k interface{}, v interface{}, strict bool) (error) {
	vs := fmt.Sprintf("%v", k)
	_, exists := m[vs]
	if exists {
		NewWarningError(eContextAll, "@Y{Duplicate key detected: %s}", vs).Warn()
		return nil
	}
	dv, err := deinterface(v, strict)
	if err != nil {
		return err
	}
	m[vs] = dv
	return nil
}

func deinterfaceMap(o map[interface{}]interface{}, strict bool) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	for k, v := range o {

		switch k.(type) {
		case string:
			err := addKeyToMap(m, k, v, strict)
			if err != nil {
				return nil, err
			}
		default:
			if (strict) {
		               return nil, fmt.Errorf("Non-string keys found during strict JSON conversion")
			} else {
				addKeyToMap(m, k, v, strict)
			}
		}

	}
	return m, nil
}

func deinterfaceList(o []interface{}, strict bool) ([]interface{}, error) {
	l := make([]interface{}, len(o))
	for i, v := range o {
		v_, err := deinterface(v, strict)
		if err != nil {
			return nil, err
		}
		l[i] = v_
	}
	return l, nil
}
