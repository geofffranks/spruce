package fakes

//go:generate go tool counterfeiter -o fake_ssm_client.go github.com/geofffranks/spruce.SSMClient
//go:generate go tool counterfeiter -o fake_secrets_manager_client.go github.com/geofffranks/spruce.SecretsManagerClient
