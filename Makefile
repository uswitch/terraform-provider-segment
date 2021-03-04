VERSION=0.0.9
OS:=$(shell uname -s | tr '[:upper:]' '[:lower:]')
# ARCH:=$(shell uname -m)
ARCH=amd64
PROVIDER_PATH=uswitch.com/segment/segment/$(VERSION)/$(OS)_$(ARCH)/

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o terraform-provider-segment
	mkdir -p ~/.terraform.d/plugins/$(PROVIDER_PATH)
	cp terraform-provider-segment ~/.terraform.d/plugins/$(PROVIDER_PATH)
