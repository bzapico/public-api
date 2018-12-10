#
#  Copyright 2018 Nalej
# 

# Name of the target applications to be built
APPS=public-api public-api-cli

# Target directory to store binaries and results
TARGET=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Docker configuration
AZURE_CR=nalejregistry
DOCKER_REGISTRY=$(AZURE_CR).azurecr.io
DOCKER_REPO=nalej
VERSION=$(shell cat .version)



# Use ldflags to pass commit and branch information
# TODO: Integrate this into the compilation process
# Build information
COMMIT=$(shell git rev-parse HEAD)
#BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
# LDFLAGS to setup the version and commit. Notice that because of changes in go 1.10, we target a
# variable inside the main package that is being built.
LDFLAGS=-ldflags "-X main.MainVersion=${VERSION} -X main.MainCommit=${COMMIT}"

COVERAGE_FILE=$(TARGET)/coverage.out

.PHONY: all
all: dep build test yaml image

.PHONY: dep
dep:
	if [ ! -d vendor ]; then \
	    echo ">>> Create vendor folder" ; \
	    mkdir vendor ; \
	fi ;
	@echo ">>> Updating dependencies..."
	dep ensure -v

test-all: test test-race test-coverage

.PHONY: test test-race test-coverage
test:
	@echo ">>> Launching tests..."
	$(GOTEST) ./...

test-race:
	@echo ">>> Launching tests... (Race detector enabled)"
	$(GOTEST) -race ./...

test-coverage:
	@echo ">>> Launching tests... (Coverage enabled)"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic  ./...

# Check the codestyle using gometalinter
.PHONY: checkstyle
checkstyle:
	gometalinter --disable-all --enable=golint --enable=vet --enable=errcheck --enable=goconst --vendor ./...

# Run go formatter
.PHONY: format
format:
	@echo ">>> Formatting..."
	gofmt -s -w .

.PHONY: clean
clean:
	@echo ">>> Cleaning project..."
	$(GOCLEAN)
	rm -Rf $(TARGET)

.PHONY: dep build-all build build-linux build-local
build-all: dep format build build-linux
build: dep local
build-linux: dep linux

# Local compilation
local:
	@echo ">>> Building ..."
	for app in $(APPS); do \
		if [ -d cmd/"$$app" ]; then \
            $(GOBUILD) $(LDFLAGS) -o $(TARGET)/"$$app" ./cmd/"$$app" ; \
			echo Built $$app binary for your OS ; \
		fi ; \
	done

# Cross compilation to obtain a linux binary
linux:
	@echo ">>> Bulding for Linux..."
	for app in $(APPS); do \
		if [ -d cmd/"$$app" ]; then \
    		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(TARGET)/linux_amd64/"$$app" ./cmd/"$$app" ; \
			echo Built $$app binary for Linux ; \
		fi ; \
	done

yaml:
	@echo ">>> Creating K8s files..."
	for app in $(APPS); do \
		if [ -d components/"$$app"/appcluster ]; then \
			mkdir -p $(TARGET)/yaml/appcluster ; \
			cp components/"$$app"/appcluster/*.yaml $(TARGET)/yaml/appcluster/. ; \
			cd $(TARGET)/yaml/appcluster && find . -type f -name '*.yaml' | xargs sed -i '' 's/VERSION/$(VERSION)/g' && cd - ; \
		fi ; \
		if [ -d components/"$$app"/mngtcluster ]; then \
			mkdir -p $(TARGET)/yaml/mngtcluster ; \
			cp components/"$$app"/mngtcluster/*.yaml $(TARGET)/yaml/mngtcluster/. ; \
			cd $(TARGET)/yaml/mngtcluster && find . -type f -name '*.yaml' | xargs sed -i '' 's/VERSION/$(VERSION)/g' && cd - ; \
		fi ; \
	done

# Package all images and components
.PHONY: image create-image
image: build-linux create-image

create-image:
	@echo ">>> Creating docker images ..."
	for app in $(APPS); do \
        echo Create image of app $$app ; \
        if [ -f components/"$$app"/Dockerfile ]; then \
            docker build --no-cache -t $(DOCKER_REGISTRY)/$(DOCKER_REPO)/"$$app":$(VERSION) -f components/"$$app"/Dockerfile $(TARGET)/linux_amd64 ; \
			echo Built $$app Docker image ; \
        else  \
            echo $$app has no Dockerfile ; \
        fi ; \
    done

# Publish the image
.PHONY: publish az-login az-logout publish-image
publish: image az-login publish-image az-logout

az-login:
	@echo ">>> Logging in Azure and Azure Container Registry ..."
	az login
	az acr login --name $(AZURE_CR)

az-logout:
	az logout

publish-image:
	@echo ">>> Publishing images into Azure Container Registry ..."
	for app in $(APPS); do \
		if [ -f components/"$$app"/Dockerfile ]; then \
			docker push $(DOCKER_REGISTRY)/$(DOCKER_REPO)/"$$app":$(VERSION) ; \
		fi ; \
  done
