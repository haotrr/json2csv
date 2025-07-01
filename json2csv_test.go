package json2csv

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

// Sample JSON data for benchmarking
const sampleJSON = `{"user": {"name":"john", "age": 30}, "remote_ip": "127.0.0.1", "timestamp": "2023-01-01T00:00:00Z"}
{"user": {"name":"jane", "age": 25}, "remote_ip": "192.168.1.1", "timestamp": "2023-01-01T00:01:00Z"}
{"user": {"name":"bob", "age": 35}, "remote_ip": "10.0.0.1", "timestamp": "2023-01-01T00:02:00Z"}`

// BenchmarkKeyCache tests the performance of key caching
func BenchmarkKeyCache(b *testing.B) {
	keys := []string{"user.name", "user.age", "remote_ip", "timestamp"}
	
	b.Run("WithCache", func(b *testing.B) {
		cache := NewKeyCache()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, key := range keys {
				cache.GetExpandedKey(key)
			}
		}
	})
	
	b.Run("WithoutCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, key := range keys {
				strings.Split(key, ".")
			}
		}
	})
}

// BenchmarkGetValue tests the performance of the optimized getValue function
func BenchmarkGetValue(b *testing.B) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "john",
			"age":  30.0,
		},
		"remote_ip": "127.0.0.1",
		"timestamp": "2023-01-01T00:00:00Z",
	}
	
	keys := []string{"user", "name"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getValue(data, keys)
	}
}

// BenchmarkBatchWriter tests the performance of batched writing
func BenchmarkBatchWriter(b *testing.B) {
	records := [][]string{
		{"john", "30", "127.0.0.1", "2023-01-01T00:00:00Z"},
		{"jane", "25", "192.168.1.1", "2023-01-01T00:01:00Z"},
		{"bob", "35", "10.0.0.1", "2023-01-01T00:02:00Z"},
	}
	
	b.Run("BatchWriter", func(b *testing.B) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		batchWriter := NewBatchWriter(writer, 100)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, record := range records {
				batchWriter.Write(record)
			}
			batchWriter.Flush()
		}
	})
	
	b.Run("DirectWriter", func(b *testing.B) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, record := range records {
				writer.Write(record)
				writer.Flush()
			}
		}
	})
}

// BenchmarkJSON2CSV tests the overall performance of the json2csv function
func BenchmarkJSON2CSV(b *testing.B) {
	keys := []string{"user.name", "user.age", "remote_ip", "timestamp"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(sampleJSON)
		lineReader := &stringLineReader{reader: reader}
		
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		
		json2csv(lineReader, writer, false, keys, false)
	}
}

// stringLineReader implements LineReader for string input
type stringLineReader struct {
	reader *strings.Reader
}

func (r *stringLineReader) ReadBytes(delim byte) ([]byte, error) {
	var result []byte
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			if len(result) > 0 {
				return result, nil
			}
			return nil, err
		}
		result = append(result, b)
		if b == delim {
			return result, nil
		}
	}
}

// BenchmarkMemoryAllocation tests memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	keys := []string{"user.name", "user.age", "remote_ip", "timestamp"}
	cache := NewKeyCache()
	
	b.Run("PreAllocatedSlice", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			expandedKeys := make([][]string, 0, len(keys))
			for _, key := range keys {
				expandedKeys = append(expandedKeys, cache.GetExpandedKey(key))
			}
		}
	})
	
	b.Run("DynamicSlice", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var expandedKeys [][]string
			for _, key := range keys {
				expandedKeys = append(expandedKeys, strings.Split(key, "."))
			}
		}
	})
}