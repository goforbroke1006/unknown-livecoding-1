BINARY_NAME=./unknown-livecoding-1

.PHONY: all dep build test benchmark clean

all: dep build test

dep:
	go mod download

build:
	go build -o ${BINARY_NAME} ./cmd

test:
	go test ./...

benchmark:
	bash ./benchmark.sh

clean:
	rm -f ${BINARY_NAME}
	find ./ -type f -name "*.log" -exec rm -f '{}' \;
	find ./ -type f -name "*.test" -exec rm -f '{}' \;
