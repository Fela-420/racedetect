package clean

import "sync"

// SafeLoop demonstrates the pre-1.22 idiom for avoiding loop variable
// capture: the variable is explicitly re-declared/passed per iteration.
func SafeLoop(items []int) []int {
	var wg sync.WaitGroup
	results := make([]int, len(items))

	for i, v := range items {
		wg.Add(1)
		go func(idx int, val int) {
			defer wg.Done()
			results[idx] = val * 2
		}(i, v)
	}
	wg.Wait()
	return results
}
