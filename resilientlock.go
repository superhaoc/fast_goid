import (
"sync"
"sync/atomic"
)


// simple resilient lock,with avoiding deadlock in mind
type ResilientLock struct {
	canOwn  chan bool // prefer chanel over runtime.Gosched()
	ownGoid uint64
}

func NewResilientLock() *ResilientLock {
	r := &ResilientLock{canOwn: make(chan bool, 1)}
	r.canOwn <- true
	return r
}

func (r *ResilientLock) TryLockWithContext(c context.Context) bool {
	return r._internalLock(c)
}

func (r *ResilientLock) TryLockWithTime(t <-chan time.Time) bool {
	return r._internalLock(t)
}

func (r *ResilientLock) _internalLock(done any) bool {
	_, goid := Goid.GetGoid()

	for {
		if atomic.LoadUint64(&r.ownGoid) == goid {
			return true
		}
		select {
		case <-r.canOwn:
			atomic.StoreUint64(&r.ownGoid, goid)
			return true
		default:
			switch done := done.(type) {
			case <-chan time.Time:
				select {
				case <-done:
					return false
				default:
					continue
				}
			case context.Context:
				select {
				case <-done.Done():
					return false
				default:
					continue
				}
			}
		}
	}

}

func (r *ResilientLock) Unlock() {
	_, goid := Goid.GetGoid()
	if atomic.CompareAndSwapUint64(&r.ownGoid, goid, 0) {
		r.canOwn <- true
	}
}
