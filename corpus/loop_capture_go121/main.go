package main

import (
	"fmt"
	"sync"
)

// BuggyLoop: under Go < 1.22 semantics, every goroutine captures the
// SAME variable i/v, not a per-iteration copy. All goroutines are
// likely to print the final values, not each distinct value.
func BuggyLoop(items []int) {
	var wg sync.WaitGroup
	for i, v := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(i, v) // BUG under go1.21: captures loop var by reference
		}()
	}
	wg.Wait()
}

func main() {
	BuggyLoop([]int{1, 2, 3, 4, 5})
}
