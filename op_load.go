package spruce

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"
)

// LoadOperator is invoked with (( load <location> ))
type LoadOperator struct{}

// Setup ...
func (LoadOperator) Setup() error {
	return nil
}

// Phase ...
func (LoadOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (LoadOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (LoadOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	log.DEBUG("running (( load ... )) operation at $.%s", ev.Here)
	defer log.DEBUG("done with (( load ... )) operation at $%s\n", ev.Here)

	if len(args) != 1 {
		return nil, ansi.Errorf("@R{load operator only expects} @r{one} @R{argument containing the location of the file to be loaded}")
	}

	// Extract the location information from the operator
	var location string
	switch args[0].Type {
	case Literal:
		location = fmt.Sprintf("%v", args[0].Literal)

	default:
		return nil, ansi.Errorf("@R{load operator only expects} @r{one} @R{literal argument (quoted string)}")
	}

	bytes, err := getBytesFromLocation(location)
	if err != nil {
		return nil, err
	}

	data, err := simpleyaml.NewYaml(bytes)
	if err != nil {
		return nil, err
	}

	if listroot, err := data.Array(); err == nil {
		return &Response{
			Type:  Replace,
			Value: listroot,
		}, nil
	}

	if maproot, err := data.Map(); err == nil {
		return &Response{
			Type:  Replace,
			Value: maproot,
		}, nil
	}

	return nil, fmt.Errorf("unsupported root type in loaded content, only map or list roots are supported")
}

func getBytesFromLocation(location string) ([]byte, error) {
	// Handle location as a URI if it looks like one
	if _, err := url.ParseRequestURI(location); err == nil {
		response, err := http.Get(location)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		data, err := ioutil.ReadAll(response.Body)
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to retrieve data from location %s: %s", location, string(data))
		}

		return data, err
	}

	// Preprend the optional Spruce base path override
	if !filepath.IsAbs(location) {
		location = filepath.Join(os.Getenv("SPRUCE_FILE_BASE_PATH"), location)
	}

	// Handle location as local file if there is a file at that location
	if _, err := os.Stat(location); err == nil {
		return ioutil.ReadFile(location)
	}

	// In any other case, bail out ...
	return nil, fmt.Errorf("unable to get any content using location %s: it is not a file or usable URI", location)
}

func init() {
	RegisterOp("load", LoadOperator{})
}
