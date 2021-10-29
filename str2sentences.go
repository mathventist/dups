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
		usageText := `usage: dups str2s [ -h | --help ] [ -f <filename> | --file <filename> ]

    -f, --file <filename>   input filename, standard input when omitted
    -h, --help              print the help message`

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
