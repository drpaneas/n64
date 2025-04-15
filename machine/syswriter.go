package machine

import (
	"unsafe"

	"github.com/drpaneas/n64/rcp/cpu"
	"github.com/drpaneas/n64/rcp/periph"
)

var regs *registers = (*registers)(unsafe.Pointer(baseAddr))

const token = 0x49533634
const baseAddr uintptr = cpu.KSEG1 | 0x13ff_0000
const bufferSize = 512 // actually 64*1024 - 0x20, but ISViewer.buf will allocate this

type registers struct {
	token    periph.U32
	readPtr  periph.U32
	_        [3]periph.U32
	writePtr periph.U32
	_        [2]periph.U32
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
	written := len(p)
	for len(p) > 0 {
		n := len(p)
		if n > bufferSize {
			n = bufferSize
		}

		for i := 0; i < n/4; i++ {
			pi := 4 * i
			regs.buf[i].StoreSafe(0 |
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
			regs.buf[n/4].StoreSafe(tail)
		}

		regs.readPtr.StoreSafe(0)
		regs.writePtr.StoreSafe(uint32(n))
		regs.token.StoreSafe(token)

		for regs.readPtr.LoadSafe() != regs.writePtr.LoadSafe() {
			// wait
		}

		regs.token.StoreSafe(0x0)
		p = p[n:]
	}

	return written
}

type defaultWriter int

const DefaultWriter defaultWriter = 0

func (v defaultWriter) Write(p []byte) (int, error) {
	return DefaultWrite(int(v), p), nil
}
