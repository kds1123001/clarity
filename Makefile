.PHONY: build test fmt vet install clean

build:
	go build -o clarity .

test:
	go test ./...

fmt:
	gofmt -l -w .

vet:
	go vet ./...

install:
	go install .

clean:
	rm -f clarity
