package main

import (
	"flag"
	"fmt"
	"os"

	"code.sajari.com/word2vec"
	"github.com/mathventist/duplicates"
	"github.com/schollz/progressbar/v3"
)

var origToNorm = make(map[string]string)

type matchGroup struct {
	A        string
	B        string
	Score    float32
	NotFound []string
}

func w2v() {
	var removeStops bool
	var modelFileName string
	var score float64

	flags := flag.NewFlagSet("w2v", flag.ExitOnError)
	flags.BoolVar(&removeStops, "r", true, "remove stop words from text")
	flags.BoolVar(&removeStops, "removeStops", true, "remove stop words from text")
	flags.Float64Var(&score, "s", float64(0.5), "minimum score for which matches will be reported")
	flags.Float64Var(&score, "score", float64(0.5), "minimum score for which matches will be reported")
	flags.StringVar(&modelFileName, "m", "", "word2vec model filename")
	flags.StringVar(&modelFileName, "model", "", "word2vec model filename")

	flags.Usage = func() {
	}

	flags.Parse(os.Args[2:])
	args := flags.Args()

	if flags.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "invalid number of arguments; please provide two input file names.")
		flags.Usage()

		os.Exit(1)
	}

	if modelFileName == "" {
		fmt.Fprintln(os.Stderr, "must provide filename of word2vec model.")
		flag.Usage()

		os.Exit(1)
	}

	fileName1, fileName2 := args[0], args[1]

	c := populateSliceFromFile(fileName1)
	d := populateSliceFromFile(fileName2)

	a, b := <-c, <-d

	// Load word2vec model
	// Ex: curl -O https://s3.amazonaws.com/dl4j-distribution/GoogleNews-vectors-negative300.bin.gz
	r, err := os.Open(modelFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", modelFileName, err)
		os.Exit(1)
	}
	defer r.Close()

	model, err := word2vec.FromReader(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading word2vec model: %v\n", err)
		os.Exit(1)
	}

	matches := w2vCompare(a, b, fileName1, fileName2, model, removeStops, score)
	for _, m := range matches {
		fmt.Println(m.A)
		fmt.Println(m.B)
		fmt.Println(m.NotFound)
		fmt.Println(m.Score)

	}

}

func w2vCompare(a, b []string, fileName1, fileName2 string, model *word2vec.Model, removeStops bool, score float64) []matchGroup {
	bar := progressbar.Default(int64(len(a)*len(b)), "comparing files...")

	matches := []matchGroup{}

	for _, aa := range a {
		na := preprocess(aa, removeStops)
		if len(na) == 0 {
			continue
		}

		for _, bb := range b {
			nb := preprocess(bb, removeStops)
			if len(nb) == 0 {
				continue
			}

			normalizedScore, notFound := duplicates.CompareWord2Vec(na, nb, model)

			if float64(normalizedScore) >= score {
				match := matchGroup{
					A:        aa,
					B:        bb,
					Score:    normalizedScore,
					NotFound: notFound,
				}
				matches = append(matches, match)
			}
			bar.Add(1)
		}
	}

	return matches
}

func preprocess(a string, removeStops bool) string {
	if len(a) == 0 {
		return a
	}

	// Check cache of preprocessed text first
	if val, ok := origToNorm[a]; ok {
		return val
	}

	normalizedText := duplicates.Preprocess(a, removeStops)

	// trim trailing punctuation
	if isSentenceTerminator(normalizedText[len(normalizedText)-1]) {
		normalizedText = normalizedText[:len(normalizedText)-1]
	}
	origToNorm[a] = normalizedText

	return normalizedText
}
