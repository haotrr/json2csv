package main

import (
	"flag"
	"fmt"

	"github.com/haotrr/json2csv"
)

var Version = "1.0.0"
var BuildTime = ""

func main() {
	var (
		inputFile   = flag.String("i", json2csv.Stdin, "/path/to/input.json (optional; default is stdin)")
		outputFile  = flag.String("o", json2csv.Stdout, "/path/to/output.csv (optional; default is stdout)")
		outputDelim = flag.String("d", ",", "delimiter used for output values")
		showVersion = flag.Bool("version", false, "print version string")
		printHeader = flag.Bool("H", false, "prints header to output")
		allKeys     = flag.Bool("a", false, "output all keys")
	)

	keys := json2csv.StringArray{}
	flag.Var(&keys, "k", "fields to output")

	flag.Parse()

	if *showVersion {
		fmt.Printf("json2csv %s\n", Version)
		return
	}

	_, err := json2csv.Do(*inputFile, *outputFile, *outputDelim, *allKeys, keys, *printHeader)
	if err != nil {
		fmt.Println(err)
	}
}
