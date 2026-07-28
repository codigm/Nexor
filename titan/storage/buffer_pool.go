package storage

import (
	"fmt"
	"os"
	"sync"
)

type BufferPool struct {
	mu       sync.Mutex
	pages    map[uint16]*Page
	lruList  []uint16 // simplified LRU
	capacity int
	file     *os.File
}

func NewBufferPool(capacity int, dbFile string) (*BufferPool, error) {
	f, err := os.OpenFile(dbFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	
	return &BufferPool{
		pages:    make(map[uint16]*Page),
		lruList:  make([]uint16, 0, capacity),
		capacity: capacity,
		file:     f,
	}, nil
}

func (bp *BufferPool) GetPage(pageID uint16) (*Page, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if page, exists := bp.pages[pageID]; exists {
		bp.markUsed(pageID)
		return page, nil
	}

	if len(bp.pages) >= bp.capacity {
		bp.evict()
	}

	page := NewPage(pageID)
	// ReadAt using standard file I/O as requested by user
	_, err := bp.file.ReadAt(page.Data, int64(pageID)*PageSize)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("failed to read page: %w", err)
	}

	bp.pages[pageID] = page
	bp.lruList = append(bp.lruList, pageID)
	return page, nil
}

func (bp *BufferPool) FlushPage(pageID uint16) error {
	// Not taking lock here as it's called by evict() and Close() which might already hold it.
	// Wait, evict() holds lock, so FlushPage must NOT take lock.
	// We'll rename this to flushPageInternal and export a wrapper.
	return bp.flushPageInternal(pageID)
}

func (bp *BufferPool) flushPageInternal(pageID uint16) error {
	page, exists := bp.pages[pageID]
	if !exists {
		return nil // Not in memory, nothing to flush
	}

	// WriteAt using standard file I/O
	_, err := bp.file.WriteAt(page.Data, int64(pageID)*PageSize)
	return err
}

func (bp *BufferPool) markUsed(pageID uint16) {
	for i, id := range bp.lruList {
		if id == pageID {
			bp.lruList = append(bp.lruList[:i], bp.lruList[i+1:]...)
			bp.lruList = append(bp.lruList, pageID)
			break
		}
	}
}

func (bp *BufferPool) evict() {
	if len(bp.lruList) == 0 {
		return
	}
	evictID := bp.lruList[0]
	bp.lruList = bp.lruList[1:]
	bp.flushPageInternal(evictID) 
	delete(bp.pages, evictID)
}

func (bp *BufferPool) Close() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	for id := range bp.pages {
		bp.flushPageInternal(id) 
	}
	bp.file.Close()
}
