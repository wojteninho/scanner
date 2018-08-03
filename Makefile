arg = $(filter-out $@,$(MAKECMDGOALS))

########################
### GLOBAL VARIABLES ###
########################

GO_PACKAGES := $(shell go list ./...)
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

###########
### APP ###
###########

install: ## install vendor dependencies
	dep ensure -v -vendor-only

#############
### TOOLS ###
#############

fmt: ## run gofmt command
	gofmt -s -l -w ${GO_FILES}

#####################
### STATIC CHECKS ###
#####################

static-check: static-check-vet ## run all static checks

static-check-vet: ## run go vet static check
	go vet ${GO_PACKAGES}

############
### TEST ###
############

test: ## run unit tests
	go test ${GO_PACKAGES} -v -race

test-with-coverage: ## run unit tests and generate coverage
	./codecov.sh

test-coverage-txt-to-html: ## transform coverage report to html
	go tool cover -html=coverage.txt -o coverage.html
