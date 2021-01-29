package json2csv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/astaxie/flatmap"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Stdin  = ""
	Stdout = ""
)

const (
	defaultDelim = ","
)

type LineReader interface {
	ReadBytes(delim byte) (line []byte, err error)
}

func flattenKeys(data map[string]interface{}) ([]string, error) {
	fm, err := flatmap.Flatten(data)
	if err != nil {
		return nil, err
	}
	var ks []string
	for k := range fm {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	return ks, nil
}

func getValue(data map[string]interface{}, keys []string) string {
	if len(keys) > 1 {
		subdata, _ := data[keys[0]].(map[string]interface{})
		return getValue(subdata, keys[1:])
	} else if v, ok := data[keys[0]]; ok {
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

func json2csv(r LineReader, w *csv.Writer, allKeys bool, keys []string, printHeader bool) (int, error) {
	var (
		line      []byte
		err       error
		lineCount int
	)

	var expandedKeys [][]string
	for _, key := range keys {
		expandedKeys = append(expandedKeys, strings.Split(key, "."))
	}

	for {
		if err == io.EOF {
			return lineCount, nil
		}
		line, err = r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				return 0, fmt.Errorf("input ERROR: %s", err)
			}

			break
		}
		lineCount++
		if len(line) == 0 {
			continue
		}

		var data map[string]interface{}
		err = json.Unmarshal(line, &data)
		if err != nil {
			continue
		}

		if allKeys {
			flattened, err := flattenKeys(data)
			if err != nil {
				return 0, err
			}
			keys = flattened
			for _, key := range keys {
				expandedKeys = append(expandedKeys, strings.Split(key, "."))
			}
			allKeys = false
		}

		if printHeader {
			w.Write(keys)
			w.Flush()
			printHeader = false
		}

		var record []string
		for _, key := range expandedKeys {
			record = append(record, getValue(data, key))
		}

		w.Write(record)
		w.Flush()
	}

	return lineCount, nil
}

func Do(inputFile, outputFile, outputDelim string, allKeys bool, keys []string, printHeader bool) (int, error) {
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

	return json2csv(reader, writer, allKeys, keys, printHeader)
}
