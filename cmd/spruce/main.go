package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/cppforlife/go-patch/patch"
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

type mergeOpts struct {
	SkipEval       bool               `goptions:"--skip-eval, description='Do not evaluate spruce logic after merging docs'"`
	Prune          []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once)'"`
	CherryPick     []string           `goptions:"--cherry-pick, description='The opposite of prune, specify keys to cherry-pick from final output (may be specified more than once)'"`
	FallbackAppend bool               `goptions:"--fallback-append, description='Default merge normally tries to key merge, then inline. This flag says do an append instead of an inline.'"`
	EnableGoPatch  bool               `goptions:"--go-patch, description='Enable the use of go-patch when parsing files to be merged'"`
	Files          goptions.Remainder `goptions:"description='Merges file2.yml through fileN.yml on top of file1.yml. To read STDIN, specify a filename of \\'-\\'.'"`
}

func main() {
	var options struct {
		Debug     bool `goptions:"-D, --debug, description='Enable debugging'"`
		Trace     bool `goptions:"-T, --trace, description='Enable trace mode debugging (very verbose)'"`
		Version   bool `goptions:"-v, --version, description='Display version information'"`
		Concourse bool `goptions:"--concourse, description='Pre/Post-process YAML for Concourse CI (handles {{ }} quoting)'"`
		Action    goptions.Verbs
		Merge     mergeOpts `goptions:"merge"`
		JSON      struct {
			Files goptions.Remainder `goptions:"description='Files to convert to JSON'"`
		} `goptions:"json"`
		Diff struct {
			Files goptions.Remainder `goptions:"description='Show the semantic differences between two YAML files'"`
		} `goptions:"diff"`
		VaultInfo struct {
			EnableGoPatch bool               `goptions:"--go-patch, description='Enable the use of go-patch when parsing files to be merged'"`
			Files         goptions.Remainder `goptions:"description='List vault references in the given files'"`
		} `goptions:"vaultinfo"`
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
		tree, err := cmdMergeEval(options.Merge)
		if err != nil {
			PrintfStdErr("%s\n", err.Error())
			exit(2)
			return
		}

		TRACE("Converting the following data back to YML:")
		TRACE("%#v", tree)
		merged, err := yaml.Marshal(tree)
		if err != nil {
			PrintfStdErr("Unable to convert merged result back to YAML: %s\nData:\n%#v", err.Error(), tree)
			exit(2)
			return
		}

		var output string
		if handleConcourseQuoting {
			PrintfStdErr(ansi.Sprintf("@Y{--concourse is deprecated. Consider using built-in spruce operators when merging Concourse YAML files}\n"))
			output = dequoteConcourse(merged)
		} else {
			output = string(merged)
		}
		printfStdOut("%s\n", output)

	case "vaultinfo":
		VaultRefs = map[string][]string{}
		SkipVault = true
		options.Merge.Files = options.VaultInfo.Files
		options.Merge.EnableGoPatch = options.VaultInfo.EnableGoPatch
		_, err := cmdMergeEval(options.Merge)
		if err != nil {
			PrintfStdErr("%s\n", err.Error())
			exit(2)
			return
		}

		printfStdOut("%s\n", formatVaultRefs())
	case "json":
		if len(options.JSON.Files) > 0 {
			jsons, err := JSONifyFiles(options.JSON.Files)
			if err != nil {
				PrintfStdErr("%s\n", err)
				exit(2)
				return
			}
			for _, output := range jsons {
				printfStdOut("%s\n", output)
			}
		} else {
			output, err := JSONifyIO(os.Stdin)
			if err != nil {
				PrintfStdErr("%s\n", err)
				exit(2)
				return
			}
			printfStdOut("%s\n", output)
		}

	case "diff":
		ansi.Color(isatty.IsTerminal(os.Stdout.Fd()))
		if len(options.Diff.Files) != 2 {
			usage()
			return
		}
		output, differences, err := diffFiles(options.Diff.Files)
		if err != nil {
			PrintfStdErr("%s\n", err)
			exit(2)
			return
		}
		printfStdOut("%s\n", output)
		if differences {
			exit(1)
		}

	default:
		usage()
		return
	}
}

func isArrayError(err error) bool {
	_, ok := err.(RootIsArrayError)
	return ok
}

func parseGoPatch(data []byte) (patch.Ops, error) {
	opdefs := []patch.OpDefinition{}
	err := yaml.Unmarshal(data, &opdefs)
	if err != nil {
		return nil, ansi.Errorf("@R{Root of YAML document is not a hash/map. Tried parsing it as go-patch, but got}: %s\n", err)
	}
	ops, err := patch.NewOpsFromDefinitions(opdefs)
	if err != nil {
		return nil, ansi.Errorf("@R{Unable to parse go-patch definitions: %s\n", err)
	}
	return ops, nil
}
func parseYAML(data []byte) (map[interface{}]interface{}, error) {
	y, err := simpleyaml.NewYaml(data)
	if err != nil {
		return nil, err
	}

	doc, err := y.Map()
	if err != nil {
		if _, arrayErr := y.Array(); arrayErr == nil {
			return nil, RootIsArrayError{msg: ansi.Sprintf("@R{Root of YAML document is not a hash/map}: %s\n", err)}
		}
		return nil, ansi.Errorf("@R{Root of YAML document is not a hash/map}: %s\n", err.Error())
	}

	return doc, nil
}

func cmdMergeEval(options mergeOpts) (map[interface{}]interface{}, error) {
	if len(options.Files) < 1 {
		options.Files = append(options.Files, "-")
	}

	root := make(map[interface{}]interface{})

	err := mergeAllDocs(root, options.Files, options.FallbackAppend, options.EnableGoPatch)
	if err != nil {
		return nil, err
	}

	ev := &Evaluator{Tree: root, SkipEval: options.SkipEval}
	err = ev.Run(options.Prune, options.CherryPick)
	return ev.Tree, err
}

type yamlVaultSecret struct {
	Key        string
	References []string
}

type byKey []yamlVaultSecret

type yamlVaultRefs struct {
	Secrets []yamlVaultSecret
}

func (refs byKey) Len() int           { return len(refs) }
func (refs byKey) Swap(i, j int)      { refs[i], refs[j] = refs[j], refs[i] }
func (refs byKey) Less(i, j int) bool { return refs[i].Key < refs[j].Key }

func formatVaultRefs() string {
	refs := yamlVaultRefs{}
	for secret, srcs := range VaultRefs {
		refs.Secrets = append(refs.Secrets, yamlVaultSecret{secret, srcs})
	}

	sort.Sort(byKey(refs.Secrets))
	for _, secret := range refs.Secrets {
		sort.Strings(secret.References)
	}

	output, err := yaml.Marshal(refs)
	if err != nil {
		panic(fmt.Sprintf("Could not marshal YAML for vault references: %+v", VaultRefs))
	}

	return string(output)
}

func mergeAllDocs(root map[interface{}]interface{}, paths []string, fallbackAppend bool, goPatchEnabled bool) error {
	m := &Merger{AppendByDefault: fallbackAppend}
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
			if isArrayError(err) && goPatchEnabled {
				DEBUG("Detected root of document as an array. Attempting go-patch parsing")
				ops, err := parseGoPatch(data)
				if err != nil {
					return ansi.Errorf("@m{%s}: @R{%s}\n", path, err.Error())
				}
				newObj, err := ops.Apply(root)
				if err != nil {
					return ansi.Errorf("@m{%s}: @R{%s}\n", path, err.Error())
				}
				if newRoot, ok := newObj.(map[interface{}]interface{}); !ok {
					return ansi.Errorf("@m{%s}: @R{Unable to convert go-patch output into a hash/map for further merging|\n", path)
				} else {
					root = newRoot
				}
			} else {
				return ansi.Errorf("@m{%s}: @R{%s}\n", path, err.Error())
			}
		} else {
			m.Merge(root, doc)
		}
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

func diffFiles(paths []string) (string, bool, error) {
	if len(paths) != 2 {
		return "", false, ansi.Errorf("incorrect number of files given to diffFiles(); please file a bug report")
	}

	data, err := ioutil.ReadFile(paths[0])
	if err != nil {
		return "", false, ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", paths[0], err)
	}
	a, err := parseYAML(data)
	if err != nil {
		return "", false, ansi.Errorf("@m{%s}: @R{%s}\n", paths[0], err)
	}

	data, err = ioutil.ReadFile(paths[1])
	if err != nil {
		return "", false, ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", paths[1], err)
	}
	b, err := parseYAML(data)
	if err != nil {
		return "", false, ansi.Errorf("@m{%s}: @R{%s}\n", paths[1], err)
	}

	d, err := Diff(a, b)
	if err != nil {
		return "", false, ansi.Errorf("@R{Failed to diff} @m{%s} -> @m{%s}: %s\n", paths[0], paths[1], err)
	}

	if !d.Changed() {
		return ansi.Sprintf("@G{both files are semantically equivalent; no differences found!}\n"), false, nil
	}
	return d.String("$"), true, nil
}

type RootIsArrayError struct {
	msg string
}

func (r RootIsArrayError) Error() string {
	return r.msg
}
