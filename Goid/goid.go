package Goid

import (
	"sync/atomic"
	"unsafe"
)

type stack struct {
	lo uintptr
	hi uintptr
}

type guintptr uintptr
type waitReason uint8
type muintptr uintptr

type gobuf struct {
	// The offsets of sp, pc, and g are known to (hard-coded in) libmach.
	//
	// ctxt is unusual with respect to GC: it may be a
	// heap-allocated funcval, so GC needs to track it, but it
	// needs to be set and cleared from assembly, where it's
	// difficult to have write barriers. However, ctxt is really a
	// saved, live register, and we only ever exchange it between
	// the real register and the gobuf. Hence, we treat it as a
	// root during stack scanning, which means assembly that saves
	// and restores it doesn't need write barriers. It's still
	// typed as a pointer so that any other writes from Go get
	// write barriers.
	sp   uintptr
	pc   uintptr
	g    guintptr
	ctxt unsafe.Pointer
	ret  uintptr
	lr   uintptr
	bp   uintptr // for framepointer-enabled architectures
}
type _panic struct {
	argp unsafe.Pointer // pointer to arguments of deferred call run during panic; cannot move - known to liblink
	arg  any            // argument to panic
	link *_panic        // link to earlier panic

	// startPC and startSP track where _panic.start was called.
	startPC uintptr
	startSP unsafe.Pointer

	// The current stack frame that we're running deferred calls for.
	sp unsafe.Pointer
	lr uintptr
	fp unsafe.Pointer

	// retpc stores the PC where the panic should jump back to, if the
	// function last returned by _panic.next() recovers the panic.
	retpc uintptr

	// Extra state for handling open-coded defers.
	deferBitsPtr *uint8
	slotsPtr     unsafe.Pointer

	recovered   bool // whether this panic has been recovered
	goexit      bool
	deferreturn bool
}
type _defer struct {
	heap      bool
	rangefunc bool    // true for rangefunc list
	sp        uintptr // sp at time of defer
	pc        uintptr // pc at time of defer
	fn        func()  // can be nil for open-coded defers
	link      *_defer // next defer on G; can point to either heap or stack!

	// If rangefunc is true, *head is the head of the atomic linked list
	// during a range-over-func execution.
	head *atomic.Pointer[_defer]
}

type g struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the //go:systemstack stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	stack       stack   // offset known to runtime/cgo
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	_panic *_panic // innermost panic - offset known to liblink
	_defer *_defer // innermost defer
	//m         *m      // current m; offset known to arm liblink
	m         unsafe.Pointer
	sched     gobuf
	syscallsp uintptr // if status==Gsyscall, syscallsp = sched.sp to use during gc
	syscallpc uintptr // if status==Gsyscall, syscallpc = sched.pc to use during gc
	stktopsp  uintptr // expected sp at top of stack, to check in traceback
	// param is a generic pointer parameter field used to pass
	// values in particular contexts where other storage for the
	// parameter would be difficult to find. It is currently used
	// in four ways:
	// 1. When a channel operation wakes up a blocked goroutine, it sets param to
	//    point to the sudog of the completed blocking operation.
	// 2. By gcAssistAlloc1 to signal back to its caller that the goroutine completed
	//    the GC cycle. It is unsafe to do so in any other way, because the goroutine's
	//    stack may have moved in the meantime.
	// 3. By debugCallWrap to pass parameters to a new goroutine because allocating a
	//    closure in the runtime is forbidden.
	// 4. When a panic is recovered and control returns to the respective frame,
	//    param may point to a savedOpenDeferState.
	param        unsafe.Pointer
	atomicstatus atomic.Uint32
	stackLock    uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
	goid         uint64
}

// current g struct pointer was stored inside M thread local storage
func getg() *g

// get what current goroutine goid is
//
//go:nosplit
func GetGoid() (sucess bool, goid uint64) {
	curg := getg()
	sucess = curg != nil
	if sucess {
		goid = curg.goid
	} else {
		goid = 0
	}
	return
}
