package tensor

/*
#cgo CFLAGS: -O3 -mavx2 -mfma -lm
#include "tensor.h"
*/
import "C"
import "unsafe"

// Tensor is a Go wrapper around the C.Tensor struct
type Tensor struct {
	cTensor C.Tensor
}

// NewTensor creates a new Tensor wrapper from existing data
func NewTensor(data []float32, rows, cols, stride int) *Tensor {
	t := &Tensor{}
	t.cTensor.data = (*C.float)(unsafe.Pointer(&data[0]))
	t.cTensor.rows = C.size_t(rows)
	t.cTensor.cols = C.size_t(cols)
	t.cTensor.stride = C.size_t(stride)
	return t
}

// MatmulAVX2 performs matrix multiplication using AVX2
func MatmulAVX2(a, b, c *Tensor) {
	C.matmul_avx2(&a.cTensor, &b.cTensor, &c.cTensor)
}

// AttentionAVX2 performs scaled dot-product attention
func AttentionAVX2(q, k, v, out *Tensor) {
	C.attention_avx2(&q.cTensor, &k.cTensor, &v.cTensor, &out.cTensor)
}

// SoftmaxAVX2 applies softmax in-place
func SoftmaxAVX2(data []float32) {
	if len(data) == 0 {
		return
	}
	C.softmax_avx2((*C.float)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
}

// GeluAVX2 applies GELU activation in-place
func GeluAVX2(data []float32) {
	if len(data) == 0 {
		return
	}
	C.gelu_avx2((*C.float)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
}

// LayerNormAVX2 applies layer normalization in-place
func LayerNormAVX2(data, gamma, beta []float32) {
	if len(data) == 0 || len(gamma) == 0 || len(beta) == 0 {
		return
	}
	C.layernorm_avx2(
		(*C.float)(unsafe.Pointer(&data[0])),
		(*C.float)(unsafe.Pointer(&gamma[0])),
		(*C.float)(unsafe.Pointer(&beta[0])),
		C.size_t(len(data)),
	)
}
