package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/golang-collections/collections/set"
	"github.com/mathventist/duplicates"
)

func main() {

	flag.Usage = func() {
		usageText := `usage: dups [-v | --version] [-h | --help] <command> [<args>]`
		fmt.Println(usageText)
	}

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "invalid number of arguments; please provide two file names.")
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "con":
		containment()
	case "res":
		resemblance()
	case "norm":
		normalize()
	case "w2v":
		w2v()
	case "eq":
		eq()
	case "str2n":
		str2ngrams()
	case "str2s":
		str2sentences()
	}

}

func populateSetFromFile(fileName string) <-chan *set.Set {
	c := make(chan *set.Set)

	go func() {
		f, err := os.Open(fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening file: ", err)
			flag.Usage()

			os.Exit(1)
		}
		defer f.Close()

		s := set.New()

		fs := bufio.NewScanner(f)
		for fs.Scan() {
			s.Insert(fs.Text())
		}

		c <- s
	}()

	return c
}

func isSentenceTerminator(b byte) bool {
	return b == '.' || b == '?' || b == '!'
}

func populateSliceFromFile(fileName string) <-chan []string {
	c := make(chan []string)

	go func() {
		f, err := os.Open(fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening file: ", err)
			flag.Usage()

			os.Exit(1)
		}
		defer f.Close()

		fileScanner := bufio.NewScanner(f)
		fileScanner.Split(duplicates.ScanSentences)

		var lines []string
		for fileScanner.Scan() {
			lines = append(lines, fileScanner.Text())
		}

		c <- lines
	}()

	return c
}
