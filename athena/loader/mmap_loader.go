package loader

import (
	"fmt"
	"os"
	"syscall"
)

// Zero-copy model loading via OS mmap
func LoadModel(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := info.Size()

	data, err := syscall.Mmap(
		int(f.Fd()), 0, int(fileSize),
		syscall.PROT_READ,
		syscall.MAP_SHARED, // OS manages page loading
	)
	if err != nil {
		return nil, fmt.Errorf("failed to mmap file: %v", err)
	}

	// Suggest huge pages for 2MB pages
	syscall.Madvise(data, syscall.MADV_HUGEPAGE)
	return data, nil
}
