VERSION?=$(shell git describe --tags --always --abbrev=0)
OS?=$(shell uname -s | tr '[:upper:]' '[:lower:]')
# ARCH:=$(shell uname -m)
ARCH?=amd64
PROVIDER_PATH=uswitch.com/segment/segment/$(VERSION)/$(OS)_$(ARCH)
EXE=./build/terraform-provider-segment_$(VERSION)_$(OS)_$(ARCH)

.PHONY: build
build:
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -o $(EXE)
	mkdir -p ~/.terraform.d/plugins/$(PROVIDER_PATH)
	cp $(EXE) ~/.terraform.d/plugins/$(PROVIDER_PATH)/terraform-provider-segment
	rm -rf example/.terraform
	rm -f example/.terraform.lock.hcl

.PHONY: release
release:
	BUMPED=$$(bin/bump.sh $(VERSION) $(TYPE) $(BETA)); \
	git tag $${BUMPED}; \
	echo "tag $${BUMPED}" created
	git push --tags

test:
	go test ./...

testacc:
	TF_ACC=1 SEGMENT_ACCESS_TOKEN=$(SEGMENT_ACCESS_TOKEN) SEGMENT_WORKSPACE=$(SEGMENT_WORKSPACE) go test ./segment -v

sweep:
	TF_ACC=1 SEGMENT_ACCESS_TOKEN=$(SEGMENT_ACCESS_TOKEN) SEGMENT_WORKSPACE=$(SEGMENT_WORKSPACE) go test ./segment -sweep t -v