package pipeline

import (
	"fmt"
	"github.com/fanyiguang/brick/channel"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	pipe := New(func(values []string) {
		t.Log(strings.Join(values, ";"))
		t.Log(strings.Repeat("+", 100))
	})
	defer pipe.Close(2)
	pipe.Start()
	for i := 0; i < 100; i++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", i))
	}
	time.Sleep(time.Second * 10)
}

func TestLimitMode(t *testing.T) {
	errorCh := make(chan bool)
	for i := 1; i < 10; i++ {
		go func(limit int) {
			pipe := New(func(values []string) {
				log.Println("limit: ", limit, "values len: ", len(values))
				if len(values) > limit {
					_ = channel.SendTimeout(errorCh, false, 5)
				}
			}, LimitReportMode[string](), SetLimit[string](limit))
			defer pipe.Close(2)
			pipe.Start()
			for j := 0; j < 100; j++ {
				go pipe.Push(fmt.Sprintf("pipeline-%v-%v", limit, j))
			}
			time.Sleep(time.Second * 10)
		}(i)
	}

	if _, err := channel.AcceptTimeout(errorCh, 10); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Log("successful")
	}
}

func TestIntervalMode(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, IntervalReportMode[string](), SetInterval[string](5))
	defer pipe.Close(2)
	pipe.Start()
	for j := 0; j < 100; j++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", j))
		time.Sleep(time.Second)
	}
}

func TestClose(t *testing.T) {
	val := sync.Map{}
	errCh := make(chan int, 1)
	pipe := New(func(values []int) {
		for _, value := range values {
			if _, loaded := val.LoadOrStore(value, struct{}{}); loaded {
				channel.NonBlockSend(errCh, value)
			}
		}
	}, LimitReportMode[int](), SetLimit[int](1))
	pipe.Start()
	for j := 3000; j < 4000; j++ {
		go pipe.Push(j)
	}
	pipe.Close(5)
	var wg sync.WaitGroup
	wg.Add(1000)
	for j := 1000; j < 2000; j++ {
		go func() {
			pipe.Push(j)
			wg.Done()
		}()
	}
	wg.Wait()
	val.Range(func(key, value any) bool {
		if key.(int) >= 1000 && key.(int) < 2000 {
			t.Fatal("error", key)
			return false
		}
		return true
	})
	if accept, err := channel.NonBlockAccept(errCh); err == nil {
		t.Fatal("error", accept)
	}
}

func TestRepeatClose(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, LimitReportMode[string](), SetLimit[string](1))
	pipe.Start()
	var wg2 sync.WaitGroup
	wg2.Add(100)
	for j := 0; j < 100; j++ {
		go func(j int) {
			pipe.Push(fmt.Sprintf("pipeline-%v", j))
			wg2.Done()
		}(j)
	}
	wg2.Wait()

	var wg sync.WaitGroup
	wg.Add(100)
	for j := 0; j < 100; j++ {
		go func() {
			pipe.Close(2)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestRepeatStart(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, LimitReportMode[string](), SetLimit[string](1))
	defer pipe.Close(2)
	pipe.Start()
	var wg1 sync.WaitGroup
	wg1.Add(100)
	for j := 0; j < 100; j++ {
		go func() {
			pipe.Start()
			wg1.Done()
		}()
	}

	var wg2 sync.WaitGroup
	wg2.Add(100)
	for j := 0; j < 100; j++ {
		go func(j int) {
			pipe.Push(fmt.Sprintf("pipeline-%v", j))
			wg2.Done()
		}(j)
	}
	wg1.Wait()
	wg2.Wait()
}
