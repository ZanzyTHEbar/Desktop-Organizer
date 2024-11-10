# Go build flags
BUILD_FLAGS := -v

# Binary name
BINARY_NAME = desktop-cleaner-
#BINARY_NAME := $(BINARY_NAME)$(shell date +'%Y%m%d%H%M%S')-
BINARY_PATH_PREFIX := ./bin

# Detect the operating system
ifeq ($(OS),Windows_NT)
BINARY_NAME := $(BINARY_NAME)windows.exe
BINARY_PATH := $(BINARY_PATH_PREFIX)$(BINARY_PATH)\$(BINARY_NAME)
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
	BINARY_NAME := $(BINARY_NAME)linux
    endif
    ifeq ($(UNAME_S),Darwin)
	BINARY_NAME := $(BINARY_NAME)darwin
    endif
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
	BINARY_NAME := $(BINARY_NAME)amd64
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
	BINARY_NAME := $(BINARY_NAME)386
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
	BINARY_NAME := $(BINARY_NAME)arm
    endif
BINARY_PATH := $(BINARY_PATH_PREFIX)$(BINARY_PATH)/$(BINARY_NAME)
endif

MODULE_PATH := cmd/app/desktop-cleaner.go

# Default target
.DEFAULT_GOAL := build

all: build

build:
	@echo "Building $(BINARY_NAME) ..."
	@echo "Moving binary to $(BINARY_PATH) ..."
	@mkdir -p $(BINARY_PATH)
	@CGO_ENABLED=1 go build $(BUILD_FLAGS) -o $(BINARY_PATH) $(MODULE_PATH)

clean-all:
	@echo "Cleaning All..."
	@rm -rf $(BINARY_PATH_PREFIX)
	@go clean -cache -modcache -i -r

clean:
	@echo "Cleaning Binary..."
	@echo "Removing $(BINARY_PATH) ..."
	@rm -rf $(BINARY_PATH_PREFIX)

run:
	@echo "Running..."
	@go run $(MODULE_PATH)

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

.PHONY: all build run test clean clean-all
