package main

import (
	"fmt"
	"os"

	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	"gopkg.in/yaml.v2"
)

var printfStdOut = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

var printfStdErr = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

var exit = func(code int) {
	os.Exit(code)
}

func main() {
	if len(os.Args) > 1 {
		root := make(map[interface{}]interface{})

		err := mergeAllDocs(root, os.Args[1:])
		if err != nil {
			printfStdErr(err.Error())
			exit(2)
		}

		merged, err := yaml.Marshal(root)
		if err != nil {
			printfStdErr("Unable to convert merged result back to YAML: %s\nData:\n%#v", err.Error(), root)
			exit(2)
		}
		printfStdOut("%s\n", string(merged))
	} else {
		printfStdErr("Usage: %s <first.yml> ... <nth.yml>\n", os.Args[0])
		exit(1)
	}
}

func parseYAML(data []byte) (map[interface{}]interface{}, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return nil, err
	}

	doc, err := y.Map()
	if err != nil {
		return nil, fmt.Errorf("Root of YAML document is not a hash/map: %s\n", err.Error())
	}

	return doc, nil
}

func mergeAllDocs(root map[interface{}]interface{}, paths []string) error {
	for _, path := range paths {
		data, err := readFile(path)
		if err != nil {
			return fmt.Errorf("Error reading file %s: %s\n", path, err.Error())
		}

		doc, err := parseYAML(data)
		if err != nil {
			return fmt.Errorf("%s: %s\n", path, err.Error())
		}

		mergeMap(root, doc, "")
	}
	return nil
}
