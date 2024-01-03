PROJECTNAME=$(shell basename "$(PWD)")
VERSION=-ldflags="-X main.Version=$(shell git describe --tags)"


.PHONY: help run build install license
all: help

get:
	@echo "  >  \033[32mDownloading & Installing all the modules...\033[0m "
	go mod tidy && go mod download

build:
	@echo "  >  \033[32mBuilding btc_layer2_committer...\033[0m "
	go build -o ./build/btc_layer2_committer

install:
	@echo "  >  \033[32mInstalling btc_layer2_committer...\033[0m "
	cd build && go install $(VERSION)