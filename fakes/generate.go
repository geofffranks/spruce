// Package fakes holds counterfeiter-generated fakes for the AWS client
// interfaces defined in the root spruce package.
//
// Generated files have their default `var _ spruce.SSMClient = ...` type-assertion
// guards removed to avoid a spruce -> fakes -> spruce import cycle. Interface
// conformance is instead enforced at the test call site in operator_test.go
// (e.g. `parameterstoreClient = fakeSSM`). If you regenerate these fakes via
// `go generate`, strip the type-assertion lines and the `spruce` import from the
// generated files.
package fakes

//go:generate go tool counterfeiter -o fake_ssm_client.go github.com/geofffranks/spruce.SSMClient
//go:generate go tool counterfeiter -o fake_secrets_manager_client.go github.com/geofffranks/spruce.SecretsManagerClient
