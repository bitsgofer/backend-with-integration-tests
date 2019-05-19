PROJECT=$$(echo "example")

clean:
	CURDIR=${CURDIR} docker-compose \
		-p ${PROJECT} \
		-f compose/infra.yml \
		-f compose/sourceCode.yml \
		-f compose/ci.yml \
		down --volumes
.PHONY: clean

build.tester:
	docker build --rm --tag=example/tester dockerfiles/tester
.PHONY: build.tester

test: test.copySrc
test: test.compile
test:
	CURDIR=${CURDIR} docker-compose \
		-p ${PROJECT} \
		-f compose/infra.yml \
		-f compose/ci.yml \
		up \
			--abort-on-container-exit \
			--exit-code-from ci.test
.PHONY: test

test.compile:
	CURDIR=${CURDIR} docker-compose \
		-p ${PROJECT} \
		-f compose/infra.yml \
		-f compose/sourceCode.yml \
		-f compose/ci.yml \
		up build.test
.PHONY: test.compile

test.copySrc:
	CURDIR=${CURDIR} docker-compose \
		-p ${PROJECT} \
		-f compose/infra.yml \
		-f compose/sourceCode.yml \
		up sourceCode.test
	docker cp ./ ${PROJECT}_sourceCode.test_1:/go/src/github.com/bitsgofer/example
.PHONY: test.copySrc

local.infra:
	CURDIR=${CURDIR} docker-compose \
		-p ${PROJECT} \
		-f compose/infra.yml \
		up
.PHONY: local.infra

build.img.ci:
	CURDIR=${CURDIR} docker-compose \
		-p ${PROJECT} \
		-f compose/infra.yml \
		-f compose/sourceCode.yml \
		-f compose/ci.yml \
		build ci.test
.PHONY: build.img.ci

govendor:
	GO111MODULE=on go mod vendor
.PHONY: govendor
