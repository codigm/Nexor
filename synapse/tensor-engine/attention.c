#include "tensor.h"
#include <math.h>
#include <stdlib.h>

// Scaled dot-product attention implementation
// Helper to calculate softmax on a row
static void inline_softmax(float* x, size_t size) {
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

void attention_avx2(Tensor* Q, Tensor* K, Tensor* V, Tensor* Out) {
    size_t seq_len = Q->rows;
    size_t d_k = Q->cols;
    size_t d_v = V->cols;
    
    float scale = 1.0f / sqrtf((float)d_k);
    
    float* scores = (float*)malloc(seq_len * seq_len * sizeof(float));

    // Q * K^T
    for (size_t i = 0; i < seq_len; ++i) {
        for (size_t j = 0; j < seq_len; ++j) {
            float sum = 0.0f;
            for (size_t k = 0; k < d_k; ++k) {
                sum += Q->data[i * Q->stride + k] * K->data[j * K->stride + k]; 
            }
            scores[i * seq_len + j] = sum * scale;
        }
        
        // Softmax
        inline_softmax(&scores[i * seq_len], seq_len);
    }

    // scores * V
    for (size_t i = 0; i < seq_len; ++i) {
        for (size_t j = 0; j < d_v; j += 8) {
            __m256 out_vec = _mm256_setzero_ps();
            for (size_t k = 0; k < seq_len; ++k) {
                __m256 score_vec = _mm256_set1_ps(scores[i * seq_len + k]);
                __m256 v_vec = _mm256_loadu_ps(&V->data[k * V->stride + j]);
                out_vec = _mm256_fmadd_ps(score_vec, v_vec, out_vec);
            }
            _mm256_storeu_ps(&Out->data[i * Out->stride + j], out_vec);
        }
    }

    free(scores);
}
