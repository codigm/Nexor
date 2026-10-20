#include <stddef.h>
#include <immintrin.h> // For AVX2
#include <math.h>

// AVX2 Cosine Similarity
float cosine_similarity_avx2(float* vecA, float* vecB, size_t dim) {
    __m256 sum_dot = _mm256_setzero_ps();
    __m256 sum_a2 = _mm256_setzero_ps();
    __m256 sum_b2 = _mm256_setzero_ps();
    
    for (size_t i = 0; i < dim; i += 8) {
        __m256 a = _mm256_loadu_ps(&vecA[i]);
        __m256 b = _mm256_loadu_ps(&vecB[i]);
        
        sum_dot = _mm256_fmadd_ps(a, b, sum_dot);
        sum_a2 = _mm256_fmadd_ps(a, a, sum_a2);
        sum_b2 = _mm256_fmadd_ps(b, b, sum_b2);
    }
    
    float dot_arr[8], a2_arr[8], b2_arr[8];
    _mm256_storeu_ps(dot_arr, sum_dot);
    _mm256_storeu_ps(a2_arr, sum_a2);
    _mm256_storeu_ps(b2_arr, sum_b2);
    
    float dot = 0.0f, a2 = 0.0f, b2 = 0.0f;
    for (int i = 0; i < 8; i++) {
        dot += dot_arr[i];
        a2 += a2_arr[i];
        b2 += b2_arr[i];
    }
    
    return dot / (sqrtf(a2) * sqrtf(b2) + 1e-8f);
}

// AVX2 Euclidean Distance
float euclidean_distance_avx2(float* vecA, float* vecB, size_t dim) {
    __m256 sum_diff2 = _mm256_setzero_ps();
    
    for (size_t i = 0; i < dim; i += 8) {
        __m256 a = _mm256_loadu_ps(&vecA[i]);
        __m256 b = _mm256_loadu_ps(&vecB[i]);
        __m256 diff = _mm256_sub_ps(a, b);
        sum_diff2 = _mm256_fmadd_ps(diff, diff, sum_diff2);
    }
    
    float diff2_arr[8];
    _mm256_storeu_ps(diff2_arr, sum_diff2);
    
    float diff2 = 0.0f;
    for (int i = 0; i < 8; i++) {
        diff2 += diff2_arr[i];
    }
    
    return sqrtf(diff2);
}

// AVX2 Dot Product
float dot_product_avx2(float* vecA, float* vecB, size_t dim) {
    __m256 sum_dot = _mm256_setzero_ps();
    
    for (size_t i = 0; i < dim; i += 8) {
        __m256 a = _mm256_loadu_ps(&vecA[i]);
        __m256 b = _mm256_loadu_ps(&vecB[i]);
        sum_dot = _mm256_fmadd_ps(a, b, sum_dot);
    }
    
    float dot_arr[8];
    _mm256_storeu_ps(dot_arr, sum_dot);
    
    float dot = 0.0f;
    for (int i = 0; i < 8; i++) {
        dot += dot_arr[i];
    }
    
    return dot;
}
