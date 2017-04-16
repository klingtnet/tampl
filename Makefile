.PHONY: test

all: test tampl

tampl: tampl.go
	go install github.com/klingtnet/tampl

test: tampl_test.go
	go test github.com/klingtnet/tampl	

