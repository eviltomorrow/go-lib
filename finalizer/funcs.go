package finalizer

import (
	"sync"
)

var (
	cleanupFuncs []NamedFunc
	mut          sync.Mutex
)

type NamedFunc struct {
	Name string
	Fn   func() error
}

func RegisterCleanupFuncs(f ...NamedFunc) {
	mut.Lock()
	defer mut.Unlock()

	if f != nil {
		cleanupFuncs = append(cleanupFuncs, f...)
	}
}

func RunCleanupFuncs() []error {
	mut.Lock()
	defer mut.Unlock()

	e := make([]error, 0, len(cleanupFuncs))
	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		f := cleanupFuncs[i]
		if f.Fn != nil {
			if err := f.Fn(); err != nil {
				e = append(e, err)
			}
		}
	}
	return e
}
