package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mathventist/duplicates"
)

func str2sentences() {
	var fileName string

	flags := flag.NewFlagSet("str2sentences", flag.ExitOnError)
	flags.StringVar(&fileName, "f", "", "input filename")
	flags.StringVar(&fileName, "file", "", "input filename")

	flags.Usage = func() {
		usageText := `str2sentences - a utility for splitting strings into sentences.

Given text input, this outputs the text back, reformated into one sentence per line.

It uses the characters .!? as markers for the end of a sentence. It also handles sentences that end in quotations, by including the end quotation mark as part of the sentence when the quotation mark immediately follows the end of a sentence.

Furthermore, it recognizes that period characters appearing in abbreviations, such as "Mrs." or "St.", do not mark the end of sentences.

USAGE
  $ str2sentences [ -h | --help ] [ -f <filename> | --file <filename> ]

OPTIONS
  -f, --file         input filename. Standard input when omitted
  -h, --help         print the help message

EXAMPLES
  $ echo "Sentence one.     sentence, two!" | str2sentences
  Sentence one.
  sentence, two!

  $ str2sentences -f myfile
  `

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
	scanner.Split(duplicates.ScanSentences)
	for scanner.Scan() {
		fmt.Println(strings.TrimSpace(scanner.Text()))
	}
}
