package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"code.sajari.com/word2vec"
	"github.com/mathventist/duplicates"
	"github.com/schollz/progressbar/v3"
)

var origToNorm = make(map[string]string)
var model *word2vec.Model

type matchGroup struct {
	A        string
	B        string
	Score    float32
	NotFound []string
}

func w2v() {
	var removeStops bool
	var score float64

	flags := flag.NewFlagSet("w2v", flag.ExitOnError)
	flags.BoolVar(&removeStops, "r", true, "remove stop words from text")
	flags.BoolVar(&removeStops, "removeStops", true, "remove stop words from text")
	flags.Float64Var(&score, "s", float64(0.5), "minimum score for which matches will be reported")
	flags.Float64Var(&score, "score", float64(0.5), "minimum score for which matches will be reported")

	flags.Usage = func() {
		usageText := `usage: dups w2v [ -h | --help ] [ -r | --removeStops ] [ -s <score> | --score <score>] <Word2Vec model file> <file A> <file B>

        -h, --help            print the help message
        -r, --removeStops     remove English stop words from the text
        -s, --score <score>   only return matches greater or equal to this value between 0 and 1`

		fmt.Println(usageText)
	}

	flags.Parse(os.Args[2:])
	args := flags.Args()

	if flags.NArg() != 3 {
		fmt.Fprintln(os.Stderr, "invalid number of arguments; please provide the Word2Vec model file, followed by two input file names.")
		flags.Usage()

		os.Exit(1)
	}

	modelFileName, fileName1, fileName2 := args[0], args[1], args[2]

	c := populateSliceFromFile(fileName1)
	d := populateSliceFromFile(fileName2)

	a, b := <-c, <-d

	// Load word2vec model
	// Ex: curl -O https://s3.amazonaws.com/dl4j-distribution/GoogleNews-vectors-negative300.bin.gz
	f, err := os.Open(modelFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", modelFileName, err)
		os.Exit(1)
	}
	defer f.Close()

	model, err = word2vec.FromReader(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading word2vec model: %v\n", err)
		os.Exit(1)
	}

	r := w2vCompare(a, b, fileName1, fileName2, removeStops)

	var results []result

	for i, n := range r {
		for j, m := range n {
			if m.score >= score {
				s := result{
					aIndex: i,
					bIndex: j,
					match:  m,
				}
				results = append(results, s)
			}
		}
	}

	fmt.Printf("%v\n", results)
}

func w2vCompare(a, b []string, fileName1, fileName2 string, removeStops bool) [][]matchedStrings {
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

				normalizedScore, _ := duplicates.CompareWord2Vec(aString.processed, bString.processed, model)
				results.Add(i, j, aString.original, bString.original, float64(normalizedScore))

				bar.Add(1)
			}(aa, bb, i, j)
		}
	}

	return results.matches
}
