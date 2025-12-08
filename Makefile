GOOS := $(shell go env GOOS)
LIBNAME := libgowalletsdk
BUILD_DIR := build

STATICLIBEXT := .a
ifeq ($(GOOS),darwin)
  SHAREDLIBEXT := .dylib
else
  SHAREDLIBEXT := .so
endif

.PHONY: shared-library static-library clean check-go

check-go:
	@if ! command -v go &> /dev/null; then \
		echo "Go is not installed or not in PATH."; \
		exit 1; \
	fi

shared-library: check-go
	mkdir -p $(BUILD_DIR)
	go build -buildmode=c-shared -o $(BUILD_DIR)/$(LIBNAME)$(SHAREDLIBEXT) ./clib
	@echo "Built $(BUILD_DIR)/$(LIBNAME)$(SHAREDLIBEXT) and header $(BUILD_DIR)/$(LIBNAME).h"

static-library: check-go
	mkdir -p $(BUILD_DIR)
	go build -buildmode=c-archive -o $(BUILD_DIR)/$(LIBNAME)$(STATICLIBEXT) ./clib
	@echo "Built $(BUILD_DIR)/$(LIBNAME)$(STATICLIBEXT) and header $(BUILD_DIR)/$(LIBNAME).h"

clean:
	rm -f $(BUILD_DIR)/$(LIBNAME).so $(BUILD_DIR)/$(LIBNAME).dylib $(BUILD_DIR)/$(LIBNAME).a $(BUILD_DIR)/$(LIBNAME).h

