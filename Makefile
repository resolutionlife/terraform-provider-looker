-include .env.local
VERSION=0.0.1
TF_PATH=local
PROVIDER_PATH=terraform.example.com/local/looker
OS_ARCH=darwin_amd64

default: local 

# removes existing binary and builds a new binary to terraform cache directory 
build: clean
	@mkdir -p ~/.terraform.d/plugins/$(PROVIDER_PATH)/$(VERSION)/$(OS_ARCH)
	@go build -o ~/.terraform.d/plugins/$(PROVIDER_PATH)/$(VERSION)/$(OS_ARCH)/terraform-provider-looker

# removes existing binary for looker terraform provider
clean: 
	@rm -f ~/.terraform.d/plugins/$(PROVIDER_PATH)/$(VERSION)/$(OS_ARCH)/terraform-provider-looker

# removes binary from terraform main.tf path, preserving the terraform state 
teardown: 
	@rm -rf $(TF_PATH)/.terraform
	@rm -rf $(TF_PATH)/.terraform.lock.hcl

# tears down previous terraform-provider-looker binary and reinitialises the terraform provider in the main.tf path
local: teardown build
	@cd $(TF_PATH) && terraform init
