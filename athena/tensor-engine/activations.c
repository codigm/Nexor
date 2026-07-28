#include "tensor.h"
#include <immintrin.h>
#include <math.h>

void softmax_avx2(float* x, size_t size) {
    float max_val = x[0];
    for (size_t i = 1; i < size; ++i) {
        if (x[i] > max_val) max_val = x[i];
    }
    float sum = 0.0f;
    for (size_t i = 0; i < size; ++i) {
        x[i] = expf(x[i] - max_val);
        sum += x[i];
    }
    for (size_t i = 0; i < size; ++i) {
        x[i] /= sum;
    }
}

void gelu_avx2(float* x, size_t size) {
    // GELU(x) = 0.5 * x * (1 + tanh(sqrt(2/pi) * (x + 0.044715 * x^3)))
    for (size_t i = 0; i < size; ++i) {
        float val = x[i];
        float cube = val * val * val;
        float arg = 0.7978845608f * (val + 0.044715f * cube);
        x[i] = 0.5f * val * (1.0f + tanhf(arg));
    }
}

void layernorm_avx2(float* x, float* gamma, float* beta, size_t size) {
    float sum = 0.0f;
    for (size_t i = 0; i < size; ++i) {
        sum += x[i];
    }
    float mean = sum / size;
    
    float sq_sum = 0.0f;
    for (size_t i = 0; i < size; ++i) {
        float diff = x[i] - mean;
        sq_sum += diff * diff;
    }
    float variance = sq_sum / size;
    float stddev = sqrtf(variance + 1e-5f);
    
    for (size_t i = 0; i < size; ++i) {
        x[i] = ((x[i] - mean) / stddev) * gamma[i] + beta[i];
    }
}
