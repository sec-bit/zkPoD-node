GO ?= $(shell which go)
BIN := zkPoD-node
LOG := pod.log
PROJ_HOME=${PWD}

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

# test params
POD_CORE_DIR := $(PWD)/../zkPoD-lib/pod_core
CGO_LDFLAGS := -L$(POD_CORE_DIR)
KEYSTORE_FILE := ./keystore
KEYSTORE_PASSWORD := 123456
PORT := 1234
NETIP := localhost:4321
INIT_FILE := ./testdata/init_plain.json
MERKLE_ROOT := a5135a3a7806f2434c384ad531a3c3e94b206d409ef3ad90ccc332cbae5cea38 
ETH_VALUE := 200
ETH_ADDR := 0x4504fccc67Dca1e1e77Ff3DA31bfc30c03414B2C
CONFIG_FILE := config.json
LIBRARY_PATH := $(POD_CORE_DIR)

ifeq ($(OS),Windows_NT)
	OS_TYPE := Windows
else
	UNAME := $(shell uname -s)
	ifeq ($(UNAME),Linux)
		OS_TYPE := Linux
	else ifeq ($(UNAME),Darwin)
		OS_TYPE := Darwin
	else
		OS_TYPE := Unknown
	endif
endif

ifeq ($(OS_TYPE), Darwin)
	LIBRARY_PATH := DYLD_LIBRARY_PATH=$(POD_CORE_DIR)
else
	LIBRARY_PATH := LD_LIBRARY_PATH=$(POD_CORE_DIR)
endif

all:
	CGO_LDFLAGS=$(CGO_LDFLAGS) \
	$(GO) build -o $(BIN)

POD_NODE_BIN := $(LIBRARY_PATH) $(PROJ_HOME)/$(BIN)

run:
	$(POD_NODE_BIN) -o start \
	-k $(KEYSTORE_FILE) -pass $(KEYSTORE_PASSWORD) \
	-port $(PORT) -ip $(NETIP)

run-initdata:
	$(POD_NODE_BIN) -o initdata \
	-init $(INIT_FILE)

run-publish:
	$(POD_NODE_BIN) -o publish \
	-mkl $(MERKLE_ROOT) -eth $(ETH_VALUE)
	
run-close:
	$(POD_NODE_BIN) -o close \
	-mkl $(MERKLE_ROOT)
	
run-withdrawA:
	$(POD_NODE_BIN) -o withdraw \
	-mkl $(MERKLE_ROOT)

run-deposit:
	$(POD_NODE_BIN) -o deposit \
	-eth $(ETH_VALUE) -addr $(ETH_ADDR)

run-undeposit:
	$(POD_NODE_BIN) -o undeposit \
	-addr $(ETH_ADDR)

run-withdrawB:
	$(POD_NODE_BIN) -o withdraw \
	-addr $(ETH_ADDR)

run-purchase:
	$(POD_NODE_BIN) -o purchase \
	-c $(CONFIG_FILE)

clean:
	@rm -f $(BIN)