package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"code.sajari.com/word2vec"
	"github.com/mathventist/duplicates"
	"github.com/mathventist/heatmap"
	"github.com/schollz/progressbar/v3"
)

var model *word2vec.Model
var cache word2vec.Coser

type matchGroup struct {
	A        string
	B        string
	Score    float32
	NotFound []string
}

func w2v() {
	var removeStops bool
	var score float64
	var heatMapFileName string

	flags := flag.NewFlagSet("w2v", flag.ExitOnError)
	flags.BoolVar(&removeStops, "r", true, "remove stop words from text")
	flags.BoolVar(&removeStops, "removeStops", true, "remove stop words from text")
	flags.Float64Var(&score, "s", float64(0.5), "minimum score for which matches will be reported")
	flags.Float64Var(&score, "score", float64(0.5), "minimum score for which matches will be reported")
	flags.StringVar(&heatMapFileName, "heatmap", "", "filename of the generatid heatmap")
	flags.StringVar(&heatMapFileName, "h", "", "filename of the generatid heatmap")

	flags.Usage = func() {
		usageText := `usage: dups w2v [ -h | --help ] [ -r | --removeStops ] [ -s <score> | --score <score>] [ -h <filename> | --heatmap <filename> ] <Word2Vec model file> <file A> <file B>

        -h, --help                 print the help message
        -r, --removeStops          remove English stop words from the text
        -s, --score <score>        only return matches greater or equal to this value between 0 and 1; ignored if -h is set
        -h, --heatmap <filename>   generate a heatmap of the scores with the given filename; -s is ignored when this is set`

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
	cache = word2vec.NewCache(model)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading word2vec model: %v\n", err)
		os.Exit(1)
	}

	r := w2vCompare(a, b, fileName1, fileName2, removeStops)

	if heatMapFileName != "" {
		generateHeatMap(r, heatMapFileName)
	} else {
		printResults(r, score)
	}
}

func generateHeatMap(r [][]matchedStrings, heatMapFileName string) {
	data := make([][]float32, len(r))
	for j := range data {
		data[j] = make([]float32, len(r[0]))
	}

	for i, n := range r {
		for j, m := range n {
			data[i][j] = float32(m.score)
		}
	}

	if err := heatmap.DrawHeatMap(data, 25, heatMapFileName); err != nil {
		fmt.Fprintf(os.Stderr, "error generating heatmap: %v\n", err)
		os.Exit(1)
	}
}

func printResults(r [][]matchedStrings, score float64) {
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

	// sort highest matches first
	sort.Slice(results, func(i, j int) bool {
		return results[i].match.score > results[j].match.score
	})

	fmt.Printf("%v\n", results)
}

// TODO: fix this concurrent version, or delete.
/*
func oldw2vCompare(a, b []string, fileName1, fileName2 string, removeStops bool) [][]matchedStrings {
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
*/

func w2vCompare(a, b []string, fileName1, fileName2 string, removeStops bool) [][]matchedStrings {
	ca := preprocess(a, fileName1, removeStops)
	cb := preprocess(b, fileName2, removeStops)

	la, lb := <-ca, <-cb

	results := newIndexedStrings(len(a), len(b))

	bar := progressbar.Default(int64(len(a)*len(b)), "comparing files...")

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

			normalizedScore, _ := duplicates.CompareWord2Vec(aa.processed, bb.processed, model, cache)
			results.Add(i, j, aa.original, bb.original, float64(normalizedScore))

			bar.Add(1)
		}
	}

	return results.matches
}
