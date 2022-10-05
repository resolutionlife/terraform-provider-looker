# Terraform Provider for Looker

This terraform provider interacts with the Looker API to configure Looker resources.

## Documentation 

Documentation for each resource and datasource supported can be found [here](https://registry.terraform.io/providers/resolutionlife/looker/latest/docs).

## Installation

Terraform uses the Terraform Registry to download and install providers. To install this provider, copy and paste the following code into your Terraform configuration. Then, run terraform init.

```terraform
terraform {
  required_providers {
    auth0 = {
      source  = "resolutionlife/looker"
      version = ">= 0.1.0" # See docs for latest version`
    }
  }
}

provider "looker" {}

```

```sh
$ terraform init
```
### Running the provider locally

To run the terraform provider locally, run `make install`. This command builds a binary of the provider and store it locally. Then, run `terraform init` using the below configuration.

_Note: To configure the provider, an API key and secret is required for your looker instance. [See documentation on creating an API key](https://cloud.google.com/looker/docs/admin-panel-users-users#edit_api3_keys). You must be an admin to create an API key._

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
## Logging and Debugging 

This provider supports logging and debugging to provide insights and aid debugging. To view the log outputs, set the `TF_LOG_PROVIDER` enviroment variable to the desired log level. For example: 

```
export TF_LOG_PROVIDER=INFO
```

See the [official documentation](https://www.terraform.io/plugin/log/managing#log-levels) for details on each log level.

## Acceptance testing 

Acceptance tests are run against a test looker instance as part of the developer workflow. To run acceptance testing locally, run the following:
 ```
 make testacc
```
Sweepers are available to clean up dangling resources that can occur when acceptance tests fail. To run the sweeper, run the following:

```
make sweep
```
