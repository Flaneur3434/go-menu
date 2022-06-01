package util

import (
	"bufio"
	"fmt"
	"os"
)

type Rank struct {
	Word     string
	distance int
}

type Ranks []Rank

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

func InitRanks(s []string) (R Ranks) {
	for i := range s {
		R = append(R, Rank{Word: s[i], distance: -1})
	}

	return
}

func levenshtein(s1 string, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)

	column := make([]int, 1, 64)

	for y := 1; y <= len(r1); y++ {
		column = append(column, y)
	}

	for x := 1; x <= len(r2); x++ {
		column[0] = x

		for y, lastDiag := 1, x-1; y <= len(r1); y++ {
			oldDiag := column[y]
			cost := 0
			if r1[y-1] != r2[x-1] {
				cost = 1
			}
			column[y] = min(column[y]+1, column[y-1]+1, lastDiag+cost)
			lastDiag = oldDiag
		}
	}
	return column[len(r1)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}

func (R Ranks) FuzzySearch(target string) {
	if len(target) < 1 {
		return
	}

	for i := range R {
		R[i].distance = levenshtein(target, R[i].Word)
	}
}

func (R Ranks) Len() int {
	return len(R)
}

func (R Ranks) Less(i, j int) bool {
	return R[i].distance < R[j].distance
}

func (R Ranks) Swap(i, j int) {
	R[i], R[j] = R[j], R[i]
}
