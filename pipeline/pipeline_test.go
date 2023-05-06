package pipeline

import (
	"brick/channel"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	pipe := New(func(values []string) {
		t.Log(strings.Join(values, ";"))
		t.Log(strings.Repeat("+", 100))
	})
	defer pipe.Close()
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
					channel.SendTimeout(errorCh, false, 5)
				}
			}, LimitReportMode[string](), SetLimit[string](limit))
			defer pipe.Close()
			pipe.Start()
			for j := 0; j < 100; j++ {
				go pipe.Push(fmt.Sprintf("pipeline-%v-%v", limit, j))
			}
			time.Sleep(time.Second * 10)
		}(i)
	}

	if _, ok := channel.AcceptTimeout(errorCh, 10); ok {
		t.Fatal("limit error")
	} else {
		t.Log("successful")
	}
}

func TestIntervalMode(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, IntervalReportMode[string](), SetInterval[string](5))
	defer pipe.Close()
	pipe.Start()
	for j := 0; j < 100; j++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", j))
		time.Sleep(time.Second)
	}
}

func TestClose(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, LimitReportMode[string](), SetLimit[string](1))
	pipe.Start()
	for j := 0; j < 10000; j++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", j))
	}
	time.Sleep(time.Millisecond * 5)
	pipe.Close()
	for j := 1000; j < 2000; j++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", j))
	}
	time.Sleep(time.Second)
}

func TestRepeatClose(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, LimitReportMode[string](), SetLimit[string](1))
	pipe.Start()
	for j := 0; j < 100; j++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", j))
	}
	time.Sleep(time.Second)

	for j := 0; j < 100; j++ {
		go pipe.Close()
	}
	time.Sleep(time.Second)
}

func TestRepeatStart(t *testing.T) {
	pipe := New(func(values []string) {
		log.Println(strings.Join(values, ";"))
	}, LimitReportMode[string](), SetLimit[string](1))
	for j := 0; j < 100; j++ {
		go pipe.Start()
	}
	time.Sleep(time.Second)
	for j := 0; j < 100; j++ {
		go pipe.Push(fmt.Sprintf("pipeline-%v", j))
	}
	time.Sleep(time.Second)
}
