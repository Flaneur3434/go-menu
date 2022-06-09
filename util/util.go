package util

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
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

func FuzzySearch(list []string, target string, caseInsensitive bool) Ranks {
	chunks := int(len(list) / numOfThreads)
	tailcaseChunk := len(list) % numOfThreads
	RanksChan := make(chan Ranks)
	// syncChan: used to signal when all of the go routines end
	syncChan := make(chan int)

	for i := 0; i < numOfThreads; i++ {
		go func(threadN int) {
			var rankSlice Ranks
			var rankNum float64
			for j := threadN * chunks; j < (threadN+1)*chunks; j++ {
				if caseInsensitive {
					rankNum = match(target, strings.ToLower(list[j]))
				} else {
					rankNum = match(target, list[j])
				}

				if rankNum != math.MaxFloat64 {
					rankSlice = append(rankSlice, Rank{Word: list[j], Rank: rankNum})
				}
			}

			RanksChan <- rankSlice
			syncChan <- 1
		}(i)
	}

	go func() {
		var rankSlice Ranks
		for i := 0; i < tailcaseChunk; i++ {
			idx := i + chunks*numOfThreads
			if rankNum := match(target, list[idx]); rankNum != math.MaxFloat64 {
				rankSlice = append(rankSlice, Rank{Word: list[idx], Rank: rankNum})
			}

			idx++
		}
		RanksChan <- rankSlice
		syncChan <- 1
	}()

	finalRanks := make(Ranks, 0, len(list))
	syncChanTotal := 0
	// numOfThreads + 1 because we need to account for the tail case go routine
	for syncChanTotal < (numOfThreads + 1) {
		finalRanks = append(finalRanks, <-RanksChan...)
		syncChanTotal += <-syncChan
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
func ConvertStrToInt32(color string) (r, g, b uint8) {
	var garbage string
	fmt.Sscanf(color, "%1s%2x%2x%2x", &garbage, &r, &g, &b)
	return
}
