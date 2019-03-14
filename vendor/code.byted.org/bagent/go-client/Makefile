thrift = thrift
ifeq (in17-164, $(shell hostname))
	thrift = /opt/tiger/thrift/bin/thrift
endif
ifeq (n6-131-078, $(shell hostname))
	thrift = /opt/tiger/thrift/bin/thrift
endif

all: generate

generate:
	mkdir -p thrift_gen
	thrift -r --gen go:thrift_import=code.byted.org/gopkg/thrift,package_prefix=code.byted.org/bagent/go-client/thrift_gen/ -out thrift_gen ../idl/agent.thrift

clean:
	rm -rf thrift_gen

