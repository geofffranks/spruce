package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce"
	. "github.com/geofffranks/spruce/log"

	"regexp"
	"strings"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/yaml"
	"github.com/voxelbrain/goptions"
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
			SkipEval   bool               `goptions:"--skip-eval, description='Do not evaluate spruce logic after merging docs'"`
			Prune      []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once)'"`
			CherryPick []string           `goptions:"--cherry-pick, description='The opposite of prune, specify keys to cherry-pick from final output (may be specified more than once)'"`
			Files      goptions.Remainder `goptions:"description='Merges file2.yml through fileN.yml on top of file1.yml. To read STDIN, specify a filename of \\'-\\'.'"`
		} `goptions:"merge"`
		JSON struct {
			Files goptions.Remainder `goptions:"description='Files to convert to JSON'"`
		} `goptions:"json"`
		Diff struct {
			Files goptions.Remainder `goption:"description='Show the semantic differences between two YAML files'"`
		} `goptions:"diff"`
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
		if len(options.Merge.Files) < 1 {
			options.Merge.Files = append(options.Merge.Files, "-")
		}

		root := make(map[interface{}]interface{})

		err := mergeAllDocs(root, options.Merge.Files)
		if err != nil {
			printfStdErr("%s\n", err.Error())
			exit(2)
			return
		}

		ev := &Evaluator{Tree: root, SkipRunPhase: options.Merge.SkipEval}
		err = ev.Run(options.Merge.Prune, options.Merge.CherryPick)
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

	case "diff":
		if len(options.Diff.Files) != 2 {
			usage()
			return
		}
		output, err := diffFiles(options.Diff.Files)
		if err != nil {
			printfStdErr("%s\n", err)
			exit(2)
			return
		}
		printfStdOut("%s\n", output)

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
		var data []byte
		var err error

		if path == "-" {
			stat, err := os.Stdin.Stat()
			if err != nil {
				return ansi.Errorf("@R{Error statting STDIN} - Bailing out: %s\n", err.Error())
			}
			if stat.Mode()&os.ModeCharDevice == 0 {
				data, err = ioutil.ReadAll(os.Stdin)
				if err != nil {
					return ansi.Errorf("@R{Error reading STDIN}: %s\n", err.Error())
				}
			}
			if len(data) == 0 {
				return ansi.Errorf("@R{Error reading STDIN}: no data found. Did you forget to pipe data to STDIN, or specify yaml files to merge?")
			}
			path = "STDIN"
		} else {
			data, err = ioutil.ReadFile(path)
			if err != nil {
				return ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", path, err.Error())
			}
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

func diffFiles (paths []string) (string, error) {
	if len(paths) != 2 {
		return "", ansi.Errorf("incorrect number of files given to diffFiles(); please file a bug report")
	}

	data, err := ioutil.ReadFile(paths[0])
	if err != nil {
		return "", ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", paths[0], err)
	}
	a, err := parseYAML(data)
	if err != nil {
		return "", ansi.Errorf("@m{%s}: @R{%s}\n", paths[0], err)
	}

	data, err = ioutil.ReadFile(paths[1])
	if err != nil {
		return "", ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", paths[1], err)
	}
	b, err := parseYAML(data)
	if err != nil {
		return "", ansi.Errorf("@m{%s}: @R{%s}\n", paths[1], err)
	}

	d, err := Diff(a, b)
	if err != nil {
		return "", ansi.Errorf("@R{Failed to diff} @m{%s} -> @m{%s}: %s\n", paths[0], paths[1], err)
	}

	if !d.Changed() {
		return ansi.Sprintf("@G{both files are semantically equivalent; no differences found!}\n"), nil
	}
	return d.String("$"), nil
}
