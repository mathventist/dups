package main

import (
	"flag"
	"fmt"
	"os"
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
	results := compare(a, b, fileName1, fileName2, removeStops)

	// Display result summary
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "File\tNumber of sentences\tPercentage of matched sentences")
	fmt.Fprintf(w, "%v\t%v\t%v\n", fileName1, len(a), 100*len(results)/len(a))
	fmt.Fprintf(w, "%v\t%v\t%v\n", fileName2, len(b), 100*len(results)/len(b))
	w.Flush()

	// Display full results
	fmt.Printf("\n\n%v matched sentences.\n\n", len(results))
	for _, r := range results {
		fmt.Fprintf(os.Stdout, "%v sentence number %v\n\n\t%v\n\nmatched to %v sentence number %v\n\n\t%v\n\n",
			fileName1, r[0].Index+1, r[0].String,
			fileName2, r[1].Index+1, r[1].String,
		)
	}
}

type indexedString struct {
	Index  int
	String string
}

func compare(a []string, b []string, fileName1 string, fileName2 string, removeStops bool) [][2]indexedString {
	ca := preprocessEq(a, fileName1, removeStops)
	cb := preprocessEq(b, fileName2, removeStops)

	la, lb := <-ca, <-cb

	var results [][2]indexedString
	bar := progressbar.Default(int64(len(a)*len(b)), "comparing files...")

	// TODO: improve performance by using goroutines to run comparisons concurrently.
	for i, aa := range la {
		for j, bb := range lb {
			bar.Add(1)
			if aa == bb {
				var match [2]indexedString
				match[0] = indexedString{i, a[i]}
				match[1] = indexedString{j, b[j]}

				results = append(results, match)
			}
		}
	}

	return results
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
