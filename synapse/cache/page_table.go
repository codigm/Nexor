package cache

// Maps logical sequence to physical non-contiguous pages
type PageTable struct {
	RequestID     string
	LogicalToPhys []int // Index = logical block, Value = physical page index
}

func NewPageTable(reqID string) *PageTable {
	return &PageTable{
		RequestID:     reqID,
		LogicalToPhys: make([]int, 0),
	}
}

func (pt *PageTable) AddPhysicalPage(physIdx int) {
	pt.LogicalToPhys = append(pt.LogicalToPhys, physIdx)
}
