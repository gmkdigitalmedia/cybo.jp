FROM nvidia/cuda:12.3.1-devel-ubuntu22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    wget \
    git \
    build-essential \
    libwebp-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.21
RUN wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz && \
    rm go1.21.6.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build CUDA kernels
RUN nvcc -arch=sm_75 -O3 --use_fast_math -Xcompiler -fPIC \
    -I/usr/local/cuda/include \
    -c internal/tiler/tile_kernel.cu -o internal/tiler/tile_kernel.o && \
    nvcc -arch=sm_75 -shared internal/tiler/tile_kernel.o \
    -o internal/tiler/libtile_kernel.so \
    -L/usr/local/cuda/lib64 -lcuda -lcudart

# Build Go application
RUN CGO_ENABLED=1 go build -o /cyto-viewer \
    -ldflags="-s -w" \
    -tags=netgo \
    ./cmd/server

# Create data directories
RUN mkdir -p /data/slides /data/temp

# Expose port
EXPOSE 8080

# Set environment variables
ENV GPU_DEVICE_ID=0
ENV GPU_CACHE_SIZE=8192
ENV SERVER_PORT=8080
ENV STORAGE_PATH=/data/slides
ENV TEMP_PATH=/data/temp

# Run the application
CMD ["/cyto-viewer"]
