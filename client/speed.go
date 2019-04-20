package main

import (
	"fmt"
	"os"
	"time"
)

type Speed struct {
	total int64
	bytes []int64
	done  chan struct{}
	wait  chan struct{}
}

func NewSpeed() *Speed {
	s := &Speed{
		bytes: make([]int64, 10),
		done:  make(chan struct{}),
		wait:  make(chan struct{}),
	}
	go s.tick()
	return s
}

func (s *Speed) Write(p []byte) (int, error) {
	s.total += int64(len(p))
	s.bytes[time.Now().Unix()%int64(len(s.bytes))] += int64(len(p))
	return len(p), nil
}

func (s *Speed) Close() error {
	close(s.done)
	<-s.wait
	return nil
}

func (s *Speed) tick() {
	start := time.Now()

LOOP:
	for {
		select {
		case now := <-time.After(time.Second):
			n := int(now.Unix()-1) % len(s.bytes)
			s.Print(n)
			s.bytes[n] = 0
		case <-s.done:
			break LOOP
		}
	}

	spent := time.Now().Sub(start)
	fmt.Printf("Total send: %v, time usage: %v\n", Size(s.total), spent)
	close(s.wait)
}

func Size(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.3f KB", float64(b)/1024)
	case b < 1024*1024*1024:
		return fmt.Sprintf("%.3f MB", float64(b)/1024/1024)
	default:
		return fmt.Sprintf("%.3f GB", float64(b)/1024/1024/1024)
	}
}

func (s *Speed) Print(n int) {
	fmt.Fprintf(os.Stdout, "send: %v/s, total: %v      \r",
		Size(s.bytes[n]), Size(s.total))
}
