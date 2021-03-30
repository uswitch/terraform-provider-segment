VERSION?=$(shell git describe --tags --always --abbrev=0)
OS?=$(shell uname -s | tr '[:upper:]' '[:lower:]')
# ARCH:=$(shell uname -m)
ARCH?=amd64
PROVIDER_PATH=uswitch.com/segment/segment/$(VERSION)/$(OS)_$(ARCH)
EXE=./build/terraform-provider-segment_$(VERSION)_$(OS)_$(ARCH)

.PHONY: build

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(EXE)
	mkdir -p ~/.terraform.d/plugins/$(PROVIDER_PATH)
	cp $(EXE) ~/.terraform.d/plugins/$(PROVIDER_PATH)/terraform-provider-segment

release:
	BUMPED=$$(bin/bump.sh $(VERSION) $(TYPE)); \
	git tag $${BUMPED}; \
	echo "tag $${BUMPED}" created
	git push --tags
