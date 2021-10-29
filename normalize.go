package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/mathventist/duplicates"
)

func normalize() {
	var removeStops bool
	var fileName string

	flags := flag.NewFlagSet("normalize", flag.ExitOnError)
	flags.BoolVar(&removeStops, "r", false, "remove stop words from text")
	flags.BoolVar(&removeStops, "removeStops", false, "remove stop words from text")
	flags.StringVar(&fileName, "f", "", "input filename")
	flags.StringVar(&fileName, "file", "", "input filename")

	flags.Usage = func() {
		usageText := `usage: dups norm [ -h | --help ] [ -f <filename> | --file <filename> ] [ -r | --removeStops ]

    -f, --file            input filename. Standard input when omitted
    -r, --removeStops     remove English stop words from the text
    -h, --help            print the help message

It removes titles, numerics, hyphens, and internal sentence punctuation, expands ligatures, and compresses multiple whitespace characters into a single whitespace character.

Optionally, this utility also strips out 100 of the most popular English stop words (see https://gist.github.com/sebleier/554280).`

		fmt.Println(usageText)
	}

	flags.Parse(os.Args[2:])

	var inputErr error
	file := os.Stdin

	if len(fileName) > 0 {
		file, inputErr = os.Open(fileName)

		if inputErr != nil {
			fmt.Fprintln(os.Stderr, "error opening input file: ", inputErr)
			flag.Usage()

			os.Exit(1)
		}

		defer file.Close()
	}

	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "stdin error: ", err)
		flag.Usage()

		os.Exit(1)
	}

	// Exit if stdin is empty.
	size := fi.Size()
	if size == 0 {
		fmt.Fprintln(os.Stderr, "input is empty")
		flag.Usage()

		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Fprintln(os.Stdout, duplicates.Preprocess(scanner.Text(), removeStops))
	}

	os.Exit(0)
}
