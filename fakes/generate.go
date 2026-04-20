package fakes

//go:generate go tool counterfeiter -o fake_ssm_api.go github.com/aws/aws-sdk-go/service/ssm/ssmiface.SSMAPI
//go:generate go tool counterfeiter -o fake_secrets_manager_api.go github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface.SecretsManagerAPI
