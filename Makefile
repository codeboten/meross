default:
	cd api && go build

deps:
	dep ensure

test:
	cd api && go test

build-client:
	cd cmd && go build -o meross-client

run-client: build-client
	./cmd/meross-client
