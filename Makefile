REPOSITORY    := your-registry
NAME          := 01cloud-api
BRANCH        := $(shell git rev-parse --abbrev-ref HEAD)
HASH          := $(shell git rev-parse --short HEAD)
SUFFIX        ?= -$(subst /,-,$(BRANCH))
VERSION       := 0.1
DIRNAME       := $(shell basename $(CURDIR))
PARENTDIRNAME := $(shell basename $(shell dirname $(CURDIR)))

.PHONY: help all test format fmtcheck vet lint     qa deps clean nuke build

# Display general help about this command
help:
	@echo ""
	@echo "The following commands are available:"
	@echo ""
	@echo "    make qa          : Run all the tests"
	@echo "    make unit-test        : Run the unit tests"
	@echo ""
	@echo "    make format      : Format the source code"
	@echo "    make fmtcheck    : Check if the source code has been formatted"
	@echo "    make vet         : Check for suspicious constructs"
	@echo "    make lint        : Check for style errors"
	@echo ""
	@echo "    make deps        : Get the dependencies"
	@echo "    make mock-dev    : Generete mock data from mockgen"
	@echo "    make clean       : Remove any build artifact"
	@echo "    make nuke        : Deletes any intermediate file"
	@echo ""

.PHONY: unit-test
unit-test:
	@docker build . --target unit-test

.PHONY: unit-test-coverage
unit-test-coverage:
	@docker build . --target unit-test-coverage \
	--output coverage/
	cat coverage/cover.out

.PHONY: lint
lint:
	@docker build . --ssh default --target lint

.PHONY: test
test:
	@docker build . --ssh default --target test


.PHONY: prepare
prepare:
	export DOCKER_BUILDKIT=1
	eval `ssh-agent`
	ssh-add ~/.ssh/id_rsa

.PHONY: test-dev
test-dev:
	go test ./... --coverprofile=cover.out
	go tool cover -func=cover.out

.PHONY: test-func
test-func:
	go tool cover -func=cover.out

.PHONY: run-dev
run-dev:
	go run main.go

mock-dev:
	mockgen -package mockdb -destination  ./api/controllers/mocks/group_mock.go -source=api/models/Groups.go
# mockgen -package mockdb -destination  ./api/controllers/mocks/mock.go -source=api/models/Activities_interface.go
# Alias for help target
all: help

# Format the source code
format:
	@find ./ -type f -name "*.go" -exec gofmt -w {} \;

# Check if the source code has been formatted
fmtcheck:
	@mkdir -p target
	@find ./ -type f -name "*.go" -exec gofmt -d {} \; | tee target/format.diff
	@test ! -s target/format.diff || { echo "ERROR: the source code has not been formatted - please use 'make format' or 'gofmt'"; exit 1; }

# Check for syntax errors
vet:
	GOPATH=$(GOPATH) go vet ./...



# Alias to run all quality-assurance checks
qa: fmtcheck test vet lint

# --- INSTALL ---

# Get the dependencies
deps:

# Remove any build artifact
clean:
	GOPATH=$(GOPATH) go clean ./...

# Deletes any intermediate file
nuke:
	rm -rf ./target
	GOPATH=$(GOPATH) go clean -i ./...

build:
	@echo ">> building $(REPOSITORY)/$(NAME):$(HASH)"
	@docker build --ssh default -t "$(REPOSITORY)/$(NAME):$(HASH)" .
	#@docker tag "$(REPOSITORY)/$(NAME):$(VERSION)-$(DIRNAME)$(SUFFIX)" "$(REPOSITORY)/$(NAME):$(PARENTDIRNAME)-$(DIRNAME)$(SUFFIX)"
