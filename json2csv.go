package json2csv

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/astaxie/flatmap"
)

const (
	Stdin  = ""
	Stdout = ""
)

const (
	defaultDelim = ","
	// Buffer size for batched writes - optimized for performance
	defaultBatchSize = 1000
)

type LineReader interface {
	ReadBytes(delim byte) (line []byte, err error)
}

// KeyCache stores pre-split keys to avoid repeated string operations
type KeyCache struct {
	cache map[string][]string
}

func NewKeyCache() *KeyCache {
	return &KeyCache{
		cache: make(map[string][]string),
	}
}

func (kc *KeyCache) GetExpandedKey(key string) []string {
	if expanded, exists := kc.cache[key]; exists {
		return expanded
	}
	expanded := strings.Split(key, ".")
	kc.cache[key] = expanded
	return expanded
}

func flattenKeys(data map[string]interface{}) ([]string, error) {
	fm, err := flatmap.Flatten(data)
	if err != nil {
		return nil, err
	}
	// Pre-allocate slice with known capacity
	ks := make([]string, 0, len(fm))
	for k := range fm {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	return ks, nil
}

// Optimized getValue function using iterative approach instead of recursive
func getValue(data map[string]interface{}, keys []string) string {
	current := data
	
	// Iterate through keys instead of recursion
	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - extract the value
			if v, ok := current[key]; ok {
				switch v.(type) {
				case nil:
					return ""
				case string:
					return v.(string)
				case float64:
					f := v.(float64)
					if math.Mod(f, 1.0) == 0.0 {
						return fmt.Sprintf("%d", int(f))
					} else {
						return fmt.Sprintf("%f", f)
					}
				case interface{}:
					raw, _ := json.Marshal(v)
					return string(raw)
				default:
					return fmt.Sprintf("%v", v)
				}
			}
			return ""
		} else {
			// Intermediate key - navigate deeper
			if next, ok := current[key].(map[string]interface{}); ok {
				current = next
			} else {
				return ""
			}
		}
	}
	
	return ""
}

// BatchWriter wraps csv.Writer to provide batched writing for better I/O performance
type BatchWriter struct {
	writer    *csv.Writer
	records   [][]string
	batchSize int
}

func NewBatchWriter(w *csv.Writer, batchSize int) *BatchWriter {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	return &BatchWriter{
		writer:    w,
		records:   make([][]string, 0, batchSize),
		batchSize: batchSize,
	}
}

func (bw *BatchWriter) Write(record []string) {
	// Create a copy of the record to avoid reference issues
	recordCopy := make([]string, len(record))
	copy(recordCopy, record)
	
	bw.records = append(bw.records, recordCopy)
	
	if len(bw.records) >= bw.batchSize {
		bw.Flush()
	}
}

func (bw *BatchWriter) Flush() {
	for _, record := range bw.records {
		bw.writer.Write(record)
	}
	bw.writer.Flush()
	// Reset slice but keep capacity
	bw.records = bw.records[:0]
}

func json2csv(r LineReader, w *csv.Writer, allKeys bool, keys []string, printHeader bool) (int, error) {
	var (
		line      []byte
		err       error
		lineCount int
	)

	// Initialize key cache for performance
	keyCache := NewKeyCache()
	
	// Pre-expand keys once instead of on every record
	var expandedKeys [][]string
	if !allKeys {
		expandedKeys = make([][]string, 0, len(keys))
		for _, key := range keys {
			expandedKeys = append(expandedKeys, keyCache.GetExpandedKey(key))
		}
	}

	// Use batch writer for better I/O performance
	batchWriter := NewBatchWriter(w, defaultBatchSize)
	defer batchWriter.Flush() // Ensure final batch is written

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
			// Pre-expand all keys once
			expandedKeys = make([][]string, 0, len(keys))
			for _, key := range keys {
				expandedKeys = append(expandedKeys, keyCache.GetExpandedKey(key))
			}
			allKeys = false
		}

		if printHeader {
			batchWriter.Write(keys)
			printHeader = false
		}

		// Pre-allocate record slice with known capacity
		record := make([]string, 0, len(expandedKeys))
		for _, key := range expandedKeys {
			record = append(record, getValue(data, key))
		}

		batchWriter.Write(record)
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
		defer f.Close() // Ensure file is closed
		reader = bufio.NewReader(f)
	}

	if outputFile == Stdout {
		writer = csv.NewWriter(os.Stdout)
	} else {
		f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return 0, fmt.Errorf("error %s opening output file %v", err, outputFile)
		}
		defer f.Close() // Ensure file is closed
		writer = csv.NewWriter(f)
	}

	if outputDelim == "" {
		outputDelim = defaultDelim
	}

	delim, _ := utf8.DecodeRuneInString(outputDelim)
	writer.Comma = delim

	return json2csv(reader, writer, allKeys, keys, printHeader)
}
