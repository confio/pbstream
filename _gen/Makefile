.PHONY: gen compile tools protoc-Windows protoc-Linux protoc-Darwin

gen: compile
	@go run ./cmd/gen.go ../testdata

compile:
	@protoc --gogo_out=. *.proto

### cross-platform check for installing protoc ###

ifeq ($(OS),Windows_NT)
    MYOS := Windows
else
    MYOS := $(shell uname -s)
endif

ifeq ($(MYOS),Windows)
	ZIP := protoc-3.4.0-win32.zip
endif
ifeq ($(MYOS),Darwin)  # Mac OS X
	ZIP := protoc-3.4.0-osx-x86_64.zip
endif
ifeq ($(MYOS),Linux)
	ZIP := protoc-3.4.0-linux-x86_64.zip
endif

##### scripts to install protoc and tooling on various platforms #####

protoc-Windows:
	@echo "Windows not yet supported"
	@false

protoc-Linux: /usr/local/bin/protoc

protoc-Darwin: /usr/local/bin/protoc

/usr/local/bin/protoc:
	@ curl -L https://github.com/google/protobuf/releases/download/v3.4.0/$(ZIP) > $(ZIP)
	@ unzip -q $(ZIP) -d protoc3
	@ rm $(ZIP)
	sudo mv protoc3/bin/protoc /usr/local/bin/
	@ sudo mv protoc3/include/* /usr/local/include/
	@ sudo chown `whoami` /usr/local/bin/protoc
	@ sudo chown -R `whoami` /usr/local/include/google
	@ rm -rf protoc3

tools: protoc-$(MYOS)
	@go get github.com/gogo/protobuf/proto
	@go get github.com/gogo/protobuf/gogoproto
