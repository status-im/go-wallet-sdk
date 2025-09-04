GOOS := $(shell go env GOOS)
LIBNAME := libgowalletsdk
BUILD_DIR := build

ifeq ($(GOOS),darwin)
  LIBEXT := .dylib
else
  LIBEXT := .so
endif

.PHONY: build-c-lib clean check-go

check-go:
	@current=$$(go version | awk '{print $$3}' | sed 's/go//'); \
	required=1.23; \
	if [ -z "$$current" ]; then \
		echo "Unable to detect Go version. Please install Go $$required or newer."; \
		exit 1; \
	fi; \
	# Compare versions using sort -V
	if [ $$(printf '%s\n' "$$required" "$$current" | sort -V | head -n1) != "$$required" ]; then \
		echo "Go $$required or newer is required. Found $$current"; \
		echo "Tip: brew install go (or ensure PATH uses a recent Go)"; \
		exit 1; \
	fi

build-c-lib: check-go
	mkdir -p $(BUILD_DIR)
	go build -buildmode=c-shared -o $(BUILD_DIR)/$(LIBNAME)$(LIBEXT) ./cshared
	@echo "Built $(BUILD_DIR)/$(LIBNAME)$(LIBEXT) and header $(BUILD_DIR)/$(LIBNAME).h"

clean:
	rm -f $(BUILD_DIR)/$(LIBNAME).so $(BUILD_DIR)/$(LIBNAME).dylib $(BUILD_DIR)/$(LIBNAME).h

