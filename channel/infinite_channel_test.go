package channel

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestSendRecv(t *testing.T) {
	ch := New[int](context.Background())
	ch.In() <- 42
	v := <-ch.Out()
	if v != 42 {
		t.Fatalf("got %d, want 42", v)
	}
	ch.Close()
	<-ch.Done()
}

func TestSendRecvMultiple(t *testing.T) {
	ch := New[int](context.Background())
	const n = 1000
	for i := 0; i < n; i++ {
		ch.In() <- i
	}
	for i := 0; i < n; i++ {
		v := <-ch.Out()
		if v != i {
			t.Fatalf("at %d: got %d, want %d", i, v, i)
		}
	}
	ch.Close()
	<-ch.Done()
}

func TestCloseEmpty(t *testing.T) {
	ch := New[int](context.Background())
	ch.Close()
	_, ok := <-ch.Out()
	if ok {
		t.Fatal("expected Out to be closed")
	}
	<-ch.Done()
}

func TestCloseDrains(t *testing.T) {
	ch := New[int](context.Background())
	ch.In() <- 1
	ch.In() <- 2
	ch.In() <- 3
	ch.Close()

	got := []int{}
	for v := range ch.Out() {
		got = append(got, v)
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("got %v, want [1 2 3]", got)
	}
	<-ch.Done()
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := New[int](ctx)
	ch.In() <- 1
	ch.In() <- 2
	ch.In() <- 3
	cancel()

	_, ok := <-ch.Out()
	if ok {
		t.Fatal("expected Out to be closed immediately")
	}
	<-ch.Done()
}

func TestLen(t *testing.T) {
	ch := New[int](context.Background())
	if l := ch.Len(); l != 0 {
		t.Fatalf("expected 0, got %d", l)
	}
	ch.In() <- 1
	waitLen(t, ch, 1)
	ch.In() <- 2
	waitLen(t, ch, 2)
	<-ch.Out()
	waitLen(t, ch, 1)
	<-ch.Out()
	waitLen(t, ch, 0)
	ch.Close()
	<-ch.Done()
}

func waitLen[T any](t *testing.T, ch *InfiniteChannel[T], want int) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		if ch.Len() == want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("waitLen(%d) timed out, got %d", want, ch.Len())
		default:
			runtime.Gosched()
		}
	}
}

func TestConcurrentSenders(t *testing.T) {
	ch := New[int](context.Background())
	const n = 1000
	const senders = 10

	var wg sync.WaitGroup
	for i := 0; i < senders; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < n; j++ {
				ch.In() <- base*n + j
			}
		}(i)
	}

	go func() {
		wg.Wait()
		ch.Close()
	}()

	count := 0
	for range ch.Out() {
		count++
	}
	if count != n*senders {
		t.Fatalf("expected %d items, got %d", n*senders, count)
	}
	<-ch.Done()
}

func TestConcurrentReceivers(t *testing.T) {
	ch := New[int](context.Background())
	const n = 10000

	go func() {
		for i := 0; i < n; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()

	var mu sync.Mutex
	got := map[int]bool{}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch.Out() {
				mu.Lock()
				got[v] = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(got) != n {
		t.Fatalf("expected %d unique items, got %d", n, len(got))
	}
	<-ch.Done()
}

func TestNonBlockingSend(t *testing.T) {
	ch := New[int](context.Background())
	done := make(chan struct{})
	go func() {
		ch.In() <- 42
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("send blocked")
	}
	v := <-ch.Out()
	if v != 42 {
		t.Fatalf("got %d, want 42", v)
	}
	ch.Close()
	<-ch.Done()
}

func TestNonBlockingRecv(t *testing.T) {
	ch := New[int](context.Background())
	select {
	case <-ch.Out():
		t.Fatal("recv should block on empty channel")
	default:
	}
	ch.Close()
	<-ch.Done()
}

func TestStringType(t *testing.T) {
	ch := New[string](context.Background())
	ch.In() <- "hello"
	ch.In() <- "world"
	ch.Close()

	got := []string{}
	for v := range ch.Out() {
		got = append(got, v)
	}
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Fatalf("got %v, want [hello world]", got)
	}
	<-ch.Done()
}

func TestStructType(t *testing.T) {
	type Point struct{ X, Y int }
	ch := New[Point](context.Background())
	ch.In() <- Point{1, 2}
	ch.In() <- Point{3, 4}
	ch.Close()

	p1 := <-ch.Out()
	p2 := <-ch.Out()
	if p1 != (Point{1, 2}) || p2 != (Point{3, 4}) {
		t.Fatalf("got %v %v", p1, p2)
	}
	<-ch.Done()
}

func TestLargeVolume(t *testing.T) {
	ch := New[int](context.Background())
	const n = 500000

	go func() {
		for i := 0; i < n; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()

	prev := -1
	for v := range ch.Out() {
		if v != prev+1 {
			t.Fatalf("out of order at %d: got %d, expected %d", prev+1, v, prev+1)
		}
		prev = v
	}
	if prev != n-1 {
		t.Fatalf("expected %d, got %d", n-1, prev)
	}
	<-ch.Done()
}

func TestDoneAfterClose(t *testing.T) {
	ch := New[int](context.Background())
	ch.Close()
	select {
	case <-ch.Done():
	case <-time.After(time.Second):
		t.Fatal("Done not closed after Out closed")
	}
}

func TestGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		ch := New[int](context.Background())
		ch.Close()
		<-ch.Done()
	}
	after := runtime.NumGoroutine()
	// allow some slack for GC
	if after-before > 5 {
		t.Fatalf("possible goroutine leak: %d -> %d", before, after)
	}
}

func TestSlowConsumer(t *testing.T) {
	ch := New[int](context.Background())
	const n = 10000

	go func() {
		for i := 0; i < n; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()

	for i := 0; i < n; i++ {
		v := <-ch.Out()
		if v != i {
			t.Fatalf("got %d, want %d", v, i)
		}
		// simulate slow consumer by yielding
		runtime.Gosched()
	}
	<-ch.Done()
}

func TestFastProducer(t *testing.T) {
	ch := New[int](context.Background())
	const n = 100000

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()

	count := 0
	for range ch.Out() {
		count++
	}
	if count != n {
		t.Fatalf("expected %d, got %d", n, count)
	}
	wg.Wait()
	<-ch.Done()
}

func TestDoubleClose(t *testing.T) {
	ch := New[int](context.Background())
	ch.Close()
	// sync.Once makes Close safe to call multiple times; must not panic.
	ch.Close()
	<-ch.Done()
}

func TestFIFOOrder(t *testing.T) {
	ch := New[int](context.Background())
	for i := 0; i < 1000; i++ {
		ch.In() <- i
	}
	for i := 0; i < 1000; i++ {
		v := <-ch.Out()
		if v != i {
			t.Fatalf("FIFO violation at %d: got %d", i, v)
		}
	}
	ch.Close()
	<-ch.Done()
}

func TestSendAfterClose(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on send to closed channel")
		}
	}()
	ch := New[int](context.Background())
	ch.Close()
	ch.In() <- 1
}

func TestContextShutdownDiscards(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := New[int](ctx)
	ch.In() <- 1
	ch.In() <- 2
	cancel()

	time.Sleep(10 * time.Millisecond)
	// Out should be closed and items discarded
	select {
	case _, ok := <-ch.Out():
		if ok {
			t.Fatal("expected Out closed after cancel")
		}
	default:
	}
	<-ch.Done()
}

func TestParallelSends(t *testing.T) {
	ch := New[int](context.Background())
	const goroutines = 20
	const sendsPerGoroutine = 5000

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < sendsPerGoroutine; j++ {
				ch.In() <- base*sendsPerGoroutine + j
			}
		}(i)
	}

	go func() {
		wg.Wait()
		ch.Close()
	}()

	seen := make(map[int]bool)
	for v := range ch.Out() {
		if seen[v] {
			t.Fatalf("duplicate value %d", v)
		}
		seen[v] = true
	}
	if len(seen) != goroutines*sendsPerGoroutine {
		t.Fatalf("expected %d unique items, got %d", goroutines*sendsPerGoroutine, len(seen))
	}
	<-ch.Done()
}

func TestImmediateClose(t *testing.T) {
	for i := 0; i < 100; i++ {
		ch := New[int](context.Background())
		ch.Close()
		<-ch.Done()
	}
}

func TestInterleavedSendRecv(t *testing.T) {
	ch := New[int](context.Background())
	for i := 0; i < 1000; i++ {
		ch.In() <- i
		v := <-ch.Out()
		if v != i {
			t.Fatalf("at %d: got %d", i, v)
		}
	}
	ch.Close()
	<-ch.Done()
}

// Benchmarks

func BenchmarkSendRecv(b *testing.B) {
	ch := New[int](context.Background())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch.In() <- 42
			<-ch.Out()
		}
	})
	ch.Close()
	<-ch.Done()
}

func BenchmarkSend(b *testing.B) {
	ch := New[int](context.Background())
	done := make(chan struct{})
	go func() {
		for range ch.Out() {
		}
		close(done)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	ch.Close()
	<-done
	<-ch.Done()
}

func BenchmarkRecv(b *testing.B) {
	ch := New[int](context.Background())
	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for range ch.Out() {
	}
	<-ch.Done()
}

func BenchmarkBatchedSend(b *testing.B) {
	ch := New[int](context.Background())
	done := make(chan struct{})
	go func() {
		for range ch.Out() {
		}
		close(done)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	ch.Close()
	<-done
	<-ch.Done()
}

func BenchmarkLargeBatch(b *testing.B) {
	ch := New[int](context.Background())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	b.StopTimer()
	ch.Close()
	for range ch.Out() {
	}
	<-ch.Done()
}

func BenchmarkParallelSendRecv4(b *testing.B) {
	ch := New[int](context.Background())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch.In() <- 42
			<-ch.Out()
		}
	})
	ch.Close()
	<-ch.Done()
}

func BenchmarkParallelSendRecv8(b *testing.B) {
	ch := New[int](context.Background())
	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch.In() <- 42
			<-ch.Out()
		}
	})
	ch.Close()
	<-ch.Done()
}

func BenchmarkThroughput(b *testing.B) {
	ch := New[int](context.Background())
	b.ResetTimer()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ch.Out() {
		}
	}()
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	ch.Close()
	wg.Wait()
	<-ch.Done()
}

func BenchmarkChanSendRecv(b *testing.B) {
	ch := make(chan int, b.N)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- 42
			<-ch
		}
	})
	close(ch)
}

func BenchmarkChanSend(b *testing.B) {
	ch := make(chan int, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	close(ch)
}

func BenchmarkChanRecv(b *testing.B) {
	ch := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	close(ch)
	b.ResetTimer()
	for range ch {
	}
}

func BenchmarkChanBatchedSend(b *testing.B) {
	ch := make(chan int, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	b.StopTimer()
	close(ch)
}

func BenchmarkChanParallelSendRecv4(b *testing.B) {
	ch := make(chan int, b.N)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- 42
			<-ch
		}
	})
	close(ch)
}

func BenchmarkChanParallelSendRecv8(b *testing.B) {
	ch := make(chan int, b.N)
	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- 42
			<-ch
		}
	})
	close(ch)
}
