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
            if m.aString != "" && m.bString != "" {
                s := result{
                    aIndex:  i,
                    bIndex: j,
                    match: m,
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

type result struct {
    aIndex int
    bIndex int
    match matchedStrings
}

type matchedStrings struct {
	aString string
	bString string
}

type indexedStrings struct {
    matches [][]matchedStrings
    mu sync.Mutex
}

func (s *indexedStrings) Add(i, j int, a, b string) {
    s.mu.Lock()
    s.matches[i][j] = matchedStrings{
        aString: a,
        bString: b,
    }
    s.mu.Unlock()
}

func compare(a []string, b []string, fileName1 string, fileName2 string, removeStops bool) [][]matchedStrings {
	ca := preprocessEq(a, fileName1, removeStops)
	cb := preprocessEq(b, fileName2, removeStops)

	la, lb := <-ca, <-cb

    m := make([][]matchedStrings, len(a))
    for n := 0; n < len(a); n++ {
        m[n] = make([]matchedStrings, len(b))
    }
    results := indexedStrings{
        matches: m,
    }

	bar := progressbar.Default(int64(len(a)*len(b)), "comparing files...")

	// TODO: improve performance by using goroutines to run comparisons concurrently.
    var wg  sync.WaitGroup

	for i, aa := range la {
		for j, bb := range lb {
            wg.Add(1)

            go func(aString, bString string, i, j int) {
                defer wg.Done()

                if aString == bString {
                    results.Add(i, j, a[i], b[j])
                }
                bar.Add(1)
            }(aa, bb, i, j)
		}
	}

    wg.Wait()

	return results.matches
}

func preprocessEq(a []string, fileName string, removeStops bool) <-chan []string {
	c := make(chan []string)

	go func() {
		var r []string
		bar := progressbar.Default(int64(len(a)), "preprocessing "+fileName+"...")

		for _, aa := range a {
			bar.Add(1)
			normalizedText := duplicates.Preprocess(aa, removeStops)

			// trim trailing punctuation
			if isSentenceTerminator(normalizedText[len(normalizedText)-1]) {
				normalizedText = normalizedText[:len(normalizedText)-1]
			}

			r = append(r, normalizedText)
		}
		c <- r
	}()

	return c
}
