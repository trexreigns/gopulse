package pool_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/trexreigns/gopulse/pool"
)

func TestNewPool(t *testing.T) {
	p := pool.NewPool(2, 10)
	if p == nil {
		t.Fatal("NewPool should not return nil")
	}

	// Clean up
	p.Stop()
}

func TestPoolBasicJobExecution(t *testing.T) {
	p := pool.NewPool(2, 10)
	p.StartWorkers()
	defer p.Stop()

	var executed int32
	job := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Submit job
	if !p.Submit(job) {
		t.Fatal("Should be able to submit job")
	}

	// Wait for job to execute
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("Expected 1 job executed, got %d", executed)
	}
}

func TestPoolMultipleJobs(t *testing.T) {
	p := pool.NewPool(3, 20)
	p.StartWorkers()
	defer p.Stop()

	var executed int32
	numJobs := 10

	for i := 0; i < numJobs; i++ {
		job := func() {
			atomic.AddInt32(&executed, 1)
		}

		if !p.Submit(job) {
			t.Errorf("Should be able to submit job %d", i)
		}
	}

	// Wait for all jobs to execute
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&executed) != int32(numJobs) {
		t.Errorf("Expected %d jobs executed, got %d", numJobs, executed)
	}
}

func TestPoolConcurrentExecution(t *testing.T) {
	p := pool.NewPool(3, 10)
	p.StartWorkers()
	defer p.Stop()

	var concurrentCount int32
	var maxConcurrent int32
	var wg sync.WaitGroup

	numJobs := 6
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		job := func() {
			defer wg.Done()

			// Increment concurrent count
			current := atomic.AddInt32(&concurrentCount, 1)

			// Track max concurrent
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			}

			// Simulate work
			time.Sleep(100 * time.Millisecond)

			// Decrement concurrent count
			atomic.AddInt32(&concurrentCount, -1)
		}

		if !p.Submit(job) {
			t.Errorf("Should be able to submit job %d", i)
		}
	}

	// Wait for all jobs to complete
	wg.Wait()

	// With 3 workers, max concurrent should be 3
	if atomic.LoadInt32(&maxConcurrent) != 3 {
		t.Errorf("Expected max concurrent to be 3, got %d", maxConcurrent)
	}
}

func TestPoolCapacityLimit(t *testing.T) {
	bufferSize := 2
	p := pool.NewPool(1, bufferSize)
	p.StartWorkers()
	defer p.Stop()

	var executed int32
	blockingJob := func() {
		atomic.AddInt32(&executed, 1)
		time.Sleep(200 * time.Millisecond) // Block worker
	}

	// Fill the buffer and worker
	successfulSubmissions := 0
	for i := 0; i < bufferSize+5; i++ {
		if p.Submit(blockingJob) {
			successfulSubmissions++
		}
	}

	// Should be able to submit at least buffer size jobs
	if successfulSubmissions < bufferSize {
		t.Errorf("Expected at least %d successful submissions, got %d", bufferSize, successfulSubmissions)
	}

	// Wait for jobs to complete
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&executed) != int32(successfulSubmissions) {
		t.Errorf("Expected %d jobs executed, got %d", successfulSubmissions, executed)
	}
}

func TestPoolSubmitAfterStop(t *testing.T) {
	p := pool.NewPool(2, 10)
	p.StartWorkers()

	// Stop the pool
	p.Stop()

	// Try to submit after stopping
	job := func() {}
	if p.Submit(job) {
		t.Error("Should not be able to submit job after stopping")
	}
}

func TestPoolShutdownWaitsForJobs(t *testing.T) {
	p := pool.NewPool(2, 10)
	p.StartWorkers()

	var executed int32
	var startTime time.Time
	var endTime time.Time

	// Submit jobs that take time
	for i := 0; i < 3; i++ {
		job := func() {
			if atomic.LoadInt32(&executed) == 0 {
				startTime = time.Now()
			}
			atomic.AddInt32(&executed, 1)
			time.Sleep(100 * time.Millisecond)
			endTime = time.Now()
		}

		if !p.Submit(job) {
			t.Errorf("Should be able to submit job %d", i)
		}
	}

	// Give jobs a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop should wait for all jobs to complete
	p.Stop()

	// All jobs should have executed
	if atomic.LoadInt32(&executed) != 3 {
		t.Errorf("Expected 3 jobs executed, got %d", executed)
	}

	// Should have taken at least some time
	if endTime.Sub(startTime) < 50*time.Millisecond {
		t.Error("Jobs should have taken some time to execute")
	}
}

func TestPoolJobPanic(t *testing.T) {
	p := pool.NewPool(2, 10)
	p.StartWorkers()
	defer p.Stop()

	var executed int32

	// Submit a job that panics
	panicJob := func() {
		panic("test panic")
	}

	normalJob := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Submit panic job
	if !p.Submit(panicJob) {
		t.Error("Should be able to submit panic job")
	}

	// Submit normal job after panic
	if !p.Submit(normalJob) {
		t.Error("Should be able to submit normal job after panic")
	}

	// Wait for jobs to process
	time.Sleep(100 * time.Millisecond)

	// Normal job should still execute even after panic
	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("Expected 1 normal job executed, got %d", executed)
	}
}

func TestPoolStressTest(t *testing.T) {
	p := pool.NewPool(5, 50)
	p.StartWorkers()
	defer p.Stop()

	var executed int32
	numJobs := 100

	// Submit many jobs quickly
	for i := 0; i < numJobs; i++ {
		job := func() {
			atomic.AddInt32(&executed, 1)
		}

		// Some jobs might be rejected if buffer is full
		p.Submit(job)
	}

	// Wait for all submitted jobs to execute
	time.Sleep(500 * time.Millisecond)

	executed_count := atomic.LoadInt32(&executed)
	if executed_count == 0 {
		t.Error("At least some jobs should have been executed")
	}

	t.Logf("Executed %d out of %d jobs", executed_count, numJobs)
}

func TestPoolWorkerCount(t *testing.T) {
	workers := 4
	p := pool.NewPool(workers, 10)
	p.StartWorkers()
	defer p.Stop()

	var concurrentCount int32
	var maxConcurrent int32
	var wg sync.WaitGroup

	// Submit more jobs than workers
	numJobs := workers * 2
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		job := func() {
			defer wg.Done()

			current := atomic.AddInt32(&concurrentCount, 1)

			// Track max concurrent
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			}

			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&concurrentCount, -1)
		}

		if !p.Submit(job) {
			t.Errorf("Should be able to submit job %d", i)
		}
	}

	wg.Wait()

	// Max concurrent should not exceed worker count
	if atomic.LoadInt32(&maxConcurrent) != int32(workers) {
		t.Errorf("Expected max concurrent to be %d, got %d", workers, maxConcurrent)
	}
}
