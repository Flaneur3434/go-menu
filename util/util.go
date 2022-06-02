package util

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
)

type Rank struct {
	Word string
	rank float64
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
		R = append(R, Rank{Word: s[i], rank: -1})
	}

	return
}

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
		/* compute rank */
		/* add penalty if match starts late (log(sidx+2))
		 * add penalty for long a match without many matching characters */
		return math.Log(float64(sidx)+2) + float64(eidx-sidx-len(s))
	} else {
		return math.MaxFloat64
	}
}

func FuzzySearch(R *Ranks, target string) {
	if len(target) < 1 {
		return
	}

	numOfThreads := 30
	chunks := int(len(*R) / numOfThreads)
	tailcaseChunk := len(*R) % numOfThreads
	out := make(chan Ranks)

	for i := 0; i < numOfThreads; i++ {
		go func(idx int, in chan Ranks) {
			var rankSlice Ranks
			for idx, j := 0, idx*chunks; idx < chunks; {
				rankSlice = append(rankSlice, Rank{Word: (*R)[j].Word, rank: match(target, (*R)[j].Word)})
				idx++
				j++
			}

			sort.Sort(rankSlice)
			in <- rankSlice
		}(i, out)
	}

	finalRanks := make(Ranks, 0, len(*R))
	for len(finalRanks) != chunks*numOfThreads {
		finalRanks = append(finalRanks, <-out...)
	}

	for i := 0; i < tailcaseChunk; i++ {
		idx := i + chunks*numOfThreads
		finalRanks = append(finalRanks, Rank{Word: (*R)[idx].Word, rank: match(target, (*R)[idx].Word)})
		idx++
	}

	sort.Sort(finalRanks)

	// why? because of thread thrashing (probably)
	*R = finalRanks
}

func (R Ranks) Len() int {
	return len(R)
}

func (R Ranks) Less(i, j int) bool {
	return R[i].rank < R[j].rank
}

func (R Ranks) Swap(i, j int) {
	R[i], R[j] = R[j], R[i]
}
