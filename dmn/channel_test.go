package dmn

import (
	"fmt"
	"testing"
)

var total int

var runningRunes []rune

// This test tries to find issues in modifying global variables within channels.
// There is no issue at this point.
func TestCount(t *testing.T) {

	wordCounter := make(chan int)

	f := func(s string) {
		total := 0
		for _, c := range s {
			fmt.Printf("Encountered %v\n", string(c))
			runningRunes = append(runningRunes, c)
			total = total + 1
		}
		wordCounter <- total
	}
	go f("hello world")

	count := <-wordCounter

	fmt.Printf("Counted %v characters\n", count)

	for _, c := range runningRunes {
		fmt.Printf("Running rune: %v\n", string(c))
	}
}
