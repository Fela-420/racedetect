package main

import (
	"fmt"
	"sync"
)

// SafeUnderNewSemantics: syntactically IDENTICAL to the go1.21 buggy
// sample. Under go1.22+ semantics, i and v are per-iteration, so this
// is safe. Your detector must NOT flag this the same way as the
// loop_capture_go121 version.
func SafeUnderNewSemantics(items []int) {
	var wg sync.WaitGroup
	for i, v := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(i, v) // safe under go1.22+
		}()
	}
	wg.Wait()
}

func main() {
	SafeUnderNewSemantics([]int{1, 2, 3, 4, 5})
}
