package hdf5

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/huangzhengshun/hdf5-go/format"
)

func TestVarLengthDebug(t *testing.T) {
	testFile := "test_varlen_debug.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	varlenDtype := Datatype{Class: DatatypeVarLength, Size: 4, BaseType: &DatatypeInt32}

	dset, err := f.CreateDataset("varlength", varlenDtype, Dataspace{Dims: []uint64{1}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create varlength dataset: %v", err)
	}
	dset.Close()

	f.Close()

	f2, err := OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f2.Close()

	dset2, err := f2.GetDataset("varlength")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset2.Close()

	ds := dset2.(*dataset)
	t.Logf("Dataset dtype.Class: %d", ds.dtype.Class)
	t.Logf("Dataset dtype.Size: %d", ds.dtype.Size)
	t.Logf("Dataset dtype.BaseType: %v", ds.dtype.BaseType)
	if ds.dtype.BaseType != nil {
		t.Logf("BaseType.Class: %d", ds.dtype.BaseType.Class)
		t.Logf("BaseType.Size: %d", ds.dtype.BaseType.Size)
	}

	for _, msg := range ds.header.Messages {
		if msg.Type == format.MessageDatatype {
			t.Logf("Datatype message size: %d", msg.Size)
			t.Logf("Datatype message content length: %d", len(msg.Content))
			t.Logf("Datatype message content hex: %s", hex.EncodeToString(msg.Content))

			dtDecoded, err := format.DecodeDatatypeMessage(msg.Content)
			if err != nil {
				t.Fatalf("Failed to decode datatype message: %v", err)
			}

			t.Logf("Decoded Class: %d", dtDecoded.Class)
			t.Logf("Decoded Size: %d", dtDecoded.Size)
			t.Logf("Decoded ArrayBase: %v", dtDecoded.ArrayBase)
			if dtDecoded.ArrayBase != nil {
				t.Logf("ArrayBase.Class: %d", dtDecoded.ArrayBase.Class)
				t.Logf("ArrayBase.Size: %d", dtDecoded.ArrayBase.Size)
			}
		}
	}
}
