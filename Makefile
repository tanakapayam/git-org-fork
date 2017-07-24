# vars
SHELL     := /bin/bash
SOURCEDIR := .
SOURCES   := $(shell find $(SOURCEDIR) -type f -name '*.go')
BINARY    := $(shell basename `pwd`)
VERSION   := $(shell cat VERSION.txt)
LDFLAGS   :=-ldflags "-X main.VERSION=${VERSION}"

# targets
.DEFAULT_GOAL := run

lint build run install::
	@source style.bash
	@printf "$${BOLD}# LINT$${RESET}\n\n"
	@printf "$${BOLD}## FMT$${RESET}\n"
	for f in $(SOURCES); do \
		go fmt $$f; \
	done
	@echo

build test run install::
	@printf "$${BOLD}# BUILD$${RESET}\n"
	go build ${LDFLAGS} ${SOURCEDIR}
	@echo

run::
	@printf "$${BOLD}# RUN$${RESET}\n"
	clear
	@./${BINARY}
	@echo

install::
	@printf "$${BOLD}# INSTALL$${RESET}\n"
	go install ${LDFLAGS} ${SOURCEDIR}
	@echo

install clean::
	@printf "$${BOLD}# CLEAN$${RESET}\n"
	rm -f ${BINARY}
	@echo

.PHONY: lint build run install clean
