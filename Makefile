
all: install

install:
	$(eval override VERSION = $(shell git describe))
	$(eval override DATE = $(shell date))
	@echo "package app;" > version.go
	@echo "const MTAppVersion = \"$(VERSION)\"" >> version.go
	@echo "const MTBuildDate = \"$(DATE)\"" >> version.go
	GO111MODULE=on go install ./cmd/mtd 
	GO111MODULE=on go install ./cmd/mtcli
