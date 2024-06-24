// Package machine is imported by the runtime and allows the target to implement
// some hooks, most importantly rt0.
package machine

import (
	"unsafe"

	"github.com/clktmr/n64/rcp/cpu"
	"github.com/clktmr/n64/rcp/periph"
)

// TODO use register definitions from isviewer package
var regs *registers = (*registers)(unsafe.Pointer(baseAddr))

const token = 0x49533634
const baseAddr uintptr = cpu.KSEG1 | 0x13ff_0000
const bufferSize = 512 // actually 64*1024 - 0x20, but ISViewer.buf will allocate this

type registers struct {
	token    periph.U32
	readPtr  periph.U32
	_        periph.U32
	_        periph.U32
	_        periph.U32
	writePtr periph.U32
	_        periph.U32
	_        periph.U32
	buf      [bufferSize / 4]periph.U32
}

// Writes to ISViewer registers, regardless if a ISViewer is present or not.  Is
// rather slow, because it avoids using DMA.  Only intended as a fail safe
// logger in very early boot.
//
//go:nowritebarrierrec
//go:nosplit
//go:linkname DefaultWrite runtime.defaultWrite
func DefaultWrite(fd int, p []byte) int {
	n := len(p)
	if n > bufferSize {
		n = bufferSize
	}

	for i := 0; i < n/4; i++ {
		pi := 4 * i
		regs.buf[i].Store(0 |
			uint32(p[pi])<<24 |
			uint32(p[pi+1])<<16 |
			uint32(p[pi+2])<<8 |
			uint32(p[pi+3]))
	}

	if n%4 != 0 {
		var tail uint32
		for i := 0; i < n%4; i++ {
			base := len(p) - n%4
			tail |= uint32(p[base+i]) << ((3 - i) * 8)
		}
		regs.buf[n/4].Store(tail)
	}

	regs.readPtr.Store(0)
	regs.writePtr.Store(uint32(n))
	regs.token.Store(token)

	for regs.readPtr.Load() != regs.writePtr.Load() {
		// wait
	}

	regs.token.Store(0x0)

	return n
}
