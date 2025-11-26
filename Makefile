GOOS := $(shell go env GOOS)
LIBNAME := libgowalletsdk
BUILD_DIR := build

ifeq ($(GOOS),darwin)
  LIBEXT := .dylib
else
  LIBEXT := .so
endif

.PHONY: shared-library clean check-go

check-go:
	@if ! command -v go &> /dev/null; then \
		echo "Go is not installed or not in PATH."; \
		exit 1; \
	fi

shared-library: check-go
	mkdir -p $(BUILD_DIR)
	go build -buildmode=c-shared -o $(BUILD_DIR)/$(LIBNAME)$(LIBEXT) ./cshared
	@echo "Built $(BUILD_DIR)/$(LIBNAME)$(LIBEXT) and header $(BUILD_DIR)/$(LIBNAME).h"

clean:
	rm -f $(BUILD_DIR)/$(LIBNAME).so $(BUILD_DIR)/$(LIBNAME).dylib $(BUILD_DIR)/$(LIBNAME).h

