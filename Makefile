.PHONY: build run clean

build:
    go build -o bin/docufiller-update-server.exe .

run:
    go run main.go

clean:
    rm -rf bin/ data/ logs/

install-deps:
    go mod tidy
    go mod download
