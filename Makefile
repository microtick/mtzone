VERSIONFILE=app/version.go
TEST_DOCKER_REPO=microtick/mtzonetest
GOPATH=$(shell go env GOPATH)

all: install

install:
	$(eval override VERSION = $(shell git describe --tags))
	$(eval override DATE = $(shell date))
	$(eval override HOST = $(shell hostname))
	$(eval override COMMIT = $(shell git log -1 --format='%H'))
	@echo "package app;" > $(VERSIONFILE)
	@echo "const MTAppVersion = \"$(VERSION)\"" >> $(VERSIONFILE)
	@echo "const MTBuildDate = \"$(DATE)\"" >> $(VERSIONFILE)
	@echo "const MTHostBuild = \"$(HOST)\"" >> $(VERSIONFILE)
	@echo "const MTCommit = \"$(COMMIT)\"" >> $(VERSIONFILE)
	GO111MODULE=on go install ./cmd/mtd 
	GO111MODULE=on go install -tags="ledger" ./cmd/mtcli
	@cp $(GOPATH)/bin/mtd .
	@cp $(GOPATH)/bin/mtcli .

test-docker:
	@docker build -f contrib/Dockerfile.test -t ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) .
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:latest
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD)
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker push ${TEST_DOCKER_REPO}:latest
