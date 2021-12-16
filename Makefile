VERSION?=$(shell git describe --tags --always --abbrev=0)
EXE=./build/terraform-provider-segment_$(VERSION)_$(OS)_$(ARCH)

.PHONY: build
build:
	go build -o $(EXE)

.PHONY: release
release:
	BUMPED=$$(bin/bump.sh $(VERSION) $(TYPE) $(BETA)); \
	git tag $${BUMPED}; \
	echo "tag $${BUMPED}" created
	git push --tags

.PHONY: test
test:
	go test ./...

.PHONY: testacc
testacc:
	TF_ACC=1 SEGMENT_ACCESS_TOKEN=$(SEGMENT_ACCESS_TOKEN) SEGMENT_WORKSPACE=$(SEGMENT_WORKSPACE) go test ./segment -v

.PHONY: sweep
sweep:
	TF_ACC=1 SEGMENT_ACCESS_TOKEN=$(SEGMENT_ACCESS_TOKEN) SEGMENT_WORKSPACE=$(SEGMENT_WORKSPACE) go test ./segment -sweep t -v
