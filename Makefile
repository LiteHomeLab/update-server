.PHONY: build run clean build-all install-deps

build:
    go build -o bin/docufiller-update-server.exe .

run:
    go run main.go

clean:
    rm -rf bin/ data/ logs/

install-deps:
    go mod tidy
    go mod download

# Build all components (server + clients)
build-all:
	@echo "========================================"
	@echo "  Update Server All-in-One Builder"
	@echo "========================================"
	@echo ""
	@echo "[1/4] Creating output directories..."
	@mkdir -p bin/clients
	@mkdir -p data/clients
	@echo "Created: bin/"
	@echo "Created: bin/clients/"
	@echo "Created: data/clients/"
	@echo ""
	@echo "[2/4] Building Update Server..."
	@go build -o bin/update-server .
	@echo "SUCCESS: Built update-server"
	@echo ""
	@echo "[3/4] Building Publish Client (update-admin)..."
	@cd clients/go/admin && go build -o ../../bin/clients/update-admin .
	@echo "SUCCESS: Built update-admin"
	@echo ""
	@echo "[4/4] Copying client executables to server data directory..."
	@cp bin/clients/update-admin data/clients/publish-client
	@echo "Copied: publish-client -> data/clients/"
	@echo ""
	@echo "========================================"
	@echo "  Build Completed Successfully!"
	@echo "========================================"
	@echo ""
	@echo "Output files:"
	@echo "  - Server: bin/update-server"
	@echo "  - Publish Client: bin/clients/update-admin"
	@echo ""
	@echo "Server deployment directory:"
	@echo "  - Clients: data/clients/"
	@echo ""
