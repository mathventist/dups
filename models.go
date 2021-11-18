package main

import (
	"github.com/mathventist/duplicates"
	"github.com/schollz/progressbar/v3"
	"sync"
)

type stringPair struct {
	original  string
	processed string
}

type matchedStrings struct {
	aString string
	bString string
	score   float64
}

type result struct {
	aIndex int
	bIndex int
	match  matchedStrings
}

type indexedStrings struct {
	matches [][]matchedStrings
	mu      sync.Mutex
}

func (s *indexedStrings) Add(i, j int, a, b string, score float64) {
	s.mu.Lock()
	s.matches[i][j] = matchedStrings{
		aString: a,
		bString: b,
		score:   score,
	}
	s.mu.Unlock()
}

func newIndexedStrings(i, j int) indexedStrings {
	m := make([][]matchedStrings, i)
	for n := 0; n < i; n++ {
		m[n] = make([]matchedStrings, j)
	}
	return indexedStrings{
		matches: m,
	}
}

func preprocess(a []string, fileName string, removeStops bool) <-chan []stringPair {
	c := make(chan []stringPair)

	go func() {
		var r []stringPair
		bar := progressbar.Default(int64(len(a)), "preprocessing "+fileName+"...")

		for _, aa := range a {
			bar.Add(1)
			normalizedText := duplicates.Preprocess(aa, removeStops)

			// trim trailing punctuation
			if isSentenceTerminator(normalizedText[len(normalizedText)-1]) {
				normalizedText = normalizedText[:len(normalizedText)-1]
			}

			r = append(r, stringPair{aa, normalizedText})
		}
		c <- r
	}()

	return c
}
