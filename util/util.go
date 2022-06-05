package util

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
)

const (
	numOfThreads = 10
)

type Rank struct {
	Word string
	Rank float64
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
		R = append(R, Rank{Word: s[i], Rank: -1})
	}

	return
}

func match(s, t string) float64 {
	sidx, eidx := -1, -1

	if len(s) < 1 {
		return -1
	}

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
		/* compute rank */
		/* add penalty if match starts late (log(sidx+2))
		 * add penalty for long a match without many matching characters */
		return math.Log(float64(sidx)+2) + float64(eidx-sidx-len(s))
	} else {
		return math.MaxFloat64
	}
}

func FuzzySearch(list []string, target string) Ranks {
	chunks := int(len(list) / numOfThreads)
	tailcaseChunk := len(list) % numOfThreads
	out := make(chan Ranks)

	for i := 0; i < numOfThreads; i++ {
		go func(threadN int, in chan Ranks) {
			var rankSlice Ranks
			for j := threadN * chunks; j < (threadN+1)*chunks; j++ {
				rankSlice = append(rankSlice, Rank{Word: list[j], Rank: match(target, list[j])})
			}

			in <- rankSlice
		}(i, out)
	}

	finalRanks := make(Ranks, 0, len(list))
	for len(finalRanks) != chunks*numOfThreads {
		finalRanks = append(finalRanks, <-out...)
	}

	for i := 0; i < tailcaseChunk; i++ {
		idx := i + chunks*numOfThreads
		finalRanks = append(finalRanks, Rank{Word: list[idx], Rank: match(target, list[idx])})
		idx++
	}

	sort.Sort(finalRanks)

	return finalRanks
}

func (R Ranks) Len() int {
	return len(R)
}

func (R Ranks) Less(i, j int) bool {
	return R[i].Rank < R[j].Rank
}

func (R Ranks) Swap(i, j int) {
	R[i], R[j] = R[j], R[i]
}

// take the menu.fg, menu.bg, ect. and return rgb values as uint8 values
func convertStrToInt32(color string) (r, g, b uint8) {
	return
}
