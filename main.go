package main

import (
	"fmt"
	"os"

	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	"github.com/voxelbrain/goptions"
	"gopkg.in/yaml.v2"
	"strings"
)

var printfStdOut = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

var printfStdErr = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

var getopts = func(o interface{}) {
	err := goptions.Parse(o)
	if err != nil {
		usage()
	}
}

var exit = func(code int) {
	os.Exit(code)
}

var usage = func() {
	goptions.PrintHelp()
	exit(1)
}

var debug bool

func DEBUG(format string, args ...interface{}) {
	if debug {
		printfStdErr(format, args...)
	}
}

func main() {
	var options struct {
		Debug  bool `goptions:"-D, --debug, description='Enable debugging'"`
		Action goptions.Verbs
		Merge  struct {
			Files goptions.Remainder `goptions:"description='Merges file2.yml through fileN.yml on top of file1.yml'"`
		} `goptions:"merge"`
	}
	getopts(&options)

	if os.Getenv("DEBUG") != "" && strings.ToLower(os.Getenv("DEBUG")) != "false" && os.Getenv("DEBUG") != "0" {
		debug = true
	}
	if options.Debug {
		debug = options.Debug
	}

	DEBUG("Debugging enabled")

	switch {
	case options.Action == "merge":
		if len(options.Merge.Files) >= 1 {
			root := make(map[interface{}]interface{})

			err := mergeAllDocs(root, options.Merge.Files)
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
			usage()
		}
	default:
		usage()
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
