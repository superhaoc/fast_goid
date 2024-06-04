import (
"sync"
"sync/atomic"
)

type ReenterentLock struct {
sync.Mutex
OwnerGoid uint64
}

func (r *ReenterentLock) Lock() {
_, goid := Goid.GetGoid()
if atomic.CompareAndSwapUint64(&r.OwnerGoid, 0, goid) {
r.Mutex.Lock()
}

}

func (r *ReenterentLock) Unlock() {
_, goid := Goid.GetGoid()
if atomic.CompareAndSwapUint64(&r.OwnerGoid, goid, 0) {
r.Mutex.Unlock()
}

}