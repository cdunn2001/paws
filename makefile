# from https://qvault.io/golang/golang-project-structure/
#export GOPATH=$(pwd)
#export GOPATH=~/gh/GO
#vpath %.go cmd

quick:
	go test ./pkg/stuff
all: test vet fmt lint build

test:
	go test ./pkg/...

vet:
	go vet ./pkg/...

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs go fmt
	#test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	#go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status
	golint --set_exit_status cmd/...
	golint --set_exit_status pkg/...
build: bin/paws

# hello, try, paws, etc. (for now)
bin/%: .FORCE
	go build -o $@ ./cmd/$*
serve: bin/paws
	./$<
.FORCE:
