# Terraform Provider for Looker

This terraform provider interacts with the Looker API to configure Looker resources.

## Documentation 

Documentation for each resource and data source supported can be found [here](https://registry.terraform.io/providers/resolutionlife/looker/latest/docs).

## Installation

Terraform uses the Terraform Registry to download and install providers. To install this provider, copy and paste the following code into your Terraform configuration. Then, run terraform init.

```terraform
terraform {
  required_providers {
    looker = {
      source  = "resolutionlife/looker"
      version = ">= 0.2.0" # See docs for latest version`
    }
  }
}

provider "looker" {}

```

```sh
$ terraform init
```
### Running the provider locally

To run the provider locally, please see the [contributors guide](TODO)