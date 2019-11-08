
all: install

install:
	$(eval override VERSION = $(shell git describe))
	@echo "package app;" > version.go
	@echo "const MTAppVersion = \"$(VERSION)\"" >> version.go
	GO111MODULE=on go install ./cmd/mtd 
	GO111MODULE=on go install ./cmd/mtcli
