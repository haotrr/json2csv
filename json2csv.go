package json2csv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Stdin  = ""
	Stdout = ""

	defaultDelim = ","
)

type LineReader interface {
	ReadBytes(delim byte) (line []byte, err error)
}

func get_value(data map[string]interface{}, keyparts []string) string {
	if len(keyparts) > 1 {
		subdata, _ := data[keyparts[0]].(map[string]interface{})
		return get_value(subdata, keyparts[1:])
	} else if v, ok := data[keyparts[0]]; ok {
		switch v.(type) {
		case nil:
			return ""
		case float64:
			f, _ := v.(float64)
			if math.Mod(f, 1.0) == 0.0 {
				return fmt.Sprintf("%d", int(f))
			} else {
				return fmt.Sprintf("%f", f)
			}
		default:
			return fmt.Sprintf("%+v", v)
		}
	}

	return ""
}

func json2csv(r LineReader, w *csv.Writer, keys []string, printHeader bool) (int, error) {
	var line []byte
	var err error
	line_count := 0

	var expanded_keys [][]string
	for _, key := range keys {
		expanded_keys = append(expanded_keys, strings.Split(key, "."))
	}

	for {
		if err == io.EOF {
			return line_count, nil
		}
		line, err = r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				return 0, fmt.Errorf("input ERROR: %s", err)
			}
		}
		line_count++
		if len(line) == 0 {
			continue
		}

		if printHeader {
			w.Write(keys)
			w.Flush()
			printHeader = false
		}

		var data map[string]interface{}
		err = json.Unmarshal(line, &data)
		if err != nil {
			continue
		}

		var record []string
		for _, expanded_key := range expanded_keys {
			record = append(record, get_value(data, expanded_key))
		}

		w.Write(record)
		w.Flush()
	}

	return line_count, nil
}

func Do(inputFile, outputFile, outputDelim string, keys []string, printHeader bool) (int, error) {
	var reader *bufio.Reader
	var writer *csv.Writer
	if inputFile == Stdin {
		reader = bufio.NewReader(os.Stdin)
	} else {
		f, err := os.OpenFile(inputFile, os.O_RDONLY, 0600)
		if err != nil {
			return 0, fmt.Errorf("error %s opening input file %v", err, inputFile)
		}
		reader = bufio.NewReader(f)
	}

	if outputFile == Stdout {
		writer = csv.NewWriter(os.Stdout)
	} else {
		f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return 0, fmt.Errorf("error %s opening output file %v", err, outputFile)
		}
		writer = csv.NewWriter(f)
	}

	if outputDelim == "" {
		outputDelim = defaultDelim
	}

	delim, _ := utf8.DecodeRuneInString(outputDelim)
	writer.Comma = delim

	return json2csv(reader, writer, keys, printHeader)
}
