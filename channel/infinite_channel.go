package channel

import (
	"context"
	"sync"
	"sync/atomic"
)

// InfiniteChannel is an unbounded FIFO queue with channel-like send/receive
// semantics. Senders never block; receivers block when the queue is empty.
//
// Close signals end of input. After Close, any buffered items are still
// delivered to Out before it is closed.
//
// If the context passed to New is cancelled, Out is closed immediately and
// buffered items are discarded.
type InfiniteChannel[T any] struct {
	in     chan T
	out    chan T
	done   chan struct{}
	cancel context.CancelFunc
	close  sync.Once

	buf []T
	len atomic.Int64
}

// New creates an InfiniteChannel. Pass a background context for lifetime
// management; cancel it to perform an immediate shutdown.
func New[T any](ctx context.Context) *InfiniteChannel[T] {
	ctx, cancel := context.WithCancel(ctx)
	ch := &InfiniteChannel[T]{
		in:     make(chan T),
		out:    make(chan T),
		done:   make(chan struct{}),
		cancel: cancel,
		buf:    make([]T, 0, 64),
	}
	go ch.loop(ctx)
	return ch
}

// In returns the send-only channel. Never block the sender.
func (ch *InfiniteChannel[T]) In() chan<- T { return ch.in }

// Out returns the receive-only channel. Blocks when the queue is empty.
func (ch *InfiniteChannel[T]) Out() <-chan T { return ch.out }

// Done returns a channel that is closed when the run loop exits
// (after Out has been closed).
func (ch *InfiniteChannel[T]) Done() <-chan struct{} { return ch.done }

// Len returns the current number of buffered items.
func (ch *InfiniteChannel[T]) Len() int { return int(ch.len.Load()) }

// Close signals that no more items will be sent to In. After all buffered
// items have been received, Out is closed. Close panics if called more than
// once.
func (ch *InfiniteChannel[T]) Close() {
	ch.close.Do(func() { close(ch.in) })
}

func (ch *InfiniteChannel[T]) loop(ctx context.Context) {
	defer close(ch.done)
	defer close(ch.out)
	defer ch.cancel()

	input := ch.in
	var output chan T
	var cur T

	for {
		select {
		case <-ctx.Done():
			return

		case val, ok := <-input:
			if !ok {
				input = nil
				if len(ch.buf) == 0 {
					return
				}
				continue
			}
			ch.buf = append(ch.buf, val)
			ch.len.Store(int64(len(ch.buf)))

		case output <- cur:
			ch.buf = ch.buf[1:]
			n := len(ch.buf)
			ch.len.Store(int64(n))
			if n == 0 && input == nil {
				return
			}
		}

		if n := len(ch.buf); n > 0 {
			output = ch.out
			cur = ch.buf[0]
		} else {
			output = nil
		}
	}
}
