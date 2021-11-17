package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/mathventist/duplicates"
	"github.com/schollz/progressbar/v3"
)

func eq() {
	var removeStops bool

	flags := flag.NewFlagSet("eq", flag.ExitOnError)
	flags.BoolVar(&removeStops, "r", true, "remove stop words from text")
	flags.BoolVar(&removeStops, "removeStops", true, "remove stop words from text")

	flags.Usage = func() {
		usageText := `usage: dups eq [ -h | --help ] [ -r | --removeStops ] <file A> <file B>

    -r, --removeStops  remove stop words from the text
    -h, --help         print the help message`

		fmt.Println(usageText)
	}

	flags.Parse(os.Args[2:])
	args := flags.Args()

	if flags.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "invalid number of arguments; please provide two file names.")
		flags.Usage()

		os.Exit(1)
	}

	fileName1, fileName2 := args[0], args[1]

	c := populateSliceFromFile(fileName1)
	d := populateSliceFromFile(fileName2)

	a, b := <-c, <-d
	r := compare(a, b, fileName1, fileName2, removeStops)

	var results []result

	for i, n := range r {
		for j, m := range n {
			if m.score == 1 {
				s := result{
					aIndex: i,
					bIndex: j,
					match:  m,
				}
				results = append(results, s)
			}
		}
	}

	// Display result summary
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "File\tNumber of sentences\tPercentage of matched sentences")
	fmt.Fprintf(w, "%v\t%v\t%v\n", fileName1, len(a), 100*len(results)/len(a))
	fmt.Fprintf(w, "%v\t%v\t%v\n", fileName2, len(b), 100*len(results)/len(b))
	w.Flush()

	// Display full results
	fmt.Printf("\n\n%v matched sentences.\n\n", len(results))
	for _, t := range results {
		fmt.Fprintf(os.Stdout, "%v sentence number %v\n\n\t%v\n\nmatched to %v sentence number %v\n\n\t%v\n\n",
			fileName1, t.aIndex, t.match.aString,
			fileName2, t.bIndex, t.match.bString,
		)
	}
}

type stringPair struct {
	original  string
	processed string
}

type result struct {
	aIndex int
	bIndex int
	match  matchedStrings
}

type matchedStrings struct {
	aString string
	bString string
	score   float64
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

func compare(a, b []string, fileName1, fileName2 string, removeStops bool) [][]matchedStrings {
	ca := preprocess(a, fileName1, removeStops)
	cb := preprocess(b, fileName2, removeStops)

	la, lb := <-ca, <-cb

	results := newIndexedStrings(len(a), len(b))

	bar := progressbar.Default(int64(len(a)*len(b)), "comparing files...")

	var wg sync.WaitGroup

	for i, aa := range la {
		if len(aa.processed) == 0 {
			bar.Add(len(lb))
			continue
		}

		for j, bb := range lb {
			if len(bb.processed) == 0 {
				bar.Add(1)
				continue
			}

			wg.Add(1)

			go func(aString, bString stringPair, i, j int) {
				defer wg.Done()

				score := float64(0)
				if aString.processed == bString.processed {
					score = float64(1)
				}
				results.Add(i, j, aString.original, bString.original, score)

				bar.Add(1)
			}(aa, bb, i, j)
		}
	}

	wg.Wait()

	return results.matches
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
