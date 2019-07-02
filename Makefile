all: install

install:
	GO111MODULE=on go install ./cmd/mtd
	GO111MODULE=on go install ./cmd/mtcli
