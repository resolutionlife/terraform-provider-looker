# Terraform Provider for Looker

This terraform provider interacts with the Looker API to configure Looker resources.

## Documentation 

This provider is not yet published to the terraform registry. Documentation for each resource and datasource supported can be found below.

- [Resources](https://github.com/resolutionlife/terraform-provider-looker/tree/main/docs/resources) 
- [Data sources](https://github.com/resolutionlife/terraform-provider-looker/tree/main/docs/data-sources)
## Installation

Terraform uses the Terraform Registry to download and install providers. This provider is not currently published on the terraform registry, so the provider binary must be built locally. 

To build a binary locally, run `make install`. Then, run `terraform init` using the below configuration.

```terraform
terraform {
  required_providers {
    looker = {
      source  = "terraform.example.com/local/looker"
      version = "0.0.1"
    }
  }
}

provider "looker" {
    base_url      = "https://my-instance.cloud.looker.com"
    client_id     = "my-client-id"
    client_secret = "my-client-secret"
}
```

```sh
$ make install
```
```sh
$ terraform init
```

## Environment Variables

You can configure the provider with the `LOOKERSDK_BASE_URL`,
`LOOKERSDK_CLIENT_ID`, `LOOKERSDK_CLIENT_SECRET` environment variables. You can
also skip SSL verification with `LOOKERSDK_VERIFY_SSL` and define the timeout
duration with `LOOKERSDK_TIMEOUT`.  For example 

```shell
LOOKERSDK_BASE_URL="<my-instance-url>" \
LOOKERSDK_CLIENT_ID="<my-client-id>" \
LOOKERSDK_CLIENT_SECRET="<my-client-secret>" \
```

## Acceptance testing 

To run acceptance testing, run the following 
 ```
 make testacc
 ``` 

## Logging and Debugging 

This provider supports logging and debugging to provide insights and aid debugging. To view the log outputs, set the `TF_LOG_PROVIDER` enviroment variable to the desired log level. For example: 

```
export TF_LOG_PROVIDER=INFO
```

See the [official documentation](https://www.terraform.io/plugin/log/managing#log-levels) for details on each log level.