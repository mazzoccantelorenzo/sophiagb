package gpu

// Here we organize all the info related to the GPU
// For now this is what I think, the GPU is basically a CPU, simply its memory
// is called VRAM

// The main difference for now is I think the gpu only reads, doesn't write on its memory

type VMemoryBus interface {
	Read(addr uint16) uint8
}

type GPU struct {
	VMemory VMemoryBus
}

/*
New creates a new GPU instance.
The Gpu receives memory from main through dependency injection
*/

func New(mem VMemoryBus) *GPU {
	return &GPU{
		VMemory: mem,
	}
}
