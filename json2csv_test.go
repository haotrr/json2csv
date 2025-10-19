package json2csv

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"testing"
)

// Sample JSON data for benchmarking
const sampleJSON = `{"user": {"name":"john", "age": 30}, "remote_ip": "127.0.0.1", "timestamp": "2023-01-01T00:00:00Z"}
{"user": {"name":"jane", "age": 25}, "remote_ip": "192.168.1.1", "timestamp": "2023-01-01T00:01:00Z"}
{"user": {"name":"bob", "age": 35}, "remote_ip": "10.0.0.1", "timestamp": "2023-01-01T00:02:00Z"}`

// Test KeyCache functionality
func TestKeyCache(t *testing.T) {
	cache := NewKeyCache()

	// Test first call - should split and cache
	result1 := cache.GetExpandedKey("user.name.address")
	if len(result1) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(result1))
	}

	// Test second call - should return cached result
	result2 := cache.GetExpandedKey("user.name.address")
	if !equalStringSlices(result1, result2) {
		t.Error("Cached result should be identical")
	}

	// Test different key
	result3 := cache.GetExpandedKey("simple.key")
	if len(result3) != 2 {
		t.Errorf("Expected 2 parts, got %d", len(result3))
	}
}

// Test flattenKeys functionality
func TestFlattenKeys(t *testing.T) {
	// Test simple nested structure
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "john",
			"age":  30,
		},
		"ip": "127.0.0.1",
	}

	keys, err := flattenKeys(data)
	if err != nil {
		t.Fatalf("flattenKeys failed: %v", err)
	}

	expectedKeys := []string{"ip", "user.age", "user.name"}
	if !equalStringSlices(keys, expectedKeys) {
		t.Errorf("Expected %v, got %v", expectedKeys, keys)
	}

	// Test deeper nesting
	deepData := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "value",
			},
		},
	}

	keys, err = flattenKeys(deepData)
	if err != nil {
		t.Fatalf("flattenKeys failed: %v", err)
	}

	if len(keys) != 1 || keys[0] != "a.b.c" {
		t.Errorf("Expected [a.b.c], got %v", keys)
	}
}

// Test getValue functionality
func TestGetValue(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "john",
			"age":  30,
			"active": true,
			"score": 95.5,
		},
		"ip": "127.0.0.1",
		"nullValue": nil,
	}

	// Test existing string field
	if result := getValue(data, []string{"user", "name"}); result != "john" {
		t.Errorf("Expected 'john', got '%s'", result)
	}

	// Test existing integer field
	if result := getValue(data, []string{"user", "age"}); result != "30" {
		t.Errorf("Expected '30', got '%s'", result)
	}

	// Test existing float field
	if result := getValue(data, []string{"user", "score"}); result != "95.500000" {
		t.Errorf("Expected '95.500000', got '%s'", result)
	}

	// Test existing boolean field
	if result := getValue(data, []string{"user", "active"}); result != "true" {
		t.Errorf("Expected 'true', got '%s'", result)
	}

	// Test null field
	if result := getValue(data, []string{"nullValue"}); result != "" {
		t.Errorf("Expected empty string for null, got '%s'", result)
	}

	// Test non-existent field
	if result := getValue(data, []string{"nonexistent", "field"}); result != "" {
		t.Errorf("Expected empty string for non-existent field, got '%s'", result)
	}
}

// Test BatchWriter functionality
func TestBatchWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	batchWriter := NewBatchWriter(writer, 2)

	// Write first record (should not flush yet)
	batchWriter.Write([]string{"col1", "col2"})
	if buf.Len() != 0 {
		t.Error("Buffer should be empty before batch size is reached")
	}

	// Write second record (should flush)
	batchWriter.Write([]string{"col3", "col4"})
	if buf.Len() == 0 {
		t.Error("Buffer should contain data after batch size is reached")
	}

	// Explicit flush
	batchWriter.Write([]string{"col5", "col6"})
	batchWriter.Flush()

	// Verify content
	content := buf.String()
	expected := `col1,col2
col3,col4
col5,col6
`
	if content != expected {
		t.Errorf("Expected %q, got %q", expected, content)
	}
}

// Test Do function - main functionality
func TestDo(t *testing.T) {
	// Create temporary input file
	tmpInput, err := os.CreateTemp("", "test_input_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer os.Remove(tmpInput.Name())

	jsonData := `{"user": {"name": "john", "age": 30}, "ip": "127.0.0.1"}
{"user": {"name": "jane", "age": 25}, "ip": "192.168.1.1"}
`

	if _, err := tmpInput.WriteString(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpInput.Close()

	// Create temporary output file
	tmpOutput, err := os.CreateTemp("", "test_output_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer os.Remove(tmpOutput.Name())
	tmpOutput.Close()

	// Test Do function
	count, err := Do(tmpInput.Name(), tmpOutput.Name(), ",", false, []string{"user.name", "ip"}, false)
	if err != nil {
		t.Fatalf("Do function failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 lines processed, got %d", count)
	}

	// Verify output content
	outputData, err := os.ReadFile(tmpOutput.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedOutput := `john,127.0.0.1
jane,192.168.1.1
`
	if string(outputData) != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, string(outputData))
	}
}

// Test Do function with all keys
func TestDoWithAllKeys(t *testing.T) {
	// Create temporary input file
	tmpInput, err := os.CreateTemp("", "test_input_all_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer os.Remove(tmpInput.Name())

	jsonData := `{"user": {"name": "john", "age": 30}, "ip": "127.0.0.1"}
{"user": {"name": "jane", "age": 25}, "score": 95.5}
`

	if _, err := tmpInput.WriteString(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpInput.Close()

	// Create temporary output file
	tmpOutput, err := os.CreateTemp("", "test_output_all_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer os.Remove(tmpOutput.Name())
	tmpOutput.Close()

	// Test Do function with all keys
	count, err := Do(tmpInput.Name(), tmpOutput.Name(), ",", true, nil, true)
	if err != nil {
		t.Fatalf("Do function failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 lines processed, got %d", count)
	}

	// Verify output content contains header
	outputData, err := os.ReadFile(tmpOutput.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	content := string(outputData)
	if !strings.Contains(content, "ip,user.age,user.name") {
		t.Error("Output should contain header with all keys")
	}
}

// Test Do function with stdin/stdout
func TestDoWithStdIO(t *testing.T) {
	// Save original stdin/stdout
	origStdin := os.Stdin
	origStdout := os.Stdout
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()

	// Create test input
	jsonData := `{"name": "john", "age": 30}
{"name": "jane", "age": 25}
`

	// Redirect stdin
	tmpInput, _ := os.CreateTemp("", "test_stdin_*.json")
	tmpInput.WriteString(jsonData)
	tmpInput.Seek(0, 0)
	os.Stdin = tmpInput

	// Redirect stdout to capture output
	var buf bytes.Buffer
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test Do function with stdio
	count, err := Do(Stdin, Stdout, ",", false, []string{"name", "age"}, false)

	// Close writer and read from pipe
	w.Close()
	os.Stdout = origStdout

	io.Copy(&buf, r)
	tmpInput.Close()
	os.Remove(tmpInput.Name())

	if err != nil {
		t.Fatalf("Do function failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 lines processed, got %d", count)
	}

	expectedOutput := `john,30
jane,25
`
	if buf.String() != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, buf.String())
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	// Test empty JSON array
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	lineReader := &stringLineReader{reader: strings.NewReader("")}

	count, err := json2csv(lineReader, writer, false, []string{"field"}, false)
	if err != nil {
		t.Fatalf("json2csv failed on empty input: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 records for empty input, got %d", count)
	}

	// Test invalid JSON
	lineReader = &stringLineReader{reader: strings.NewReader("{invalid json}\n")}
	buf.Reset()
	writer = csv.NewWriter(&buf)

	count, err = json2csv(lineReader, writer, false, []string{"field"}, false)
	if err != nil {
		t.Fatalf("json2csv failed on invalid JSON: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 line processed (even if invalid), got %d", count)
	}
}

// Helper functions
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

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

// BenchmarkFlattenKeys compares the performance of custom vs external flatmap
func BenchmarkFlattenKeys(b *testing.B) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "john",
			"age":  30,
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "New York",
				"zip":    "10001",
			},
		},
		"remote_ip": "127.0.0.1",
		"timestamp": "2023-01-01T00:00:00Z",
		"metadata": map[string]interface{}{
			"source": "api",
			"version": "1.0",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := flattenKeys(data)
		if err != nil {
			b.Fatal(err)
		}
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