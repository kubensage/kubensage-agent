package utils

import "sync"

func SafeGo(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}
