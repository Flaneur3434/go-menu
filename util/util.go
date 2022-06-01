package util

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

type Rank struct {
	Word     string
	distance float64
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

// TODO: this is really slow, need to parallelism somehow
func match(s, t string) float64 {
	sidx, eidx := -1, -1

	for i, j := 0, 0; i < len(t); i++ {
		if string(s[j]) == string(t[i]) {
			if sidx == -1 {
				sidx = i
			}
			j++
			if j == len(s) {
				eidx = i
				break
			}
		}
	}

	// pattern s was contained in string t
	if eidx != -1 {
		/* compute distance */
		/* add penalty if match starts late (log(sidx+2))
		 * add penalty for long a match without many matching characters */
		return math.Log(float64(sidx)+2) + float64(eidx-sidx-len(s))
	} else {
		return 100
	}
}

func (R Ranks) FuzzySearch(target string) {
	if len(target) < 1 {
		return
	}

	for i := range R {
		R[i].distance = match(target, R[i].Word)
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
