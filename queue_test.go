package user_thread_pool

import (
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestQueue(t *testing.T) {
	current := &node{
		Next:  nil,
		Value: nil,
	}
	queue_cas := DefaultTaskQueue{
		Head: unsafe.Pointer(current),
		Tail: unsafe.Pointer(current),
	}
	node := &Node{
		Next: nil,
		Val:  nil,
	}
	queue_two_lock := TwoLockQueue{
		HeadLock: &sync.Mutex{},
		TailLock: &sync.Mutex{},
		Head:     node,
		Tail:     node,
	}
	queue_one_lock := LockQueue{
		lock: &sync.Mutex{},
		q:    []interface{}{},
	}
	benchmarkOffer("cas", &queue_cas, t)
	benchmarkOffer("two_lock", &queue_two_lock, t)
	benchmarkOffer("one_lock", &queue_one_lock, t)
	benchmarkPoll("cas", &queue_cas, t)
	benchmarkPoll("two_lock", &queue_two_lock, t)
	benchmarkPoll("one_lock", &queue_one_lock, t)
}

func benchmarkOffer(queueName string, q Queue, t *testing.T) {
	wait := sync.WaitGroup{}
	wait.Add(10000)
	start := time.Now()
	for i := 0; i < 10000; i++ {
		go func() {
			q.Offer("test")
			wait.Done()
		}()
	}
	t.Logf("%s add 10000 item use %v\n", queueName, time.Now().Sub(start))

}

func benchmarkPoll(queueName string, q Queue, t *testing.T) {
	wait := sync.WaitGroup{}
	wait.Add(10000)
	start := time.Now()
	for i := 0; i < 10000; i++ {
		go func() {
			for {
				if q.Poll() == nil {
					continue
				}
				wait.Done()
				break
			}
		}()
	}
	t.Logf("%s poll 10000 item use %v\n", queueName, time.Now().Sub(start))

}
