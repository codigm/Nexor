#ifndef TENSOR_H
#define TENSOR_H

#include <stddef.h>
#include <immintrin.h> // For AVX2

typedef struct {
    float* data;
    size_t rows;
    size_t cols;
    size_t stride;
} Tensor;

// CGo functions
void matmul_avx2(Tensor* A, Tensor* B, Tensor* C);
void attention_avx2(Tensor* Q, Tensor* K, Tensor* V, Tensor* Out);
void softmax_avx2(float* x, size_t size);
void gelu_avx2(float* x, size_t size);
void layernorm_avx2(float* x, float* gamma, float* beta, size_t size);

#endif
