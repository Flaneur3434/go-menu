package util

import (
	"bufio"
	"fmt"
	_ "github.com/Flaneur3434/go-menu/draw"
	"os"
)

func ReadStdIn(c chan string) {
	scanner := bufio.NewScanner(os.Stdin)

	// store stdin into 'input' slice
	for scanner.Scan() {
		c <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading err: ", err)
	}

	close(c)
}

// func (m *draw.Menu) fuzzySearch(out chan string) {

// }
