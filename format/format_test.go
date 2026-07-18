﻿package format

import (
	"reflect"
	"testing"
)

func TestDatatypeEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		dt   DatatypeMessage
	}{
		{
			name: "integer 32-bit signed",
			dt: DatatypeMessage{
				Version:   DatatypeVersion2,
				Class:     ClassInteger,
				ByteOrder: OrderLittleEndian,
				Size:      4,
				Sign:      1,
			},
		},
		{
			name: "float 32-bit",
			dt: DatatypeMessage{
				Version:     DatatypeVersion2,
				Class:       ClassFloat,
				ByteOrder:   OrderLittleEndian,
				Size:        4,
				FloatBits:   24,
			},
		},
		{
			name: "float 64-bit",
			dt: DatatypeMessage{
				Version:     DatatypeVersion2,
				Class:       ClassFloat,
				ByteOrder:   OrderLittleEndian,
				Size:        8,
				FloatBits:   53,
			},
		},
		{
			name: "string null-terminated",
			dt: DatatypeMessage{
				Version:       DatatypeVersion2,
				Class:         ClassString,
				ByteOrder:     OrderLittleEndian,
				Size:          64,
				StringPadding: 0,
				StringSize:    64,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeDatatypeMessage(tc.dt)
			if err != nil {
				t.Fatalf("EncodeDatatypeMessage failed: %v", err)
			}

			decoded, err := DecodeDatatypeMessage(encoded)
			if err != nil {
				t.Fatalf("DecodeDatatypeMessage failed: %v", err)
			}

			if decoded.Version != tc.dt.Version {
				t.Errorf("Version mismatch: got %d, want %d", decoded.Version, tc.dt.Version)
			}
			if decoded.Class != tc.dt.Class {
				t.Errorf("Class mismatch: got %d, want %d", decoded.Class, tc.dt.Class)
			}
			if decoded.ByteOrder != tc.dt.ByteOrder {
				t.Errorf("ByteOrder mismatch: got %d, want %d", decoded.ByteOrder, tc.dt.ByteOrder)
			}
			if decoded.Size != tc.dt.Size {
				t.Errorf("Size mismatch: got %d, want %d", decoded.Size, tc.dt.Size)
			}
			if decoded.Class == ClassInteger && decoded.Sign != tc.dt.Sign {
				t.Errorf("Sign mismatch: got %d, want %d", decoded.Sign, tc.dt.Sign)
			}
			if decoded.Class == ClassFloat && decoded.FloatBits != tc.dt.FloatBits {
				t.Errorf("FloatBits mismatch: got %d, want %d", decoded.FloatBits, tc.dt.FloatBits)
			}
			if decoded.Class == ClassString {
				if decoded.StringPadding != tc.dt.StringPadding {
					t.Errorf("StringPadding mismatch: got %d, want %d", decoded.StringPadding, tc.dt.StringPadding)
				}
				if decoded.StringSize != tc.dt.StringSize {
					t.Errorf("StringSize mismatch: got %d, want %d", decoded.StringSize, tc.dt.StringSize)
				}
			}
		})
	}
}

func TestDataspaceEncodeDecode(t *testing.T) {
	testCases := []struct {
		name    string
		dims    []uint64
		maxDims []uint64
	}{
		{
			name:    "scalar",
			dims:    []uint64{},
			maxDims: []uint64{},
		},
		{
			name:    "1D array",
			dims:    []uint64{10},
			maxDims: []uint64{10},
		},
		{
			name:    "2D array",
			dims:    []uint64{10, 20},
			maxDims: []uint64{10, 20},
		},
		{
			name:    "3D array",
			dims:    []uint64{10, 20, 30},
			maxDims: []uint64{10, 20, 30},
		},
		{
			name:    "1D array with unlimited max",
			dims:    []uint64{10},
			maxDims: []uint64{0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeDataspaceMessage(tc.dims, tc.maxDims)
			if err != nil {
				t.Fatalf("EncodeDataspaceMessage failed: %v", err)
			}

			decoded, err := DecodeDataspaceMessage(encoded)
			if err != nil {
				t.Fatalf("DecodeDataspaceMessage failed: %v", err)
			}

			if len(tc.dims) == 0 {
				if decoded.Type != DataspaceScalar {
					t.Errorf("Type mismatch for scalar: got %d, want %d", decoded.Type, DataspaceScalar)
				}
			} else {
				if decoded.Type != DataspaceSimple {
					t.Errorf("Type mismatch for simple: got %d, want %d", decoded.Type, DataspaceSimple)
				}
				if !reflect.DeepEqual(decoded.Dims, tc.dims) {
					t.Errorf("Dims mismatch: got %v, want %v", decoded.Dims, tc.dims)
				}
				if !reflect.DeepEqual(decoded.MaxDims, tc.maxDims) {
					t.Errorf("MaxDims mismatch: got %v, want %v", decoded.MaxDims, tc.maxDims)
				}
			}
		})
	}
}

func TestLayoutEncodeDecode(t *testing.T) {
	testCases := []struct {
		name        string
		layoutClass uint8
		dataSize    uint64
		dataAddress uint64
		chunkDims   []uint64
	}{
		{
			name:        "contiguous",
			layoutClass: LayoutContiguous,
			dataSize:    40,
			dataAddress: 1000,
			chunkDims:   nil,
		},
		{
			name:        "compact",
			layoutClass: LayoutCompact,
			dataSize:    16,
			dataAddress: 0,
			chunkDims:   nil,
		},
		{
			name:        "chunked 1D",
			layoutClass: LayoutChunked,
			dataSize:    1000,
			dataAddress: 2000,
			chunkDims:   []uint64{10},
		},
		{
			name:        "chunked 2D",
			layoutClass: LayoutChunked,
			dataSize:    1000,
			dataAddress: 2000,
			chunkDims:   []uint64{10, 20},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeLayoutMessage(tc.layoutClass, tc.dataSize, tc.dataAddress, tc.chunkDims)
			if err != nil {
				t.Fatalf("EncodeLayoutMessage failed: %v", err)
			}

			decoded, err := DecodeLayoutMessage(encoded)
			if err != nil {
				t.Fatalf("DecodeLayoutMessage failed: %v", err)
			}

			if decoded.LayoutClass != tc.layoutClass {
				t.Errorf("LayoutClass mismatch: got %d, want %d", decoded.LayoutClass, tc.layoutClass)
			}
			if decoded.DataSize != tc.dataSize {
				t.Errorf("DataSize mismatch: got %d, want %d", decoded.DataSize, tc.dataSize)
			}
			if decoded.DataAddress != tc.dataAddress {
				t.Errorf("DataAddress mismatch: got %d, want %d", decoded.DataAddress, tc.dataAddress)
			}
			if tc.layoutClass == LayoutChunked {
				if !reflect.DeepEqual(decoded.ChunkDims, tc.chunkDims) {
					t.Errorf("ChunkDims mismatch: got %v, want %v", decoded.ChunkDims, tc.chunkDims)
				}
			}
		})
	}
}

