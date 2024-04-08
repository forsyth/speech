SHELL=/bin/rc
all:V:
	go build ./...

fmt:V:
	go fmt /...

test:V:
	. ./secrets.rc
	go test -v ./polly

testcov:V:
	go test -v -coverprofile=c.out ./...

vet:V:
	go vet .
