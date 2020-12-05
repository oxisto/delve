//+build arm64 darwin,macnative

package native

// #include "threads_darwin.h"
import "C"
import (
	"errors"
	"fmt"

	"github.com/go-delve/delve/pkg/proc"
	"golang.org/x/arch/arm64/arm64asm"
)

// Regs represents CPU registers on an AMD64 processor.
type Regs struct {
	x0     uint64
	x1     uint64
	x2     uint64
	x3     uint64
	x4     uint64
	x5     uint64
	x6     uint64
	x7     uint64
	x8     uint64
	x9     uint64
	x10    uint64
	x11    uint64
	x12    uint64
	x13    uint64
	x14    uint64
	x15    uint64
	x16    uint64
	x17    uint64
	x18    uint64
	x19    uint64
	x20    uint64
	x21    uint64
	x22    uint64
	x23    uint64
	x24    uint64
	x25    uint64
	x26    uint64
	x27    uint64
	x28    uint64
	x29    uint64
	x30    uint64
	gsBase uint64
	fpregs []proc.Register
}

func (r *Regs) Slice(floatingPoint bool) ([]proc.Register, error) {
	var regs = []struct {
		k string
		v uint64
	}{
		{"X0", r.x0},
		{"X1", r.x0},
		{"X2", r.x2},
		{"X3", r.x3},
		{"X4", r.x4},
		{"X5", r.x5},
		{"X6", r.x6},
		{"X7", r.x7},
		{"X8", r.x8},
		{"X9", r.x9},
		{"X10", r.x10},
		{"X11", r.x11},
		{"X12", r.x12},
		{"X13", r.x13},
		{"X14", r.x14},
		{"X15", r.x15},
		{"X16", r.x16},
		{"X17", r.x17},
		{"X18", r.x18},
		{"X19", r.x19},
		{"X20", r.x20},
		{"X21", r.x21},
		{"X22", r.x22},
		{"X23", r.x23},
		{"X24", r.x24},
		{"X25", r.x25},
		{"X26", r.x26},
		{"X27", r.x27},
		{"X28", r.x28},
		{"X29", r.x29},
		{"X30", r.x30},
		{"Gs_base", r.gsBase},
	}
	out := make([]proc.Register, 0, len(regs)+len(r.fpregs))
	for _, reg := range regs {
		if reg.k == "Rflags" {
			out = proc.AppendUint64Register(out, reg.k, reg.v)
		} else {
			out = proc.AppendUint64Register(out, reg.k, reg.v)
		}
	}
	if floatingPoint {
		out = append(out, r.fpregs...)
	}
	return out, nil
}

// PC returns the current program counter
// i.e. the RIP CPU register.
/*func (r *Regs) PC() uint64 {
	return r.xip
}*/

// SP returns the stack pointer location,
// i.e. the RSP register.
/*func (r *Regs) SP() uint64 {
	return r.xsp
}

func (r *Regs) BP() uint64 {
	return r.xbp
}*/

// TLS returns the value of the register
// that contains the location of the thread
// local storage segment.
func (r *Regs) TLS() uint64 {
	return r.gsBase
}

func (r *Regs) GAddr() (uint64, bool) {
	return 0, false
}

// SetPC sets the RIP register to the value specified by `pc`.
func (thread *nativeThread) SetPC(pc uint64) error {
	kret := C.set_pc(thread.os.threadAct, C.uint64_t(pc))
	if kret != C.KERN_SUCCESS {
		return fmt.Errorf("could not set pc")
	}
	return nil
}

// SetSP sets the RSP register to the value specified by `pc`.
func (thread *nativeThread) SetSP(sp uint64) error {
	return errors.New("not implemented")
}

func (thread *nativeThread) SetDX(dx uint64) error {
	return errors.New("not implemented")
}

func (r *Regs) Get(n int) (uint64, error) {
	reg := arm64asm.Reg(n)
	const (
		mask8  = 0x000f
		mask16 = 0x00ff
		mask32 = 0xffff
	)

	switch reg {
	// 64-bit
	case arm64asm.X0:
		return r.x0, nil
	case arm64asm.X1:
		return r.x0, nil
	case arm64asm.X2:
		return r.x2, nil
	case arm64asm.X3:
		return r.x3, nil
	case arm64asm.X4:
		return r.x4, nil
	case arm64asm.X5:
		return r.x5, nil
	case arm64asm.X6:
		return r.x6, nil
	case arm64asm.X7:
		return r.x7, nil
	case arm64asm.X8:
		return r.x8, nil
	case arm64asm.X9:
		return r.x9, nil
	case arm64asm.X10:
		return r.x10, nil
	case arm64asm.X11:
		return r.x11, nil
	case arm64asm.X12:
		return r.x12, nil
	case arm64asm.X13:
		return r.x13, nil
	case arm64asm.X14:
		return r.x14, nil
	case arm64asm.X15:
		return r.x15, nil
	}

	return 0, proc.ErrUnknownRegister
}

func registers(thread *nativeThread) (proc.Registers, error) {
	var state C.arm_thread_state64_t
	var identity C.thread_identifier_info_data_t
	kret := C.get_registers(C.mach_port_name_t(thread.os.threadAct), &state)
	if kret != C.KERN_SUCCESS {
		return nil, fmt.Errorf("could not get registers")
	}
	kret = C.get_identity(C.mach_port_name_t(thread.os.threadAct), &identity)
	if kret != C.KERN_SUCCESS {
		return nil, fmt.Errorf("could not get thread identity informations")
	}
	/*
		thread_identifier_info::thread_handle contains the base of the
		thread-specific data area, which on x86 and x86_64 is the thread’s base
		address of the %gs segment. 10.9.2 xnu-2422.90.20/osfmk/kern/thread.c
		thread_info_internal() gets the value from
		machine_thread::cthread_self, which is the same value used to set the
		%gs base in xnu-2422.90.20/osfmk/i386/pcb_native.c
		act_machine_switch_pcb().
		--
		comment copied from chromium's crashpad
		https://chromium.googlesource.com/crashpad/crashpad/+/master/snapshot/mac/process_reader.cc
	*/
	regs := &Regs{
		x0:  uint64(state.__rax),
		x1:  uint64(state.__rbx),
		x2:  uint64(state.__rcx),
		x3:  uint64(state.__rdx),
		x4:  uint64(state.__rdi),
		x5:  uint64(state.__rsi),
		x6:  uint64(state.__rbp),
		x7:  uint64(state.__rsp),
		x8:  uint64(state.__r8),
		x9:  uint64(state.__r9),
		x10: uint64(state.__r10),
		x11: uint64(state.__r11),
		x12: uint64(state.__r12),
		x13: uint64(state.__r13),
		x14: uint64(state.__r14),
		x15: uint64(state.__r15),
		x16: uint64(state.__rip),
		/*rflags: uint64(state.__rflags),
		cs:     uint64(state.__cs),
		fs:     uint64(state.__fs),
		gs:     uint64(state.__gs),*/
		gsBase: uint64(identity.thread_handle),
	}

	return regs, nil
}

func (r *Regs) Copy() (proc.Registers, error) {
	//TODO(aarzilli): implement this to support function calls
	return nil, nil
}
