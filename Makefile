MODULE         = github.com/venafi/csm-opa-plugin
GIT_TAG        = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
BUILD_METADATA =
ifeq ($(GIT_TAG),) # unreleased build
    GIT_COMMIT     = $(shell git rev-parse HEAD)
    GIT_STATUS     = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "unreleased")
	BUILD_METADATA = $(GIT_COMMIT).$(GIT_STATUS)
endif
LDFLAGS=-buildid= -X sigs.k8s.io/release-utils/version.gitVersion=$(GIT_VERSION) \
        -X sigs.k8s.io/release-utils/version.gitCommit=$(GIT_HASH) \
        -X sigs.k8s.io/release-utils/version.gitTreeState=$(GIT_TREESTATE) \
        -X sigs.k8s.io/release-utils/version.buildDate=$(BUILD_DATE)

GO_BUILD_FLAGS = --ldflags="$(LDFLAGS)"

PLATFORMS=darwin linux windows
ARCHITECTURES=amd64

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build

.PHONY: FORCE
FORCE:

.PHONY: build
build:
	go build $(GO_BUILD_FLAGS) -o bin/opa ./cmd/opa

.PHONY: cross
cross:
	$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); \
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o opa-$(GOOS)-$(GOARCH) ./cmd/opa ))) \
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o opa-darwin-arm64 ./cmd/opa
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o opa-linux-arm64 ./cmd/opa
	
.PHONY: download
download: ## download dependencies via go mod
	go mod download

.PHONY: clean
clean:
	git status --short | grep '^!! ' | sed 's/!! //' | xargs rm -rf

.PHONY: test
test:
	go test ./... -coverprofile cover.out

.PHONY: signing_test
signing_test: 
	./bin/opa build --bundle ./policy --output ./policy/bundle.tar.gz --signing-key vsign\\rsa2048-cert --signing-plugin csm-opa-plugin
	./bin/opa sign --bundle --signing-key vsign\\rsa2048-cert --signing-plugin csm-opa-plugin ./policy

.PHONY: run
run:
	./bin/opa run --bundle --verification-key vsign\\rsa2048-cert --verification-key-id vsign\\rsa2048-cert --exclude-files-verify data.json --exclude-files-verify policy/awesome.rego --exclude-files-verify .manifest --exclude-files-verify .signatures.json ./policy/bundle.tar.gz