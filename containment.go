package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mathventist/duplicates"
)

func containment() {
	flags := flag.NewFlagSet("containment", flag.ExitOnError)

	flags.Usage = func() {
		usageText := `usage: dups con [ -h | --help ] <file A> <file B>

     -h, --help  print the help message

Given two documents A and B and their respective sets of ngrams, S(A), S(B), the measure to which document B is contained in document A, C(A,B), is defined as:

    C(A,B) = |S(A) ∩ S(B)| / |S(B)|

Input files must each contain a single ngram per line, and the ngrams must all be the same size for an accurate calculation.

The output is a floating point value, greater or equal to 0.0 and less than or equal to 1.0. A value of 1.0 indicates complete containment.`

		fmt.Println(usageText)
	}

	flags.Parse(os.Args[2:])
	args := flags.Args()

	if flags.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "invalid number of arguments; please provide two file names.")
		flags.Usage()

		os.Exit(1)
	}

	// use separate channels here because order is important!
	c := populateSetFromFile(args[0])
	d := populateSetFromFile(args[1])

	a, b := <-c, <-d

	fmt.Fprintf(os.Stdout, "%v\n", duplicates.Containment(a, b))
	os.Exit(0)
}
