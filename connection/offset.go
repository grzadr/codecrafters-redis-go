package connection

import "sync"

type OffsetTracker struct {
	bytes int
	mutex sync.Mutex
}

func NewOffsetTracker() *OffsetTracker {
	return &OffsetTracker{}
}

func (o *OffsetTracker) Add(b int) (current int) {
	current = o.bytes
	o.bytes += b

	return
}

func (o *OffsetTracker) Current() int {
	return o.bytes
}

var (
	tracker     *OffsetTracker
	trackerOnce sync.Once
)

func GetOffsetTracker() *OffsetTracker {
	trackerOnce.Do(func() {
		tracker = NewOffsetTracker()
	})

	return tracker
}
