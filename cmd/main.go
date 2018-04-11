package main

import (
	"flag"
	"fmt"

	"github.com/qinhao/json2csv"
)

var Version = "1.0.0"
var BuildTime = ""

type LineReader interface {
	ReadBytes(delim byte) (line []byte, err error)
}

func main() {
	inputFile := flag.String("i", json2csv.Stdin, "/path/to/input.json (optional; default is stdin)")
	outputFile := flag.String("o", json2csv.Stdout, "/path/to/output.csv (optional; default is stdout)")
	outputDelim := flag.String("d", ",", "delimiter used for output values")
	showVersion := flag.Bool("version", false, "print version string")
	printHeader := flag.Bool("H", false, "prints header to output")
	keys := json2csv.StringArray{}
	flag.Var(&keys, "k", "fields to output")
	flag.Parse()

	if *showVersion {
		fmt.Printf("json2csv %s\n", Version)
		return
	}

	n, err := json2csv.Do(*inputFile, *outputFile, *outputDelim, keys, *printHeader)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println()
	fmt.Println(n, "line(s) data handled.")
}
