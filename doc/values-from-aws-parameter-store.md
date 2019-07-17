## How can I pull values from AWS Parameter Store?

The `(( awsparam ))` operator can be used to fetch values from AWS Parameter Store.

As an example, with a couple of secrets pushed to the parameter store like so:
```
$ aws ssm put-parameter --name="/services/myservice/db" --value='{"user": "myuser", "pass": "mypass"}' --type=SecureString --key-id=<your-kms-key-id>
$ aws ssm put-parameter --name="/services/myservice/test" --value='test' --type=SecureString --key-id=<your-kms-key-id>
```

And `base.yml` defined as:
```
service: myservice
database:
  password: (( awsparam "/services/myservice/db?key=pass" ))
  test: (( awsparam "/services/myservice/test" ))
  username: (( awsparam "/services/" service "/db?key=user" ))
```

When we run `spruce merge base.yml` we will get:
```
$ spruce merge base.yml
database:
  password: mypass
  test: test
  username: myuser
```

If we omit the `key` argument then we'd get the unparsed JSON back instead.

It is not currently possible to parse the value and merge it as a structure.

## Which SSM parameter store types are supported?
The `(( awsparam ))` operator supports all types currently available (`String`, `SecureString`, `StringList`) but will return `StringList` types as a single comma separated string rather than a list.

## What IAM permissions are required?
In order to execute `spruce merge` with `awsparam` the calling user must have the appropriate IAM policy to `ssm:GetParameter` and, for `SecureString` type values, the permission to `kms:Decrypt` with the KMS key used to encrypt the value.

## How can I use a profile / role?
The `(( awsparam ))` operator uses three optional environment variables to determine role, profile and region configuration when establishing an AWS session.

These are:
- `AWS_REGION` - AWS region to use
- `AWS_ROLE` - AWS IAM role to assume
- `AWS_PROFILE` - AWS profile to use (typically defined in a combination of `~/.aws/config` / `~/.aws/credentials`)