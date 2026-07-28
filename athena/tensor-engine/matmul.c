#include "tensor.h"

// AVX2 matrix multiply — 8 floats in parallel
void matmul_avx2(Tensor* A, Tensor* B, Tensor* C) {
    size_t M = A->rows;
    size_t K = A->cols;
    size_t N = B->cols;
    
    for (size_t i = 0; i < M; i++) {
        for (size_t j = 0; j < N; j += 8) {
            __m256 sum = _mm256_setzero_ps();
            for (size_t k = 0; k < K; k++) {
                __m256 a = _mm256_broadcast_ss(&A->data[i * A->stride + k]);
                __m256 b = _mm256_loadu_ps(&B->data[k * B->stride + j]);
                sum = _mm256_fmadd_ps(a, b, sum); // fused multiply-add
            }
            _mm256_storeu_ps(&C->data[i * C->stride + j], sum);
        }
    }
}
