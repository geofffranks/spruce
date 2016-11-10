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

func jsonifyData(data []byte) (string, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return "", err
	}

	doc, err := y.Map()
	if err != nil {
		return "", ansi.Errorf("@R{Root of YAML document is not a hash/map}: %s\n", err.Error())
	}

	b, err := json.Marshal(deinterface(doc))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func JSONifyIO(in io.Reader) (string, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return "", ansi.Errorf("@R{Error reading input}: %s", err)
	}
	return jsonifyData(data)
}

func JSONifyFiles(paths []string) ([]string, error) {
	l := make([]string, len(paths))

	for i, path := range paths {
		DEBUG("Processing file '%s'", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, ansi.Errorf("@R{Error reading file} @m{%s}: %s", path, err)
		}

		if l[i], err = jsonifyData(data); err != nil {
			return nil, ansi.Errorf("%s: %s", path, err)
		}
	}

	return l, nil
}

func deinterface(o interface{}) interface{} {
	switch o.(type) {
	case map[interface{}]interface{}:
		return deinterfaceMap(o.(map[interface{}]interface{}))
	case []interface{}:
		return deinterfaceList(o.([]interface{}))
	default:
		return o
	}
}

func deinterfaceMap(o map[interface{}]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range o {
		m[fmt.Sprintf("%v", k)] = deinterface(v)
	}
	return m
}

func deinterfaceList(o []interface{}) []interface{} {
	l := make([]interface{}, len(o))
	for i, v := range o {
		l[i] = deinterface(v)
	}
	return l
}
