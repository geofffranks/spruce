## How can I pull values from AWS Secrets Manager?

The `(( awssecret ))` operator can be used to fetch values from AWS Secrets Manager.

As an example, with a secret pushed to secrets manager like so:
```
$ aws secretsmanager create-secret --name="services/myservice/db" --secret-string='{"user": "myuser", "pass": "mypass"}' --kms-key-id=<your-kms-key-id>
```

And `base.yml` defined as:
```
database:
  password: (( awssecret "services/myservice/db?key=pass" ))
  username: (( awssecret "services/myservice/db?key=user" ))
```

When we run `spruce merge base.yml` we will get:
```
$ spruce merge base.yml
database:
  password: mypass
  username: myuser
```

If we omit the `key` argument then we'd get the unparsed JSON back instead.

It is not currently possible to parse the value and merge it as a structure.

## How can I get a secret at a specific version / stage?
In order to get a specific version / stage of a secret you can pass the `stage=...` or `version=...` argument.
Note that only one of these will apply if you specify both and `stage` takes precedence.
As per the AWS API if neither `stage` nor `version` are provided then the equivalent of `stage=AWSCURRENT` is used.

Example:
```
pinned_api_key: (( awssecret "external_service/api_key?version=ccb5ef6f-43fc-409d-87d1-fe099d465074" ))
previous_api_key: (( awssecret "external_service/api_key?stage=AWSPREVIOUS" ))
current_api_key: (( awssecret "external_service/api_key" ))
```

Note that you can combine `stage` / `version` with `key` like so:
```
some_value: (( awssecret "external_service/some_value?stage=AWSPREVIOUS&key=subkey" ))
```

## Can I retrieve secrets stored as binary?
It is not currently possible to retrieve secrets in binary as spruce is primarily a tool for working with text based formats.
If you need to push something binary consider base64 encoding your binary data and pushing that as a string.

## What IAM permissions are required?
The only permission required to use `(( awssecret ))` is `secretsmanager:GetSecretValue` on the secret(s) you need to access.

## How can I use a profile / role?
The `(( awssecret ))` operator uses three optional environment variables to determine role, profile and region configuration when establishing an AWS session.

These are:
- `AWS_REGION` - AWS region to use
- `AWS_ROLE` - AWS IAM role to assume
- `AWS_PROFILE` - AWS profile to use (typically defined in a combination of `~/.aws/config` / `~/.aws/credentials`)