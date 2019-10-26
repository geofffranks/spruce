package spruce

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/yaml"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

// awsSession holds a shared AWS session struct
var awsSession *session.Session

// secretsManagerClient holds a secretsmanager client configured with a session
// We use secretsmanageriface.SecretsManagerAPI to be able to provide mocks in testing
var secretsManagerClient secretsmanageriface.SecretsManagerAPI

// parameterstoreClient holds a parameterstore client configured with a session
// We use ssmiface.SSMAPI to be able to provide mocks in testing
var parameterstoreClient ssmiface.SSMAPI

// awsSecretsCache caches values from AWS Secretsmanager
var awsSecretsCache = make(map[string]string)

// awsParamsCache caches values from AWS Parameterstore
var awsParamsCache = make(map[string]string)

// SkipAws toggles whether AwsOperator will attempt to query AWS for any value
// When true will always return "REDACTED"
var SkipAws bool

// AwsOperator provides two operators;  (( awsparam "path" )) and (( awssecret "name_or_arn" ))
// It will fetch parameters / secrets from the respective AWS service
type AwsOperator struct {
	variant string
}

// initializeAwsSession will configure an AWS session with profile, region and role assume including loading shared config (e.g. ~/.aws/credentials)
func initializeAwsSession(profile string, region string, role string) (s *session.Session, err error) {
	options := session.Options{
		Config:            aws.Config{},
		SharedConfigState: session.SharedConfigEnable,
	}

	if region != "" {
		options.Config.Region = aws.String(region)
	}

	if profile != "" {
		options.Profile = profile
	}

	s, err = session.NewSessionWithOptions(options)
	if err != nil {
		return nil, err
	}

	if role != "" {
		options.Config.Credentials = stscreds.NewCredentials(s, role, func(p *stscreds.AssumeRoleProvider) {})
		s, err = session.NewSession(&options.Config)
	}

	return s, err
}

// getAwsSecret will fetch the specified secret from AWS Secretsmanager at the specified (if provided) stage / version
func getAwsSecret(awsSession *session.Session, secret string, params url.Values) (string, error) {
	val, cached := awsSecretsCache[secret]
	if cached {
		return val, nil
	}

	if secretsManagerClient == nil {
		secretsManagerClient = secretsmanager.New(awsSession)
	}

	input := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secret),
	}

	if params.Get("stage") != "" {
		input.VersionStage = aws.String(params.Get("stage"))
	} else if params.Get("version") != "" {
		input.VersionId = aws.String(params.Get("version"))
	}

	output, err := secretsManagerClient.GetSecretValue(&input)
	if err != nil {
		return "", err
	}

	awsSecretsCache[secret] = aws.StringValue(output.SecretString)

	return awsSecretsCache[secret], nil
}

// getAwsParam will fetch the specified parameter from AWS SSM Parameterstore
func getAwsParam(awsSession *session.Session, param string) (string, error) {
	val, cached := awsParamsCache[param]
	if cached {
		return val, nil
	}

	if parameterstoreClient == nil {
		parameterstoreClient = ssm.New(awsSession)
	}

	input := ssm.GetParameterInput{
		Name:           aws.String(param),
		WithDecryption: aws.Bool(true),
	}

	output, err := parameterstoreClient.GetParameter(&input)
	if err != nil {
		return "", err
	}

	awsParamsCache[param] = aws.StringValue(output.Parameter.Value)

	return awsParamsCache[param], nil
}

// Setup ...
func (AwsOperator) Setup() error {
	return nil
}

// Phase ...
func (AwsOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies is not used by AwsOperator
func (AwsOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run will invoke the appropriate getAws* function for each instance of the AwsOperator
// and extract the specified key (if provided).
func (o AwsOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	var err error
	DEBUG("running (( %s ... )) operation at $.%s", o.variant, ev.Here)
	defer DEBUG("done with (( %s ... )) operation at $.%s\n", o.variant, ev.Here)

	if len(args) < 1 {
		return nil, fmt.Errorf("%s operator requires at least one argument", o.variant)
	}

	var l []string
	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
			DEBUG("     [%d]: error was: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			DEBUG("  arg[%d]: using string literal '%v'", i, v.Literal)
			l = append(l, fmt.Sprintf("%v", v.Literal))

		case Reference:
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
			s, err := v.Reference.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
			}

			switch s.(type) {
			case map[interface{}]interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, ansi.Errorf("@c{$.%s}@R{ is a map; only scalars are supported here}", v.Reference)

			case []interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, ansi.Errorf("@c{$.%s}@R{ is a list; only scalars are supported here}", v.Reference)

			default:
				l = append(l, fmt.Sprintf("%v", s))
			}

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("%s operator only accepts string literals and key reference arguments", o.variant)
		}
	}

	key, params, err := parseAwsOpKey(strings.Join(l, ""))
	if err != nil {
		return nil, err
	}

	DEBUG("     [0]: Using %s key '%s'\n", o.variant, key)

	value := "REDACTED"

	if !SkipAws {
		if awsSession == nil {
			awsSession, err = initializeAwsSession(os.Getenv("AWS_PROFILE"), os.Getenv("AWS_REGION"), os.Getenv("AWS_ROLE"))
			if err != nil {
				return nil, fmt.Errorf("error during AWS session initialization: %s", err)
			}
		}

		if o.variant == "awsparam" {
			value, err = getAwsParam(awsSession, key)
		} else if o.variant == "awssecret" {
			value, err = getAwsSecret(awsSession, key, params)
		}

		if err != nil {
			return nil, fmt.Errorf("$.%s error fetching %s: %s", key, o.variant, err)
		}

		subkey := params.Get("key")
		if subkey != "" {
			tmp := make(map[string]interface{})
			err := yaml.Unmarshal([]byte(value), &tmp)

			if err != nil {
				return nil, fmt.Errorf("$.%s error extracting key: %s", key, err)
			}

			if _, ok := tmp[subkey]; !ok {
				return nil, fmt.Errorf("$.%s invalid key '%s'", key, subkey)
			}

			value = fmt.Sprintf("%v", tmp[subkey])
		}
	}

	return &Response{
		Type:  Replace,
		Value: value,
	}, nil
}

// parseAwsOpKey parsed the parameters passed to AwsOperator.
// Primarily it splits the key from the extra arguments (specified as a query string)
func parseAwsOpKey(key string) (string, url.Values, error) {
	split := strings.SplitN(key, "?", 2)
	if len(split) == 1 {
		split = append(split, "")
	}

	values, err := url.ParseQuery(split[1])
	if err != nil {
		return "", values, fmt.Errorf("invalid argument string: %s", err)
	}

	return split[0], values, nil
}

// init registers the two variants of the AwsOperator
func init() {
	RegisterOp("awsparam", AwsOperator{variant: "awsparam"})
	RegisterOp("awssecret", AwsOperator{variant: "awssecret"})
}
