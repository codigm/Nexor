package storage

import "encoding/binary"

const PageSize = 4096

// A Page is a 4KB block in memory or on disk.
// Layout:
// [0-2]: Page ID
// [2-4]: Number of slots
// [4-8]: Free space pointer
// [8-...]: Slot Array (offset, length pairs)
// [...-4096]: Row Data (growing backwards)
type Page struct {
	ID   uint16
	Data []byte
}

func NewPage(id uint16) *Page {
	p := &Page{
		ID:   id,
		Data: make([]byte, PageSize),
	}
	binary.LittleEndian.PutUint16(p.Data[0:2], id)
	binary.LittleEndian.PutUint16(p.Data[2:4], 0) // num slots
	binary.LittleEndian.PutUint32(p.Data[4:8], PageSize) // free space pointer
	return p
}

func (p *Page) InsertTuple(data []byte) bool {
	numSlots := binary.LittleEndian.Uint16(p.Data[2:4])
	freePtr := binary.LittleEndian.Uint32(p.Data[4:8])

	slotSize := 4 // 2 bytes offset, 2 bytes length
	headerSize := 8 + int(numSlots)*slotSize

	if int(freePtr)-headerSize < len(data)+slotSize {
		return false // Not enough space
	}

	// Calculate new free pointer
	newFreePtr := freePtr - uint32(len(data))

	// Write data
	copy(p.Data[newFreePtr:freePtr], data)

	// Write slot
	slotOffset := 8 + int(numSlots)*slotSize
	binary.LittleEndian.PutUint16(p.Data[slotOffset:slotOffset+2], uint16(newFreePtr))
	binary.LittleEndian.PutUint16(p.Data[slotOffset+2:slotOffset+4], uint16(len(data)))

	// Update header
	binary.LittleEndian.PutUint16(p.Data[2:4], numSlots+1)
	binary.LittleEndian.PutUint32(p.Data[4:8], newFreePtr)

	return true
}

func (p *Page) GetTuple(slotIdx uint16) []byte {
	numSlots := binary.LittleEndian.Uint16(p.Data[2:4])
	if slotIdx >= numSlots {
		return nil
	}

	slotOffset := 8 + int(slotIdx)*4
	offset := binary.LittleEndian.Uint16(p.Data[slotOffset : slotOffset+2])
	length := binary.LittleEndian.Uint16(p.Data[slotOffset+2 : slotOffset+4])

	return p.Data[offset : offset+length]
}
