package user_thread_pool

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func GetGoid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return int64(id)
}

func TestSubmit(t *testing.T) {
	group := sync.WaitGroup{}
	group.Add(1000000)
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		cur := i
		Submit(func() {
			t.Logf("run %d\n", cur)
			group.Done()
		})
	}
	group.Wait()
	t.Logf("end start in %v", time.Now().Sub(start))
}

func TestSubmit1(t *testing.T) {
	group := sync.WaitGroup{}
	group.Add(1000000)
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		cur := i
		go func() {
			t.Logf("run %d\n", cur)
			group.Done()
		}()
	}
	group.Wait()
	t.Logf("end start in %v", time.Now().Sub(start))
}
