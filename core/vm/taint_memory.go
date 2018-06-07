// Author: Jianbo-Gao
// Recording EVM memory for taint analysis.

package vm

import "fmt"
import "encoding/json"

// Memory implements a simple memory model for the ethereum virtual machine.
type TaintMemory struct {
	store       []int
	lastGasCost uint64
}

func NewTaintMemory() *TaintMemory {
	return &TaintMemory{}
}

// Set sets offset + size to value
func (m *TaintMemory) Set(offset, size uint64, value []int) {
	// length of store may never be less than offset + size.
	// The store should be resized PRIOR to setting the memory
	if size > uint64(len(m.store)) {
		panic("INVALID memory: store empty")
	}

	// It's possible the offset is greater than 0 and size equals 0. This is because
	// the calcMemSize (common.go) could potentially return 0 when size is zero (NO-OP)
	if size > 0 {
		copy(m.store[offset:offset+size], value)
	}
}

// Resize resizes the memory to size
func (m *TaintMemory) Resize(size uint64) {
	if uint64(m.Len()) < size {
		m.store = append(m.store, make([]int, size-uint64(m.Len()))...)
	}
}

// Get returns offset + size as a new slice
func (m *TaintMemory) Get(offset, size int64) (cpy []int) {
	if size == 0 {
		return nil
	}

	if len(m.store) > int(offset) {
		cpy = make([]int, size)
		copy(cpy, m.store[offset:offset+size])

		return
	}

	return
}

// GetPtr returns the offset + size
func (m *TaintMemory) GetPtr(offset, size int64) []int {
	if size == 0 {
		return nil
	}

	if len(m.store) > int(offset) {
		return m.store[offset : offset+size]
	}

	return nil
}

// Len returns the length of the backing slice
func (m *TaintMemory) Len() int {
	return len(m.store)
}

// Data returns the backing slice
func (m *TaintMemory) Data() []int {
	return m.store
}

func (m *TaintMemory) Print() {
	fmt.Printf("### mem %d bytes ###\n", len(m.store))
	if len(m.store) > 0 {
		addr := 0
		for i := 0; i+32 <= len(m.store); i += 32 {
			fmt.Printf("%03d: % x\n", addr, m.store[i:i+32])
			addr++
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("####################")
}

func (m *TaintMemory) JPrint() {
	if j_data, err := json.Marshal(m.store); err == nil {
		fmt.Printf("TaintMemory:%s\n", j_data)
	}
}
