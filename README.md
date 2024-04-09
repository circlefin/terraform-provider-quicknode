# Terraform Provider Quicknode

This provider allows for managing [Quicknode](https://www.quicknode.com/) resources via Terraform.
The structure of the repository is outlined below:
- A resource and a data source (`internal/provider/`),
- Examples (`examples/`)
- Generated documentation (`docs/`),

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

The provider is intended to be configured:
```hcl
terraform {
  required_providers {
    quicknode = {
      source = "registry.terraform.io/hashicorp/quicknode"
    }
  }
}

provider "quicknode" {
  // endpoint = "https://api.quicknode.com"

  // Also set via QUICKNODE_APIKEY
  // apikey = "todo"
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation or other generated code, run `make generate`.

### Acceptance Tests
In order to run the full suite of Acceptance tests, run `make testacc`, ensuring that the environment variable QUICKNODE_APIKEY is set with an apikey with at least the scope `CONSOLE_REST`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
QUICKNODE_APIKEY="qn_******" make testacc
```

### On My Machine
In order to use a compiled provider for a local terraform plan/apply. Configure your `~.terraformrc` as follows:

```hcl
provider_installation {
  dev_overrides {
      "registry.terraform.io/hashicorp/quicknode" = "/path/to/go/bin"
  }
}
```

This made be done via the following command:
```shell
tee ~/.terraformrc <<EOT
provider_installation {
  dev_overrides {
      "registry.terraform.io/hashicorp/quicknode" = "$(echo `go env GOPATH`/bin)"
  }
  direct {}
}
EOT
```
