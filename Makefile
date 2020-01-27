VERSIONFILE=app/version.go

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
