.PHONY: all build bench clean cover deflake fmt lint test test-clean

GOENV=GO111MODULE=on
GO=${GOENV} go

COVERAGE_OUT=/tmp/coverage.out
PACKAGE=./...

TEST_CLAUSE= $(if ${TEST}, -run ${TEST})

.PHONY: build
build:
	${GO} build ./...

.PHONY: bench
bench:
	${GO} test -short -bench=. -test.timeout=0 -run=^noTests ./...

.PHONY: clean
clean:
	${GO} clean -cache -modcache -i -r

.PHONY: cover
cover:
	${GO} test -short -coverprofile=${COVERAGE_OUT} ${PACKAGE}
	${GO} tool cover -html=${COVERAGE_OUT}

.PHONY: deflake
deflake:
	${GO} test -count=1000 -short -timeout 0 ${TEST_CLAUSE} ./...

.PHONY: lint
lint:
	${GO} vet ./...
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	${GO} vet ./...
	golangci-lint run --fix

.PHONY: test
test:
	${GO} test -short ${TEST_CLAUSE} ./...

.PHONY: test-clean
test-clean:
	${GO} clean -testcache && ${GO} test -count=1 -short ${TEST_CLAUSE} ./...
