package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/cppforlife/go-patch/patch"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/gonvenience/ytbx"
	"github.com/mattn/go-isatty"
	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce"
	. "github.com/geofffranks/spruce/log"

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

func envFlag(varname string) bool {
	val := os.Getenv(varname)
	return val != "" && strings.ToLower(val) != "false" && val != "0"
}

type YamlFile struct {
	Path   string
	Reader io.ReadCloser
}

type jsonOpts struct {
	Strict bool               `goptions:"--strict, description='Refuse to convert non-string keys to strings'"`
	Help   bool               `goptions:"--help, -h"`
	Files  goptions.Remainder `goptions:"description='Files to convert to JSON'"`
}

type mergeOpts struct {
	SkipEval       bool               `goptions:"--skip-eval, description='Do not evaluate spruce logic after merging docs'"`
	Prune          []string           `goptions:"--prune, description='Specify keys to prune from final output (may be specified more than once)'"`
	CherryPick     []string           `goptions:"--cherry-pick, description='The opposite of prune, specify keys to cherry-pick from final output (may be specified more than once)'"`
	FallbackAppend bool               `goptions:"--fallback-append, description='Default merge normally tries to key merge, then inline. This flag says do an append instead of an inline.'"`
	EnableGoPatch  bool               `goptions:"--go-patch, description='Enable the use of go-patch when parsing files to be merged'"`
	MultiDoc       bool               `goptions:"--multi-doc, -m, description='Treat multi-doc yaml as multiple files.'"`
	Help           bool               `goptions:"--help, -h"`
	Files          goptions.Remainder `goptions:"description='List of files to merge. To read STDIN, specify a filename of \\'-\\'.'"`
}

func main() {
	var options struct {
		Debug   bool `goptions:"-D, --debug, description='Enable debugging'"`
		Trace   bool `goptions:"-T, --trace, description='Enable trace mode debugging (very verbose)'"`
		Version bool `goptions:"-v, --version, description='Display version information'"`
		Action  goptions.Verbs
		Merge   mergeOpts `goptions:"merge"`
		Fan     mergeOpts `goptions:"fan"`
		JSON    jsonOpts  `goptions:"json"`
		Diff    struct {
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

	if options.JSON.Help || options.Merge.Help || options.Fan.Help {
		usage()
		return
	}

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

		printfStdOut("%s\n", string(merged))

	case "fan":
		trees, err := cmdFanEval(options.Fan)
		if err != nil {
			PrintfStdErr("%s\n", err.Error())
			exit(2)
			return
		}

		for _, tree := range trees {
			TRACE("Converting the following data back to YML:")
			TRACE("%#v", tree)
			merged, err := yaml.Marshal(tree)
			if err != nil {
				PrintfStdErr("Unable to convert merged result back to YAML: %s\nData:\n%#v", err.Error(), tree)
				exit(2)
				return
			}

			printfStdOut("---\n%s\n", string(merged))
		}

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
		jsons, err := cmdJSONEval(options.JSON)
		if err != nil {
			PrintfStdErr("%s\n", err)
			exit(2)
			return
		}
		for _, output := range jsons {
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
	exit(0)
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

	if empty_y, _ :=  simpleyaml.NewYaml([]byte{}); *y == *empty_y {
		DEBUG("YAML doc is empty, creating empty hash/map")
		return make(map[interface{}]interface{}), nil
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

func loadYamlFile(file string) (YamlFile, error) {
	var target YamlFile
	if file == "-" {
		target = YamlFile{Reader: os.Stdin, Path: "-"}
	} else {
		f, err := os.Open(file)
		if err != nil {
			return YamlFile{}, ansi.Errorf("@R{Error reading file} @m{%s}: %s", file, err.Error())
		}
		target = YamlFile{Path: file, Reader: f}
	}
	return target, nil
}

func splitLoadYamlFile(file string) ([]YamlFile, error) {
	docs := []YamlFile{}

	yamlFile, err := loadYamlFile(file)
	if err != nil {
		return nil, err
	}

	fileData, err := readFile(&yamlFile)
	if err != nil {
		return nil, err
	}

	rawDocs := bytes.Split(fileData, []byte("\n---\n"))
	// strip off empty document created if the first three bytes of the file are the doc separator
	// keeps the indexing correct for when used with error messages
	if len(rawDocs[0]) == 0 {
		rawDocs = rawDocs[1:]
	}

	for i, docBytes := range rawDocs {
		buf := bytes.NewBuffer(docBytes)
		doc := YamlFile{Path: fmt.Sprintf("%s[%d]", yamlFile.Path, i), Reader: ioutil.NopCloser(buf)}
		docs = append(docs, doc)
	}
	return docs, nil
}

func cmdMergeEval(options mergeOpts) (map[interface{}]interface{}, error) {
	files := []YamlFile{}

	if len(options.Files) < 1 {
		stdinInfo, err := os.Stdin.Stat()
		if err != nil {
			return nil, ansi.Errorf("@R{Error statting STDIN} - Bailing out: %s\n", err.Error())
		}

		if stdinInfo.Mode()&os.ModeCharDevice != 0 {
			return nil, ansi.Errorf("@R{Error reading STDIN}: no data found. Did you forget to pipe data to STDIN, or specify yaml files to merge?")
		}

		options.Files = append(options.Files, "-")
	}

	for _, file := range options.Files {
		if options.MultiDoc {
			docs, err := splitLoadYamlFile(file)
			if err != nil {
				return nil, err
			}
			files = append(files, docs...)
		} else {
			yamlFile, err := loadYamlFile(file)
			if err != nil {
				return nil, err
			}
			files = append(files, yamlFile)
		}
	}

	ev, err := mergeAllDocs(files, options)
	if err != nil {
		return nil, err
	}

	return ev.Tree, nil
}

func cmdFanEval(options mergeOpts) ([]map[interface{}]interface{}, error) {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return nil, ansi.Errorf("@R{Error statting STDIN} - Bailing out: %s\n", err.Error())
	}
	if stdinInfo.Mode()&os.ModeCharDevice == 0 {
		options.Files = append(options.Files, "-")
	}

	if len(options.Files) == 0 {
		return nil, ansi.Errorf("@R{Missing Input:} You must specify at least a source document to spruce fan. If no files are specified, STDIN is used. Using STDIN for source and target docs only works with -m.")
	}

	roots := []map[interface{}]interface{}{}
	sourcePath := options.Files[0]
	options.Files = options.Files[1:]

	docs := []YamlFile{}
	source := YamlFile{}
	if options.MultiDoc {
		sourceDocs, err := splitLoadYamlFile(sourcePath)
		if err != nil {
			return nil, err
		}
		// only the first yaml document of the source will be treated as actual source, all others
		// will be treated as target documents
		source = sourceDocs[0]
		docs = append(sourceDocs[1:], docs...)
	} else {
		source, err = loadYamlFile(sourcePath)
		if err != nil {
			return nil, err
		}
	}

	for _, file := range options.Files {
		yamlDocs, err := splitLoadYamlFile(file)
		if err != nil {
			return nil, err
		}
		docs = append(docs, yamlDocs...)
	}

	sourceBytes, err := readFile(&source)
	if err != nil {
		return nil, err
	}

	if len(docs) < 1 {
		return nil, ansi.Errorf("@R{Missing Input:} You must specify at least one target document to spruce fan. If no files are specified, STDIN is used. Using STDIN for source and target docs only works with -m.")
	}

	for _, doc := range docs {
		sourceBuffer := bytes.NewBuffer(sourceBytes)
		source = YamlFile{Path: source.Path, Reader: ioutil.NopCloser(sourceBuffer)}
		ev, err := mergeAllDocs([]YamlFile{source, doc}, options)
		if err != nil {
			return nil, err
		}
		roots = append(roots, ev.Tree)
	}

	return roots, nil
}

func cmdJSONEval(options jsonOpts) ([]string, error) {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return nil, ansi.Errorf("@R{Error statting STDIN} - Bailing out: %s\n", err.Error())
	}
	if stdinInfo.Mode()&os.ModeCharDevice == 0 {
		options.Files = append(options.Files, "-")
	}

	output, err := JSONifyFiles(options.Files, options.Strict)
	if err != nil {
		return nil, err
	}

	return output, nil
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

func readFile(file *YamlFile) ([]byte, error) {
	var data []byte
	var err error

	if file.Path == "-" {
		file.Path = "STDIN"
		stat, err := os.Stdin.Stat()
		if err != nil {
			return nil, ansi.Errorf("@R{Error statting STDIN} - Bailing out: %s\n", err.Error())
		}
		if stat.Mode()&os.ModeCharDevice == 0 {
			data, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				return nil, ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", file.Path, err.Error())
			}
		}
	} else {
		data, err = ioutil.ReadAll(file.Reader)
		if err != nil {
			return nil, ansi.Errorf("@R{Error reading file} @m{%s}: %s\n", file.Path, err.Error())
		}
	}
	if len(data) == 0 && file.Path == "STDIN" {
		return nil, ansi.Errorf("@R{Error reading STDIN}: no data found. Did you forget to pipe data to STDIN, or specify yaml files to merge?")
	}

	return data, nil
}

func mergeAllDocs(files []YamlFile, options mergeOpts) (*Evaluator, error) {
	m := &Merger{AppendByDefault: options.FallbackAppend}
	root := make(map[interface{}]interface{})

	for _, file := range files {
		DEBUG("Processing file '%s'", file.Path)

		data, err := readFile(&file)
		if err != nil {
			return nil, err
		}

		doc, err := parseYAML(data)
		if err != nil {
			if isArrayError(err) && options.EnableGoPatch {
				DEBUG("Detected root of document as an array. Attempting go-patch parsing")
				ops, err := parseGoPatch(data)
				if err != nil {
					return nil, ansi.Errorf("@m{%s}: @R{%s}\n", file.Path, err.Error())
				}
				newObj, err := ops.Apply(root)
				if err != nil {
					return nil, ansi.Errorf("@m{%s}: @R{%s}\n", file.Path, err.Error())
				}
				if newRoot, ok := newObj.(map[interface{}]interface{}); !ok {
					return nil, ansi.Errorf("@m{%s}: @R{Unable to convert go-patch output into a hash/map for further merging|\n", file.Path)
				} else {
					root = newRoot
				}
			} else {
				return nil, ansi.Errorf("@m{%s}: @R{%s}\n", file.Path, err.Error())
			}
		} else {
			m.Merge(root, doc)
		}
		tmpYaml, _ := yaml.Marshal(root) // we don't care about errors for debugging
		TRACE("Current data after processing '%s':\n%s", file.Path, tmpYaml)
	}

	if m.Error() != nil {
		return nil, m.Error()
	}

	ev := &Evaluator{Tree: root, SkipEval: options.SkipEval}
	err := ev.Run(options.Prune, options.CherryPick)
	return ev, err
}

func diffFiles(paths []string) (string, bool, error) {
	if len(paths) != 2 {
		return "", false, ansi.Errorf("incorrect number of files given to diffFiles(); please file a bug report")
	}

	from, to, err := ytbx.LoadFiles(paths[0], paths[1])
	if err != nil {
		return "", false, err
	}

	report, err := dyff.CompareInputFiles(from, to)
	if err != nil {
		return "", false, err
	}

	reportWriter := &dyff.HumanReport{
		Report:            report,
		DoNotInspectCerts: false,
		NoTableStyle:      false,
		ShowBanner:        false,
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	reportWriter.WriteReport(out)
	out.Flush()

	return buf.String(), len(report.Diffs) > 0, nil
}

type RootIsArrayError struct {
	msg string
}

func (r RootIsArrayError) Error() string {
	return r.msg
}
