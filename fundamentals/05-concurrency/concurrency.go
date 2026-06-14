package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// --- Базові goroutines + WaitGroup ---

func basicGoroutines() {
	fmt.Println("=== Goroutines + WaitGroup ===")
	var wg sync.WaitGroup

	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  worker %d done\n", id)
		}(i)
	}

	wg.Wait()
	fmt.Println("  всі workers завершились")
}

// --- Channels: небуферизований ---

func unbufferedChannel() {
	fmt.Println("\n=== Unbuffered Channel ===")
	ch := make(chan string)

	go func() {
		ch <- "ping" // блокує до отримання
	}()

	msg := <-ch // блокує до відправки
	fmt.Println(" ", msg)
}

// --- Channels: буферизований ---

func bufferedChannel() {
	fmt.Println("\n=== Buffered Channel ===")
	ch := make(chan int, 3) // не блокує поки є місце

	ch <- 1
	ch <- 2
	ch <- 3
	// ch <- 4 // заблокує — буфер повний

	fmt.Printf("  len=%d, cap=%d\n", len(ch), cap(ch))
	fmt.Println(" ", <-ch, <-ch, <-ch)
}

// --- range по каналу ---

func rangeChannel() {
	fmt.Println("\n=== Range over Channel ===")
	ch := make(chan int, 5)

	go func() {
		for i := range 5 {
			ch <- i * i
		}
		close(ch) // обов'язково закрити, інакше range зависне
	}()

	for v := range ch {
		fmt.Printf("  %d\n", v)
	}
}

// --- select ---

func selectDemo() {
	fmt.Println("\n=== Select ===")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "one"
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch2 <- "two"
	}()

	for range 2 {
		select {
		case msg := <-ch1:
			fmt.Println("  ch1:", msg)
		case msg := <-ch2:
			fmt.Println("  ch2:", msg)
		}
	}
}

// --- select з default (non-blocking) ---

func nonBlockingSelect() {
	fmt.Println("\n=== Non-blocking Select ===")
	ch := make(chan int, 1)

	select {
	case v := <-ch:
		fmt.Println("  got:", v)
	default:
		fmt.Println("  канал порожній, не блокуємо")
	}
}

// --- Mutex ---

type SafeCounter struct {
	mu sync.Mutex
	v  int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v
}

func mutexDemo() {
	fmt.Println("\n=== Mutex ===")
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}

	wg.Wait()
	fmt.Printf("  counter = %d (завжди 1000)\n", counter.Value())
}

// --- sync.Once ---

var (
	once     sync.Once
	resource string
)

func getResource() string {
	once.Do(func() {
		fmt.Println("  ініціалізація ресурсу (тільки один раз)")
		resource = "initialized"
	})
	return resource
}

func onceDemo() {
	fmt.Println("\n=== sync.Once ===")
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			getResource()
		}()
	}
	wg.Wait()
	fmt.Println(" ", getResource())
}

// --- Worker Pool ---

func workerPool() {
	fmt.Println("\n=== Worker Pool ===")
	const numWorkers = 3
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	// Запустити workers
	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range jobs {
				results <- j * j
				fmt.Printf("  worker %d обробив job %d\n", id, j)
			}
		}(w)
	}

	// Відправити jobs
	for j := range 6 {
		jobs <- j + 1
	}
	close(jobs)

	// Закрити results коли всі workers завершаться
	go func() {
		wg.Wait()
		close(results)
	}()

	// Зібрати результати
	var sum int
	for r := range results {
		sum += r
	}
	fmt.Printf("  sum of squares: %d\n", sum)
}

// --- Context: таймаут ---

func contextTimeout() {
	fmt.Println("\n=== Context Timeout ===")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := make(chan string, 1)
	go func() {
		time.Sleep(200 * time.Millisecond) // повільніше за таймаут
		ch <- "result"
	}()

	select {
	case result := <-ch:
		fmt.Println("  result:", result)
	case <-ctx.Done():
		fmt.Println("  timeout:", ctx.Err())
	}
}

// --- Context: скасування ---

func contextCancel() {
	fmt.Println("\n=== Context Cancel ===")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("  done (timeout)")
	case <-ctx.Done():
		fmt.Println("  cancelled:", ctx.Err())
	}
}

func main() {
	basicGoroutines()
	unbufferedChannel()
	bufferedChannel()
	rangeChannel()
	selectDemo()
	nonBlockingSelect()
	mutexDemo()
	onceDemo()
	workerPool()
	contextTimeout()
	contextCancel()
}
