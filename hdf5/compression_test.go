package hdf5

import (
	"testing"
)

func TestCompressionFunctions(t *testing.T) {
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Test LZF compression
	compressed, err := compressData(data, "lzf")
	if err != nil {
		t.Fatalf("LZF compression failed: %v", err)
	}
	t.Logf("LZF compression: original=%d, compressed=%d", len(data), len(compressed))

	decompressed, err := decompressData(compressed, "lzf", len(data))
	if err != nil {
		t.Fatalf("LZF decompression failed: %v", err)
	}
	t.Logf("LZF decompression: decompressed=%d", len(decompressed))

	if string(decompressed) != string(data) {
		t.Fatalf("LZF decompression mismatch")
	}

	// Test ZSTD compression
	compressed, err = compressData(data, "zstd")
	if err != nil {
		t.Fatalf("ZSTD compression failed: %v", err)
	}
	t.Logf("ZSTD compression: original=%d, compressed=%d", len(data), len(compressed))

	decompressed, err = decompressData(compressed, "zstd", len(data))
	if err != nil {
		t.Fatalf("ZSTD decompression failed: %v", err)
	}
	t.Logf("ZSTD decompression: decompressed=%d", len(decompressed))

	if string(decompressed) != string(data) {
		t.Fatalf("ZSTD decompression mismatch")
	}

	t.Log("Compression functions test passed")
}