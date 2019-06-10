GO ?= $(shell which go)
BIN := "zkPoD-node"
LOG := "pod.log"
PROJ_HOME=${PWD}

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

# test params
POD_CORE_DIR :=  $(GOPATH)/src/github.com/sec-bit/zkPoD-lib/pod_core
CGO_LDFLAGS := -L$(POD_CORE_DIR)
KEYSTORE_FILE := ./keystore
KEYSTORE_PASSWORD := 123456
PORT := 1234
NETIP := localhost:4321
INIT_FILE := ./testdata/init_plain.json
MERKLE_ROOT := a5135a3a7806f2434c384ad531a3c3e94b206d409ef3ad90ccc332cbae5cea38 
ETH_VALUE := 200
ETH_ADDR := 0x4eC1B88456547e3Fe169510D3FfE2EC7de757B6f
CONFIG_FILE := config.json
LIBRARY_PATH := $(POD_CORE_DIR)

all:
	CGO_LDFLAGS=$(CGO_LDFLAGS) \
	$(GO) build -o $(BIN)

run:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o start \
	-k $(KEYSTORE_FILE) -pass $(KEYSTORE_PASSWORD) \
	-port $(PORT) -ip $(NETIP)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o start \
	-k $(KEYSTORE_FILE) -pass $(KEYSTORE_PASSWORD) \
	-port $(PORT) -ip $(NETIP)
endif

run-initdata:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o initdata \
	-init $(INIT_FILE)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o initdata \
	-init $(INIT_FILE)
endif

run-publish:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o publish \
	-mkl $(MERKLE_ROOT) -eth $(ETH_VALUE)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o publish \
	-mkl $(MERKLE_ROOT) -eth $(ETH_VALUE)
endif
	
run-close:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o close \
	-mkl $(MERKLE_ROOT)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o close \
	-mkl $(MERKLE_ROOT)
endif
	
run-seller-withdraw:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o withdraw \
	-mkl $(MERKLE_ROOT)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o withdraw \
	-mkl $(MERKLE_ROOT)
endif

run-deposit:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o deposit \
	-eth $(ETH_VALUE) -addr $(ETH_ADDR)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o deposit \
	-eth $(ETH_VALUE) -addr $(ETH_ADDR)
endif

run-undeposit:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o undeposit \
	-addr $(ETH_ADDR)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o undeposit \
	-addr $(ETH_ADDR)
endif

run-withdraw:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o withdraw \
	-addr $(ETH_ADDR)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o withdraw \
	-addr $(ETH_ADDR)
endif

run-purchase:
ifeq ($(OS_TYPE), Darwin)
	DYLD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o purchase \
	-c $(CONFIG_FILE)
else
	LD_LIBRARY_PATH=$(POD_CORE_DIR) \
	$(PROJ_HOME)/$(BIN) -o purchase \
	-c $(CONFIG_FILE)
endif

clean:
	@rm -f $(BIN)