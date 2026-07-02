IMAGE   ?= saw
TAG     ?= dev
PORT    ?= 8080
BINARY  := saw

.PHONY: build run test vet clean docker-build docker-run

build:
	go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-run:
	docker run --rm -p $(PORT):8080 $(IMAGE):$(TAG)
