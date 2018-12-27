default:
	cd api && go build

deps:
	dep ensure

test:
	cd api && go test

cover:
	cd api && go test -cover

build-client:
	cd cmd && go build -o meross-client

run-client: build-client
	./cmd/meross-client
