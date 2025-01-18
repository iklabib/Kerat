package memo

import (
	"log"
	"sync"
	"time"

	"codeberg.org/iklabib/kerat/processor/toolchains"
)

type BoxCaches struct {
	intervarl int
	mu        sync.Mutex
	timers    map[string]time.Timer
	exercises map[string]toolchains.Toolchain
}

func NewBoxCaches(interval int) BoxCaches {
	return BoxCaches{
		intervarl: interval,
		mu:        sync.Mutex{},
		timers:    make(map[string]time.Timer),
		exercises: make(map[string]toolchains.Toolchain),
	}
}

func (b *BoxCaches) LoadToolchain(id string) (toolchains.Toolchain, bool) {
	b.mu.Lock()

	tc, ok := b.exercises[id]
	if timer, exists := b.timers[id]; exists {
		timer.Stop()
		intervalTime := time.Duration(b.intervarl) * time.Minute
		timer = *time.NewTimer(intervalTime)
	}
	b.mu.Unlock()
	b.timers[id] = *b.CleanTimer(id)

	return tc, ok
}

func (b *BoxCaches) AddToolchain(id string, tc toolchains.Toolchain) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.exercises[id] = tc
	b.timers[id] = *b.CleanTimer(id)
}

func (b *BoxCaches) CleanTimer(id string) *time.Timer {
	intervalTime := time.Duration(b.intervarl) * time.Minute
	return time.AfterFunc(intervalTime, func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if tc, ok := b.exercises[id]; !ok {
			return
		} else {
			tc.Clean()
		}

		delete(b.exercises, id)
		delete(b.timers, id)
		log.Printf("Cleaned up entry: %s", id)
	})
}
