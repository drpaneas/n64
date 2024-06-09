package periph

import (
	"unsafe"

	"github.com/clktmr/n64/debug"
	"github.com/clktmr/n64/rcp/cpu"
)

// TODO protect access to DMA with mutex

// Loads bytes from PI bus into RDRAM via DMA
func DMALoad(piAddr uintptr, size int) []byte {
	debug.Assert(size%2 == 0, "pi: unaligned dma load")

	buf := cpu.MakePaddedSliceAligned[byte](size, 2)
	addr := uintptr(unsafe.Pointer(unsafe.SliceData(buf)))
	regs.dramAddr.Store(cpu.PhysicalAddress(addr))
	regs.cartAddr.Store(cpu.PhysicalAddress(piAddr))

	cpu.InvalidateSlice(buf)

	regs.writeLen.Store(uint32(size - 1))

	waitDMA()

	return buf
}

// Stores bytes from RDRAM to PI bus via DMA
func DMAStore(piAddr uintptr, p []byte) {
	buf := p

	debug.Assert(len(p)%2 == 0, "pi: unaligned dma store")

	p = cpu.PaddedSlice(p)

	addr := uintptr(unsafe.Pointer(unsafe.SliceData(buf)))
	regs.dramAddr.Store(cpu.PhysicalAddress(addr))
	regs.cartAddr.Store(cpu.PhysicalAddress(piAddr))

	cpu.WritebackSlice(buf)

	regs.readLen.Store(uint32(len(buf) - 1))

	waitDMA()
}

// Blocks until DMA has finished.
func waitDMA() {
	for {
		// TODO runtime.Gosched() ?
		if regs.status.Load()&(dmaBusy|ioBusy) == 0 {
			break
		}
	}

}
