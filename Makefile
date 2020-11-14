export GOPATH=$(CURDIR)

test:
	go test lfile -v -count=1

.PHONY: all test