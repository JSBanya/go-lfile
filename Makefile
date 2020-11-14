export GOPATH=$(CURDIR)

test:
	go test -v -count=1

.PHONY: all test