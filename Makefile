SOURCES := $(shell find . -name '*.go') ## when any of these files change, we need to rebuild the binary

VERSION=$(shell git describe --tags --long --dirty 2>/dev/null)

## we must have tagged the repo at least once for VERSION to work
ifeq ($(VERSION),)
	VERSION = UNKNOWN
endif

web: $(SOURCES)
	go build -ldflags "-X main.version=${VERSION}" -o $@ ./cmd/web

.PHONY: lint
lint:
	golangci-lint run

.PHONY: committed
committed:
	@git diff --exit-code > /dev/null || (echo "** COMMIT YOUR CHANGES FIRST **"; exit 1)

docker: $(SOURCES) build/Dockerfile
	echo "VERSION=$(VERSION)" > .env
	docker-compose -f build/docker-compose.yml build --build-arg VERSION=$(VERSION)

.PHONY: run
run:
    docker-compose -f build/docker-compose.yml up

.PHONY: publish
publish: committed #lint
	make docker
	docker tag  snippets:latest vladcoman/snippets:$(VERSION)
	docker push vladcoman/snippets:$(VERSION)
