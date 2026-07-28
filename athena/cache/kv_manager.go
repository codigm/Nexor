package cache

import "sync"

const PageSize = 16 // 16 tokens per page

type PagedAttentionCache struct {
	mu          sync.Mutex
	PhysicalMem []float32 // The giant tensor holding all KV pages
	FreePages   []int
}

func NewPagedAttentionCache(totalTokens int) *PagedAttentionCache {
	numPages := totalTokens / PageSize
	freePages := make([]int, numPages)
	for i := 0; i < numPages; i++ {
		freePages[i] = i
	}
	return &PagedAttentionCache{
		PhysicalMem: make([]float32, totalTokens*128), // Assuming 128-dim embeddings
		FreePages:   freePages,
	}
}

func (c *PagedAttentionCache) AllocatePage() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.FreePages) == 0 {
		return -1 // OOM
	}
	pageIdx := c.FreePages[0]
	c.FreePages = c.FreePages[1:]
	return pageIdx
}
