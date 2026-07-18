package hdf5

import (
	"os"
	"testing"
)

func TestDebugSimpleCompound(t *testing.T) {
	testFile := "test_debug_simple_compound.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	compoundDtype := Datatype{
		Class: DatatypeCompound,
		Size:  12,
		Fields: []CompoundField{
			{Name: "ID", Offset: 0, Datatype: DatatypeInt32},
			{Name: "Value", Offset: 4, Datatype: DatatypeFloat64},
		},
	}

	t.Logf("Compound dtype: Class=%d, Size=%d", compoundDtype.Class, compoundDtype.Size)
	for i, field := range compoundDtype.Fields {
		t.Logf("Field %d: Name=%s, Offset=%d, Class=%d, Size=%d", i, field.Name, field.Offset, field.Datatype.Class, field.Datatype.Size)
	}

	compoundSpace := Dataspace{Dims: []uint64{1}}
	compoundDset, err := f.CreateDataset("compound_data", compoundDtype, compoundSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create compound dataset: %v", err)
	}

	compoundData := []struct {
		ID    int32
		Value float64
	}{
		{1, 10.5},
	}
	err = compoundDset.Write(compoundData)
	if err != nil {
		t.Fatalf("Failed to write compound dataset: %v", err)
	}
	compoundDset.Close()

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	f2, err := OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f2.Close()

	compoundDset2, err := f2.GetDataset("compound_data")
	if err != nil {
		t.Fatalf("Failed to get compound dataset: %v", err)
	}

	var readCompoundData []struct {
		ID    int32
		Value float64
	}
	err = compoundDset2.Read(&readCompoundData)
	if err != nil {
		t.Fatalf("Failed to read compound dataset: %v", err)
	}

	t.Logf("Read compound data: ID=%d, Value=%f", readCompoundData[0].ID, readCompoundData[0].Value)

	if readCompoundData[0].ID != compoundData[0].ID || readCompoundData[0].Value != compoundData[0].Value {
		t.Errorf("Compound data: expected (%d, %f), got (%d, %f)",
			compoundData[0].ID, compoundData[0].Value,
			readCompoundData[0].ID, readCompoundData[0].Value)
	}
	compoundDset2.Close()

	t.Log("Simple compound dataset test passed!")
}

func TestDebugRawDataRead(t *testing.T) {
	testFile := "test_debug_raw_data.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	compoundDtype := Datatype{
		Class: DatatypeCompound,
		Size:  12,
		Fields: []CompoundField{
			{Name: "ID", Offset: 0, Datatype: DatatypeInt32},
			{Name: "Value", Offset: 4, Datatype: DatatypeFloat64},
		},
	}

	compoundSpace := Dataspace{Dims: []uint64{1}}
	compoundDset, err := f.CreateDataset("compound_data", compoundDtype, compoundSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create compound dataset: %v", err)
	}

	compoundData := []struct {
		ID    int32
		Value float64
	}{
		{1, 10.5},
	}
	err = compoundDset.Write(compoundData)
	if err != nil {
		t.Fatalf("Failed to write compound dataset: %v", err)
	}
	compoundDset.Close()

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	f2, err := OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f2.Close()

	compoundDset2, err := f2.GetDataset("compound_data")
	if err != nil {
		t.Fatalf("Failed to get compound dataset: %v", err)
	}
	defer compoundDset2.Close()

	var readCompoundData []struct {
		ID    int32
		Value float64
	}
	err = compoundDset2.Read(&readCompoundData)
	if err != nil {
		t.Fatalf("Failed to read compound data: %v", err)
	}

	t.Logf("Read compound data: ID=%d, Value=%f", readCompoundData[0].ID, readCompoundData[0].Value)

	if readCompoundData[0].ID != compoundData[0].ID || readCompoundData[0].Value != compoundData[0].Value {
		t.Errorf("Compound data: expected (%d, %f), got (%d, %f)",
			compoundData[0].ID, compoundData[0].Value,
			readCompoundData[0].ID, readCompoundData[0].Value)
	}
}
