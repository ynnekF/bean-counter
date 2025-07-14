export CWD 	:= $(shell pwd)

.PHONY: run
run:
	go run $(CWD)/cmd/main.go

.PHONY: build
build:
	go build -o $(CWD)/bin/app $(CWD)/cmd/main.go

.PHONY: clean
clean:
	rm -rf $(CWD)/bin/*	
