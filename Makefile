FIRST_GOPATH              := $(firstword $(subst :, ,$(GOPATH)))
PKGS                      := $(shell go list ./... | grep -v /tests | grep -v /xcpb | grep -v /gpb)
GOFILES_NOVENDOR          := $(shell find . -name vendor -prune -o -type f -name '*.go' -not -name '*.pb.go' -print)
GOFILES_BUILD             := $(shell find . -type f -name '*.go' -not -name '*_test.go')
PROTOFILES                := $(shell find . -name vendor -prune -o -type f -name '*.proto' -print)
GOPASS_VERSION            ?= $(shell cat VERSION)
GOPASS_OUTPUT             ?= gopass
GOPASS_REVISION           := $(shell cat COMMIT 2>/dev/null || git rev-parse --short=8 HEAD)
BASH_COMPLETION_OUTPUT    := bash.completion
FISH_COMPLETION_OUTPUT    := fish.completion
ZSH_COMPLETION_OUTPUT     := zsh.completion
# Support reproducible builds by embedding date according to SOURCE_DATE_EPOCH if present
DATE                      := $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" '+%FT%T%z' 2>/dev/null || date -u '+%FT%T%z')
BUILDFLAGS_NOPIE          := -tags=netgo -trimpath -ldflags="-s -w -X main.version=$(GOPASS_VERSION) -X main.commit=$(GOPASS_REVISION) -X main.date=$(DATE)" -gcflags="-trimpath=$(GOPATH)" -asmflags="-trimpath=$(GOPATH)"
BUILDFLAGS                ?= $(BUILDFLAGS_NOPIE) -buildmode=pie
TESTFLAGS                 ?=
PWD                       := $(shell pwd)
PREFIX                    ?= $(GOPATH)
BINDIR                    ?= $(PREFIX)/bin
GO                        := GO111MODULE=on CGO_ENABLED=0 go
GOOS                      ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH                    ?= $(shell go version | cut -d' ' -f4 | cut -d'/' -f2)
TAGS                      ?= netgo
export GO111MODULE=on

OK := $(shell tput setaf 6; echo ' [OK]'; tput sgr0;)

all: build completion man
build: $(GOPASS_OUTPUT)
completion: $(BASH_COMPLETION_OUTPUT) $(FISH_COMPLETION_OUTPUT) $(ZSH_COMPLETION_OUTPUT)
travis: sysinfo crosscompile build fulltest codequality completion full
travis-osx: sysinfo build test completion full
travis-windows: sysinfo build test-win completion

sysinfo:
	@echo ">> SYSTEM INFORMATION"
	@echo -n "     PLATFORM: $(shell uname -a)"
	@printf '%s\n' '$(OK)'
	@echo -n "     PWD:    : $(shell pwd)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GO      : $(shell go version)"
	@printf '%s\n' '$(OK)'
	@echo -n "     BUILDFLAGS: $(BUILDFLAGS)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GIT     : $(shell git version)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPG1    : $(shell which gpg) $(shell gpg --version | head -1)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPG2    : $(shell which gpg2) $(shell gpg2 --version | head -1)"
	@printf '%s\n' '$(OK)'
	@echo -n "     GPG-Agent    : $(shell which gpg-agent) $(shell gpg-agent --version | head -1)"
	@printf '%s\n' '$(OK)'

clean:
	@echo -n ">> CLEAN"
	@$(GO) clean -i ./...
	@rm -f ./coverage-all.html
	@rm -f ./coverage-all.out
	@rm -f ./coverage.out
	@find . -type f -name "coverage.out" -delete
	@rm -f gopass_*.deb
	@rm -f gopass-*.pkg.tar.xz
	@rm -f gopass-*.rpm
	@rm -f gopass-*.tar.bz2
	@rm -f gopass-*.tar.gz
	@rm -f gopass-*-*
	@rm -f tests/tests
	@rm -f *.test
	@rm -rf dist/*
	@rm -f *.completion
	@printf '%s\n' '$(OK)'

$(GOPASS_OUTPUT): $(GOFILES_BUILD)
	@echo -n ">> BUILD, version = $(GOPASS_VERSION)/$(GOPASS_REVISION), output = $@"
	@$(GO) build -o $@ $(BUILDFLAGS)
	@printf '%s\n' '$(OK)'

install: all install-completion install-man
	@echo -n ">> INSTALL, version = $(GOPASS_VERSION)"
	@install -m 0755 -d $(DESTDIR)$(BINDIR)
	@install -m 0755 $(GOPASS_OUTPUT) $(DESTDIR)$(BINDIR)/gopass
	@printf '%s\n' '$(OK)'

fulltest: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"full-mode\": race detector off, build tags: xc"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		go test -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -coverprofile=coverage.out -covermode=atomic $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html

fulltest-nocover: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"full-mode-no-coverage\": race detector off, build tags: xc"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		go test -run '(Test|Example)' $(TESTFLAGS) $(pkg) || exit 1;)

racetest: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"full-mode\": race detector on"
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		go test -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) -race -coverprofile=coverage.out -covermode=atomic $(pkg) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)
	@$(GO) tool cover -html=coverage-all.out -o coverage-all.html

test: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"fast-mode\": race detector off"
	@$(foreach pkg, $(PKGS),\
	    echo -n "     ";\
		$(GO) test -test.short -run '(Test|Example)' $(BUILDFLAGS) $(TESTFLAGS) $(pkg) || exit 1;)

test-win: $(GOPASS_OUTPUT)
	@echo ">> TEST, \"fast-mode-win\": race detector off"
	@$(foreach pkg, $(PKGS),\
		$(GO) test -test.short -run '(Test|Example)' $(pkg) || exit 1;)

test-integration: $(GOPASS_OUTPUT)
	cd tests && GOPASS_BINARY=$(PWD)/$(GOPASS_OUTPUT) GOPASS_TEST_DIR=$(PWD)/tests go test -v

crosscompile:
	@echo -n ">> CROSSCOMPILE linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-linux-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-darwin-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(GOPASS_OUTPUT)-windows-amd64
	@printf '%s\n' '$(OK)'

full:
	@echo -n ">> COMPILE linux/amd64 xc"
	$(GO) build -o $(GOPASS_OUTPUT)-full

%.completion: $(GOPASS_OUTPUT)
	@printf ">> $* completion, output = $@"
	@./gopass completion $* > $@
	@printf "%s\n" "$(OK)"

install-completion: completion
	@install -d $(DESTDIR)$(PREFIX)/share/zsh/site-functions $(DESTDIR)$(PREFIX)/share/bash-completion/completions $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d
	@install -m 0755 $(ZSH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_gopass
	@install -m 0755 $(BASH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/bash-completion/completions/gopass
	@install -m 0755 $(FISH_COMPLETION_OUTPUT) $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d/gopass.fish
	@printf '%s\n' '$(OK)'

codequality:
	@echo ">> CODE QUALITY"

	@echo -n "     REVIVE    "
	@which revive > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mgechev/revive; \
	fi
	@revive -formatter friendly -exclude vendor/... ./...
	@printf '%s\n' '$(OK)'

	@echo -n "     FMT       "
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			out=$$(gofmt -s -l -d -e $(gofile) | tee /dev/stderr); if [ -n "$$out" ]; then exit 1; fi;)
	@printf '%s\n' '$(OK)'

	@echo -n "     CLANGFMT  "
	@$(foreach pbfile, $(PROTOFILES),\
			if [ $$(clang-format -output-replacements-xml $(pbfile) | wc -l) -gt 3  ]; then exit 1; fi;)
	@printf '%s\n' '$(OK)'

	@echo -n "     VET       "
	@$(GO) vet ./...
	@printf '%s\n' '$(OK)'

	@echo -n "     CYCLO     "
	@which gocyclo > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/fzipp/gocyclo/cmd/gocyclo; \
	fi
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			gocyclo -over 22 $(gofile) || exit 1;)
	@printf '%s\n' '$(OK)'

	@echo -n "     LINT      "
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u golang.org/x/lint/golint; \
	fi
	@$(foreach pkg, $(PKGS),\
			golint -set_exit_status $(pkg) || exit 1;)
	@printf '%s\n' '$(OK)'

	@echo -n "     INEFF     "
	@which ineffassign > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/gordonklaus/ineffassign; \
	fi
	@ineffassign . || exit 1
	@printf '%s\n' '$(OK)'

	@echo -n "     SPELL     "
	@which misspell > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	@$(foreach gofile, $(GOFILES_NOVENDOR),\
			misspell --error $(gofile) || exit 1;)
	@printf '%s\n' '$(OK)'

	@echo -n "     STATICCHECK "
	@which staticcheck > /dev/null; if [ $$? -ne 0  ]; then \
		$(GO) get -u honnef.co/go/tools/cmd/staticcheck; \
	fi
	@staticcheck $(PKGS) || exit 1
	@printf '%s\n' '$(OK)'

	@echo -n "     UNPARAM "
	@which unparam > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) get -u mvdan.cc/unparam; \
	fi
	@unparam -exported=false $(PKGS)
	@printf '%s\n' '$(OK)'

gen:
	@go generate ./...

fmt:
	@gofmt -s -l -w $(GOFILES_NOVENDOR)
	@goimports -l -w $(GOFILES_NOVENDOR)
	@which clang-format > /dev/null; if [ $$? -eq 0 ]; then \
		clang-format -i $(PROTOFILES); \
	fi
	@go mod tidy

fuzz-gpg:
	mkdir -p workdir/gpg-cli/corpus
	go-fuzz-build github.com/gopasspw/gopass/backend/gpg/cli
	go-fuzz -bin=cli-fuzz.zip -workdir=workdir/gpg-cli

check-release-env:
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is undefined)
endif

release: goreleaser

goreleaser: check-release-env travis clean
	@echo ">> RELEASE, goreleaser"
	@goreleaser

deps:
	@go build -v ./...

upgrade: gen fmt
	@go get -u

man:
	@go run helpers/man/main.go > gopass.1

install-man: man
	@install -D -m 0644 gopass.1 $(DESTDIR)$(PREFIX)/share/man/man1/gopass.1

.PHONY: clean build completion install sysinfo crosscompile test codequality release goreleaser debsign man
