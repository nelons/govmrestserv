GO   := go

DIRS_TO_CLEAN:=
FILES_TO_CLEAN:=

ifeq ($(origin GO), undefined)
  GO:=$(shell which go)
endif
ifeq ($(GO),)
  $(error Could not find 'go' in path. Please install go, or if already installed either add it to your path or set GO to point to its directory)
endif

#-------------------------
# Download libraries and tools
#-------------------------

.PHONY: get.tools

## Retrieve tools packages
get.tools:
	$(GO) get -u github.com/frapposelli/wwhrd

# ------------------------------------------------------
# Generate
# ------------------------------------------------------

.PHONY: swagger.generate

## Generate go code
swagger.generate:
	swagger generate server --quiet --target pkg/swagger/server --name vSphere --spec pkg/swagger/swagger.yml --exclude-main

# ------------------------------------------------------
# Validate Swagger YAML
# ------------------------------------------------------

.PHONY: swagger.validate

swagger.validate:
	swagger validate pkg/swagger/swagger.yml

# ------------------------------------------------------
# Generate Swagger Documentation
# ------------------------------------------------------

.PHONY: swagger.doc

swagger.doc:
	docker run -i yousan/swagger-yaml-to-html < pkg/swagger/swagger.yml > doc/index.html

# Swagger jagger

swagger: swagger.validate swagger.generate swagger.doc

# ------------------------------------------------------
# Dependencies
# ------------------------------------------------------

.PHONY: depend vendor.check depend.status depend.update depend.cleanlock depend.update.full

## Use go modules
depend: depend.tidy depend.verify depend.vendor

depend.tidy:
	@echo "==> Running dependency cleanup"
	$(GO) mod tidy -v

depend.verify:
	@echo "==> Verifying dependencies"
	$(GO) mod verify

depend.vendor:
	@echo "==> Freezing dependencies"
	$(GO) mod vendor

depend.update:
	@echo "==> Update go modules"
	$(GO) get -u -v

# ------------------------------------------------------
# License Check
# ------------------------------------------------------
.PHONY: license

license:
	@echo "==> license check"
	wwhrd check

# ------------------------------------------------------
# Build
# ------------------------------------------------------
.PHONY: build build.exe build.update build.full

build: build.update build.exe

build.full: depend build

build.update:
	$(GO) get ./internal

build.exe:
	$(GO) build -o bin/vsphere-rest-server.exe main.go