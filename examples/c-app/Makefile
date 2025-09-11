CC := cc
ROOT := ../..
BUILD_DIR := $(ROOT)/build
LIB := $(BUILD_DIR)/libgowalletsdk
BIN_DIR := bin

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  LIBEXT := .dylib
  RPATH := -Wl,-rpath,@executable_path/../bin
else
  LIBEXT := .so
  RPATH := -Wl,-rpath,$(BIN_DIR)
endif

LIBFILE = $(LIB)$(LIBEXT)

.PHONY: run build clean

build:
	mkdir -p $(BIN_DIR)
	$(MAKE) -C $(ROOT) build-c-lib
	cp -f $(LIBFILE) $(BIN_DIR)/
	$(CC) -I$(BUILD_DIR) -L$(BIN_DIR) -o $(BIN_DIR)/c-app main.c -lgowalletsdk $(RPATH)

run: build
	$(BIN_DIR)/c-app

clean:
	rm -rf $(BIN_DIR)
