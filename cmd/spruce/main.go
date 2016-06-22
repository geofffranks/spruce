package main

import (
	"fmt"
	"github.com/jhunt/ansi"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"

	. "github.com/geofffranks/spruce"
	. "github.com/geofffranks/spruce/log"

	"regexp"
	"strings"

	"github.com/smallfish/simpleyaml"
	"github.com/voxelbrain/goptions"
	"gopkg.in/yaml.v2"
)

// Version holds the Current version of spruce
var Version = "(development)"

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

var handleConcourseQuoting bool

func envFlag(varname string) bool {
	val := os.Getenv(varname)
	return val != "" && strings.ToLower(val) != "false" && val != "0"
}

func main() {
	var options struct {
		Debug     bool `goptions:"-D, --debug, description='Enable debugging'"`
		Trace     bool `goptions:"-T, --trace, description='Enable trace mode debugging (very verbose)'"`
		Version   bool `goptions:"-v, --version, description='Display version information'"`
		Concourse bool `goptions:"--concourse, description='Pre/Post-process YAML for Concourse CI (handles {{ }} quoting)'"`
		Action    goptions.Verbs
		Merge     struct {
			Prune []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once'"`
			Files goptions.Remainder `goptions:"description='Merges file2.yml through fileN.yml on top of file1.yml'"`
		} `goptions:"merge"`
		JSON struct {
			Files goptions.Remainder `goptions:"description='Files to convert to JSON'"`
		} `goptions:"json"`
	}
	getopts(&options)

	if envFlag("DEBUG") || options.Debug {
		DebugOn = true
	}

	if envFlag("TRACE") || options.Trace {
		TraceOn = true
		DebugOn = true
	}

	handleConcourseQuoting = options.Concourse

	if options.Version {
		printfStdOut("%s - Version %s\n", os.Args[0], Version)
		exit(0)
		return
	}

	ansi.Color(isatty.IsTerminal(os.Stderr.Fd()))

	switch options.Action {
	case "merge":
		if len(options.Merge.Files) >= 1 {
			root := make(map[interface{}]interface{})

			err := mergeAllDocs(root, options.Merge.Files)
			if err != nil {
				printfStdErr("%s\n", err.Error())
				exit(2)
				return
			}

			ev := &Evaluator{Tree: root}
			err = ev.Run(options.Merge.Prune)
			if err != nil {
				printfStdErr("%s\n", err.Error())
				exit(2)
				return
			}

			TRACE("Converting the following data back to YML:")
			TRACE("%#v", ev.Tree)
			merged, err := yaml.Marshal(ev.Tree)
			if err != nil {
				printfStdErr("Unable to convert merged result back to YAML: %s\nData:\n%#v", err.Error(), ev.Tree)
				exit(2)
				return

			}

			var output string
			if handleConcourseQuoting {
				output = dequoteConcourse(merged)
			} else {
				output = string(merged)
			}
			printfStdOut("%s\n", output)

		} else {
			usage()
			return
		}

	case "json":
		if len(options.JSON.Files) > 0 {
			jsons, err := JSONifyFiles(options.JSON.Files)
			if err != nil {
				printfStdErr("%s\n", err)
				exit(2)
				return
			}
			for _, output := range jsons {
				printfStdOut("%s\n", output)
			}
		} else {
			output, err := JSONifyIO(os.Stdin)
			if err != nil {
				printfStdErr("%s\n", err)
				exit(2)
				return
			}
			printfStdOut("%s\n", output)
		}

	default:
		usage()
		return
	}
}

func parseYAML(data []byte) (map[interface{}]interface{}, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return nil, err
	}

	doc, err := y.Map()
	if err != nil {
		return nil, ansi.Errorf("@R{Root of YAML document is not a hash/map}: %s\n", err.Error())
	}

	return doc, nil
}

func mergeAllDocs(root map[interface{}]interface{}, paths []string) error {
	m := &Merger{}

	for _, path := range paths {
		DEBUG("Processing file '%s'", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", path, err.Error())
		}

		if handleConcourseQuoting {
			data = quoteConcourse(data)
		}

		doc, err := parseYAML(data)
		if err != nil {
			return ansi.Errorf("@m{%s}: @R{%s}\n", path, err.Error())
		}

		m.Merge(root, doc)

		tmpYaml, _ := yaml.Marshal(root) // we don't care about errors for debugging
		TRACE("Current data after processing '%s':\n%s", path, tmpYaml)
	}

	return m.Error()
}

var concourseRegex = `\{\{([-\w\p{L}]+)\}\}`

func quoteConcourse(input []byte) []byte {
	re := regexp.MustCompile("(" + concourseRegex + ")")
	return re.ReplaceAll(input, []byte("\"$1\""))
}

func dequoteConcourse(input []byte) string {
	re := regexp.MustCompile("['\"](" + concourseRegex + ")[\"']")
	return re.ReplaceAllString(string(input), "$1")
}
