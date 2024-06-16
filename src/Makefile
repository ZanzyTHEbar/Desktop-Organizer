# Go build flags
BUILD_FLAGS := -v

# Binary name
BINARY_NAME = hbat-blmp-api-
#BINARY_NAME := $(BINARY_NAME)$(shell date +'%Y%m%d%H%M%S')
BINARY_PATH := bin

# Detect the operating system
ifeq ($(OS),Windows_NT)
BINARY_NAME := $(BINARY_NAME)windows.exe
BINARY_PATH := \$(BINARY_NAME)
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
BINARY_PATH := /$(BINARY_NAME)
endif


MODULE_PATH := cmd/api/main.go

# Default target
.DEFAULT_GOAL := build

all: build

build:
	@echo "Building $(BINARY_NAME) ..."
	@go build $(BUILD_FLAGS) -o $(BINARY_PATH) $(MODULE_PATH)

clean-all:
	@echo "Cleaning All..."
	@rm -f $(BINARY_PATH)
	@go clean -cache -modcache -i -r

clean:
	@echo "Cleaning Binary..."
	@rm -f $(BINARY_PATH)

run:
	@echo "Running..."
	@go run $(MODULE_PATH)

# Create DB container
docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

# Live Reload
watch:
	@if command -v air > /dev/null; then \
	    air --root "." --testdata_dir "testdata" --tmp_dir "tmp" --build.cmd "make build" --build.bin "$(BINARY_PATH)" --build.exclude_dir "templates,build,assets,tmp,.tmp,vendor,testdata,.git,test-proj,supabase" --build.exclude_regex "_test.go" --build.include_ext "go,tpl,tmpl,html" --build.log "build-errors.log" --color.build "yellow" --color.main "magenta" --color.runner "green" --color.watcher "cyan" --screen.keep_scroll "true" --screen.clear_on_rebuild "true" --misc.clean_on_exit "false"; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/cosmtrek/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

.PHONY: all build run test clean

# TODO: Database Helpers

#up:

#down:

#drop:

#migrations:

#seed:
	
