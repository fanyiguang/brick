package timeralarm

import (
	"log"
	"testing"
	"time"
)

func TestTimerAlarmNotRepeated(t *testing.T) {
	alarm := New(func() {
		log.Printf("timeout\n")
	}, false)
	alarm.Start(1)
	log.Printf("start\n")
	go func() {
		for i := 0; i < 10; i++ {
			alarm.ResetTimer()
			time.Sleep(1 * time.Second)
		}
		log.Printf("reset end!\n")
	}()

	time.Sleep(15 * time.Second)
	log.Printf("close\n")
	alarm.Close()
	time.Sleep(10 * time.Second)
}

func TestTimerAlarmRepeated(t *testing.T) {
	alarm := New(func() {
		log.Printf("timeout\n")
	}, true)
	alarm.Start(1)
	log.Printf("start\n")
	go func() {
		for i := 0; i < 10; i++ {
			alarm.ResetTimer()
			time.Sleep(1 * time.Second)
		}
		log.Printf("reset end!\n")
	}()

	time.Sleep(15 * time.Second)
	log.Printf("close\n")
	alarm.Close()
	time.Sleep(10 * time.Second)
}

func TestTimerAlarmRepeatStart(t *testing.T) {
	alarm := New(func() {
		log.Printf("timeout\n")
	}, true)
	alarm.Start(1)
	alarm.Start(1)
	alarm.Start(1)
	alarm.Start(1)
	alarm.Start(1)
	time.Sleep(5 * time.Second)
	alarm.Close()
}

func TestTimerAlarmRepeatClose(t *testing.T) {
	alarm := New(func() {
		log.Printf("timeout\n")
	}, true)
	alarm.Start(1)
	time.Sleep(5 * time.Second)
	go alarm.Close()
	go alarm.Close()
	go alarm.Close()
	alarm.Close()
	alarm.Close()
	time.Sleep(5 * time.Second)
	t.Log("end")
}
