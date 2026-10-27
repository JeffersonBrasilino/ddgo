PACKAGES_TESTS := $(shell go list ./...)

# install dependencies
deps:
	go mod vendor
	go clean -modcache

# run go vet
vet:
	go vet $(PACKAGES_TESTS)

# run tests
test:
	go test -count=1 -race -v $(PACKAGES_TESTS)

# run tests with terminal coverage
coverage-terminal:
	go test -covermode=atomic -count=1 -race -coverprofile=coverage.out $(PACKAGES_TESTS)
	go tool cover -func=coverage.out

# run tests with html coverage
coverage-html:
	go test -count=1 -race -coverprofile=coverage.out $(PACKAGES_TESTS)
	go tool cover -html=coverage.out -o coverage.html

# build
build:
	go build ./...