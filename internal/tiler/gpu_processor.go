package tiler

import (
	"context"
	"fmt"
	"image"
	"sync"
	"unsafe"

	"cyto-viewer/internal/config"
)

/*
#cgo LDFLAGS: -L/usr/local/cuda/lib64 -lcuda -lcudart
#cgo CFLAGS: -I/usr/local/cuda/include

#include <cuda_runtime.h>
#include <cuda.h>

// CUDA kernel for fast image decompression and color correction
extern void processTile(unsigned char* input, unsigned char* output, 
                       int width, int height, float* colorMatrix);
*/
import "C"

type GPUTileProcessor struct {
	config       *config.GPUConfig
	deviceID     int
	stream       C.cudaStream_t
	bufferPool   sync.Pool
	tileCache    *TileCache
	colorCorrect bool
	mu           sync.RWMutex
}

type TileRequest struct {
	SlideID  string
	Layer    int
	X        int
	Y        int
	Z        int // Focus layer
	Width    int
	Height   int
	Format   string // "jpeg", "webp", "avif"
	Quality  int
}

type TileResponse struct {
	Data        []byte
	Width       int
	Height      int
	ContentType string
	CacheKey    string
}

func NewGPUTileProcessor(cfg *config.GPUConfig) (*GPUTileProcessor, error) {
	// Initialize CUDA
	var deviceCount C.int
	if err := C.cudaGetDeviceCount(&deviceCount); err != C.cudaSuccess {
		return nil, fmt.Errorf("failed to get CUDA device count: %v", err)
	}

	if deviceCount == 0 {
		return nil, fmt.Errorf("no CUDA devices found")
	}

	// Set device
	if err := C.cudaSetDevice(C.int(cfg.DeviceID)); err != C.cudaSuccess {
		return nil, fmt.Errorf("failed to set CUDA device: %v", err)
	}

	// Create CUDA stream for async operations
	var stream C.cudaStream_t
	if err := C.cudaStreamCreate(&stream); err != C.cudaSuccess {
		return nil, fmt.Errorf("failed to create CUDA stream: %v", err)
	}

	processor := &GPUTileProcessor{
		config:       cfg,
		deviceID:     cfg.DeviceID,
		stream:       stream,
		colorCorrect: cfg.ColorCorrection,
		tileCache:    NewTileCache(cfg.CacheSize),
	}

	// Initialize buffer pool for zero-copy operations
	processor.bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 4096*4096*4) // Max tile size
		},
	}

	return processor, nil
}

func (p *GPUTileProcessor) ProcessTile(ctx context.Context, req *TileRequest) (*TileResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%d:%d:%d:%d", req.SlideID, req.Layer, req.X, req.Y, req.Z)
	
	if cached, ok := p.tileCache.Get(cacheKey); ok {
		return cached, nil
	}

	// Load raw tile data from storage
	rawData, err := p.loadRawTile(req)
	if err != nil {
		return nil, fmt.Errorf("failed to load raw tile: %w", err)
	}

	// Allocate GPU memory
	var dInput, dOutput unsafe.Pointer
	tileSize := req.Width * req.Height * 4 // RGBA

	if err := C.cudaMalloc(&dInput, C.size_t(len(rawData))); err != C.cudaSuccess {
		return nil, fmt.Errorf("failed to allocate GPU input memory: %v", err)
	}
	defer C.cudaFree(dInput)

	if err := C.cudaMalloc(&dOutput, C.size_t(tileSize)); err != C.cudaSuccess {
		C.cudaFree(dInput)
		return nil, fmt.Errorf("failed to allocate GPU output memory: %v", err)
	}
	defer C.cudaFree(dOutput)

	// Copy input to GPU
	if err := C.cudaMemcpyAsync(dInput, unsafe.Pointer(&rawData[0]), 
		C.size_t(len(rawData)), C.cudaMemcpyHostToDevice, p.stream); err != C.cudaSuccess {
		return nil, fmt.Errorf("failed to copy to GPU: %v", err)
	}

	// Prepare color correction matrix if enabled
	var colorMatrix *C.float
	if p.colorCorrect {
		matrix := p.getColorCorrectionMatrix()
		colorMatrix = (*C.float)(unsafe.Pointer(&matrix[0]))
	}

	// Execute GPU kernel for decompression and processing
	C.processTile((*C.uchar)(dInput), (*C.uchar)(dOutput), 
		C.int(req.Width), C.int(req.Height), colorMatrix)

	// Synchronize stream
	if err := C.cudaStreamSynchronize(p.stream); err != C.cudaSuccess {
		return nil, fmt.Errorf("CUDA stream sync failed: %v", err)
	}

	// Copy result back to host
	output := p.bufferPool.Get().([]byte)[:tileSize]
	if err := C.cudaMemcpy(unsafe.Pointer(&output[0]), dOutput, 
		C.size_t(tileSize), C.cudaMemcpyDeviceToHost); err != C.cudaSuccess {
		return nil, fmt.Errorf("failed to copy from GPU: %v", err)
	}

	// Encode to requested format (JPEG, WebP, or AVIF)
	encoded, contentType, err := p.encodeTile(output, req)
	if err != nil {
		p.bufferPool.Put(output)
		return nil, fmt.Errorf("failed to encode tile: %w", err)
	}

	response := &TileResponse{
		Data:        encoded,
		Width:       req.Width,
		Height:      req.Height,
		ContentType: contentType,
		CacheKey:    cacheKey,
	}

	// Cache the result
	p.tileCache.Set(cacheKey, response)

	// Return buffer to pool
	p.bufferPool.Put(output)

	return response, nil
}

func (p *GPUTileProcessor) loadRawTile(req *TileRequest) ([]byte, error) {
	// This would interface with your actual storage backend
	// For now, placeholder for the raw tile loading logic
	// In production, this would read from your compressed tile storage
	return nil, fmt.Errorf("not implemented: loadRawTile")
}

func (p *GPUTileProcessor) encodeTile(data []byte, req *TileRequest) ([]byte, string, error) {
	// Create image from raw RGBA data
	img := &image.RGBA{
		Pix:    data,
		Stride: req.Width * 4,
		Rect:   image.Rect(0, 0, req.Width, req.Height),
	}

	// Use optimized encoders based on format
	switch req.Format {
	case "webp":
		return encodeWebP(img, req.Quality)
	case "avif":
		return encodeAVIF(img, req.Quality)
	default:
		return encodeJPEG(img, req.Quality)
	}
}

func (p *GPUTileProcessor) getColorCorrectionMatrix() []float32 {
	// Color correction matrix for medical imaging
	// These values should be calibrated for your specific scanner
	return []float32{
		1.05, 0.0, 0.0, 0.0,  // R
		0.0, 1.02, 0.0, 0.0,  // G
		0.0, 0.0, 1.08, 0.0,  // B
		0.0, 0.0, 0.0, 1.0,   // A
	}
}

func (p *GPUTileProcessor) ProcessBatch(ctx context.Context, requests []*TileRequest) ([]*TileResponse, error) {
	// Batch processing for multiple tiles - much faster than one-by-one
	responses := make([]*TileResponse, len(requests))
	errChan := make(chan error, len(requests))
	
	// Use worker pool
	workers := 4 // Adjust based on GPU capability
	requestChan := make(chan struct {
		idx int
		req *TileRequest
	}, len(requests))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range requestChan {
				resp, err := p.ProcessTile(ctx, work.req)
				if err != nil {
					errChan <- err
					return
				}
				responses[work.idx] = resp
			}
		}()
	}

	// Send requests
	for i, req := range requests {
		requestChan <- struct {
			idx int
			req *TileRequest
		}{i, req}
	}
	close(requestChan)

	// Wait for completion
	wg.Wait()
	close(errChan)

	// Check for errors
	if err := <-errChan; err != nil {
		return nil, err
	}

	return responses, nil
}

func (p *GPUTileProcessor) Close() error {
	if p.stream != nil {
		C.cudaStreamDestroy(p.stream)
	}
	return nil
}
