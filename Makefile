.PHONY: build clean run test cuda docker install-deps

# Build configuration
BINARY_NAME=cyto-viewer
CUDA_PATH=/usr/local/cuda
GO=go
NVCC=nvcc

# CUDA compilation flags
CUDA_FLAGS=-arch=sm_75 -O3 --use_fast_math -Xcompiler -fPIC
CUDA_INCLUDES=-I$(CUDA_PATH)/include
CUDA_LIBS=-L$(CUDA_PATH)/lib64 -lcuda -lcudart

all: cuda build

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	$(GO) mod download
	@echo "Installing system dependencies..."
	sudo apt-get update
	sudo apt-get install -y libwebp-dev cuda-toolkit-12-3

# Build CUDA kernels
cuda:
	@echo "Compiling CUDA kernels..."
	$(NVCC) $(CUDA_FLAGS) $(CUDA_INCLUDES) -c internal/tiler/tile_kernel.cu -o internal/tiler/tile_kernel.o
	$(NVCC) $(CUDA_FLAGS) -shared internal/tiler/tile_kernel.o -o internal/tiler/libtile_kernel.so $(CUDA_LIBS)

# Build Go application
build: cuda
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=1 $(GO) build -o bin/$(BINARY_NAME) \
		-ldflags="-s -w" \
		-tags=netgo \
		./cmd/server

# Build with debugging symbols
build-debug: cuda
	@echo "Building $(BINARY_NAME) with debug symbols..."
	CGO_ENABLED=1 $(GO) build -o bin/$(BINARY_NAME) \
		-gcflags="all=-N -l" \
		./cmd/server

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

# Run tests
test:
	$(GO) test -v ./...

# Run benchmarks
bench:
	$(GO) test -bench=. -benchmem ./internal/tiler/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f internal/tiler/*.o
	rm -f internal/tiler/*.so
	$(GO) clean

# Build Docker image
docker:
	docker build -t cyto-viewer:latest .

# Development server with hot reload
dev:
	air -c .air.toml

# Generate password hash for authentication
gen-password:
	@read -sp "Enter password: " password; \
	echo ""; \
	$(GO) run scripts/gen_password.go $$password

# Install as systemd service
install-service: build
	@echo "Installing systemd service..."
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	sudo cp scripts/cyto-viewer.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable cyto-viewer
	sudo systemctl start cyto-viewer

# Show GPU info
gpu-info:
	nvidia-smi
	$(NVCC) --version
