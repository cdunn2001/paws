GIT_COMMIT=$(shell git rev-parse HEAD)

# from https://qvault.io/golang/golang-project-structure/
#export GOPATH=$(pwd)
#export GOPATH=~/gh/GO
#vpath %.go cmd
#export PAWS_TEST_LOG=1
DUMMYDIR:=${PWD}/pkg/web/testdata
export PATH:=${DUMMYDIR}:${PATH}
#NOTIFY_SOCKET=/tmp/kpaws.sock
WATCHDOG_USEC=4000000
#export NOTIFY_SOCKET
export WATCHDOG_USEC

echo:
	echo "PATH=${PATH}"
quick:
	go test ./pkg/stuff
all: test vet fmt lint build

test:
	go test ./pkg/... -v -timeout 30s -short -failfast
clean: # Use this if you alter bash scripts or data.
	go clean -testcache ./pkg/web/...
vet:
	go vet ./pkg/...

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs go fmt
	#test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	#go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status
	golint --set_exit_status cmd/...
	golint --set_exit_status pkg/...
build: bin/pawsgo

# hello, try, paws, etc. (for now)
bin/%: .FORCE
	go build -o $@ ./cmd/$*
serve: bin/pawsgo
	${MAKE} local
local:
	bin/pawsgo --data-dir tmp --logoutput pa-wsgo.log #--config SNAFU.json
# this target runs paws slower, to be used by end to end python testing
slowlocal: bin/pawsgo
	STATUS_COUNT=5 STATUS_DELAY_SECONDS=1 ./bin/pawsgo --console

release:
	go build -ldflags "-X pacb.com/seq/paws/pkg/config.Version=${VERSION}-${GIT_COMMIT}" -o bin/pawsgo ./cmd/pawsgo

.FORCE:
.PHONY: test
