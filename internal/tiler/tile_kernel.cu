// tile_kernel.cu
#include <cuda_runtime.h>
#include <device_launch_parameters.h>

// Fast JPEG/video frame decompression on GPU using NVJPEG
__global__ void decompressTileKernel(const unsigned char* __restrict__ input,
                                      unsigned char* __restrict__ output,
                                      int width, int height,
                                      const float* __restrict__ colorMatrix) {
    int x = blockIdx.x * blockDim.x + threadIdx.x;
    int y = blockIdx.y * blockDim.y + threadIdx.y;
    
    if (x >= width || y >= height) return;
    
    int idx = (y * width + x) * 4;
    
    // Load pixel
    float r = input[idx];
    float g = input[idx + 1];
    float b = input[idx + 2];
    float a = input[idx + 3];
    
    // Apply color correction matrix
    if (colorMatrix != nullptr) {
        float nr = colorMatrix[0] * r + colorMatrix[1] * g + 
                   colorMatrix[2] * b + colorMatrix[3] * a;
        float ng = colorMatrix[4] * r + colorMatrix[5] * g + 
                   colorMatrix[6] * b + colorMatrix[7] * a;
        float nb = colorMatrix[8] * r + colorMatrix[9] * g + 
                   colorMatrix[10] * b + colorMatrix[11] * a;
        
        r = fminf(fmaxf(nr, 0.0f), 255.0f);
        g = fminf(fmaxf(ng, 0.0f), 255.0f);
        b = fminf(fmaxf(nb, 0.0f), 255.0f);
    }
    
    // Store corrected pixel
    output[idx] = (unsigned char)r;
    output[idx + 1] = (unsigned char)g;
    output[idx + 2] = (unsigned char)b;
    output[idx + 3] = (unsigned char)a;
}

// Multi-layer focus stacking kernel
__global__ void focusStackKernel(const unsigned char** __restrict__ layers,
                                  unsigned char* __restrict__ output,
                                  const float* __restrict__ focusWeights,
                                  int numLayers, int width, int height) {
    int x = blockIdx.x * blockDim.x + threadIdx.x;
    int y = blockIdx.y * blockDim.y + threadIdx.y;
    
    if (x >= width || y >= height) return;
    
    int idx = (y * width + x) * 4;
    
    float r = 0.0f, g = 0.0f, b = 0.0f, a = 0.0f;
    float totalWeight = 0.0f;
    
    // Weighted average of all focus layers
    for (int layer = 0; layer < numLayers; layer++) {
        float weight = focusWeights[layer];
        const unsigned char* layerData = layers[layer];
        
        r += layerData[idx] * weight;
        g += layerData[idx + 1] * weight;
        b += layerData[idx + 2] * weight;
        a += layerData[idx + 3] * weight;
        totalWeight += weight;
    }
    
    if (totalWeight > 0.0f) {
        output[idx] = (unsigned char)(r / totalWeight);
        output[idx + 1] = (unsigned char)(g / totalWeight);
        output[idx + 2] = (unsigned char)(b / totalWeight);
        output[idx + 3] = (unsigned char)(a / totalWeight);
    }
}

// Sharpening kernel for enhanced detail
__global__ void sharpenKernel(const unsigned char* __restrict__ input,
                               unsigned char* __restrict__ output,
                               int width, int height, float amount) {
    int x = blockIdx.x * blockDim.x + threadIdx.x;
    int y = blockIdx.y * blockDim.y + threadIdx.y;
    
    if (x < 1 || x >= width - 1 || y < 1 || y >= height - 1) return;
    
    int idx = (y * width + x) * 4;
    int stride = width * 4;
    
    // Unsharp mask
    for (int c = 0; c < 3; c++) {
        float center = input[idx + c];
        float blur = (
            input[idx - stride - 4 + c] + input[idx - stride + c] + input[idx - stride + 4 + c] +
            input[idx - 4 + c] + input[idx + c] + input[idx + 4 + c] +
            input[idx + stride - 4 + c] + input[idx + stride + c] + input[idx + stride + 4 + c]
        ) / 9.0f;
        
        float sharpened = center + amount * (center - blur);
        output[idx + c] = (unsigned char)fminf(fmaxf(sharpened, 0.0f), 255.0f);
    }
    output[idx + 3] = input[idx + 3]; // Alpha unchanged
}

extern "C" void processTile(unsigned char* input, unsigned char* output,
                           int width, int height, float* colorMatrix) {
    dim3 blockSize(16, 16);
    dim3 gridSize((width + blockSize.x - 1) / blockSize.x,
                  (height + blockSize.y - 1) / blockSize.y);
    
    decompressTileKernel<<<gridSize, blockSize>>>(input, output, width, height, colorMatrix);
    cudaDeviceSynchronize();
}

extern "C" void processFocusStack(unsigned char** layers, unsigned char* output,
                                  float* focusWeights, int numLayers,
                                  int width, int height) {
    dim3 blockSize(16, 16);
    dim3 gridSize((width + blockSize.x - 1) / blockSize.x,
                  (height + blockSize.y - 1) / blockSize.y);
    
    focusStackKernel<<<gridSize, blockSize>>>(layers, output, focusWeights,
                                               numLayers, width, height);
    cudaDeviceSynchronize();
}

extern "C" void applySharpen(unsigned char* input, unsigned char* output,
                             int width, int height, float amount) {
    dim3 blockSize(16, 16);
    dim3 gridSize((width + blockSize.x - 1) / blockSize.x,
                  (height + blockSize.y - 1) / blockSize.y);
    
    sharpenKernel<<<gridSize, blockSize>>>(input, output, width, height, amount);
    cudaDeviceSynchronize();
}
