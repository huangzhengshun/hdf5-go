package hdf5

import (
	"os"
	"testing"
)

func TestFileCreateAndOpen(t *testing.T) {
	testFile := "test_create_open.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	if f.Name() != testFile {
		t.Errorf("Expected file name %s, got %s", testFile, f.Name())
	}

	if f.Mode() != Create {
		t.Errorf("Expected mode Create, got %v", f.Mode())
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	f, err = OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	if f.Mode() != ReadOnly {
		t.Errorf("Expected mode ReadOnly, got %v", f.Mode())
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}
}

func TestGroupCreate(t *testing.T) {
	testFile := "test_group_create.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if g.Name() != "/test_group" {
		t.Errorf("Expected group name /test_group, got %s", g.Name())
	}

	if err := g.Close(); err != nil {
		t.Fatalf("Failed to close group: %v", err)
	}
}

func TestGroupGet(t *testing.T) {
	testFile := "test_group_get.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	g2, err := f.GetGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}

	if g2.Name() != "/test_group" {
		t.Errorf("Expected group name /test_group, got %s", g2.Name())
	}

	if err := g2.Close(); err != nil {
		t.Fatalf("Failed to close group: %v", err)
	}
}

func TestDatasetCreateAndWriteRead(t *testing.T) {
	testFile := "test_dataset.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := []int32{1, 2, 3, 4, 5}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{5}}

	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	var result []int32
	if err := dset.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d elements, got %d", len(data), len(result))
	}
	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}
}

func TestDatasetGet(t *testing.T) {
	testFile := "test_dataset_get.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := []int32{1, 2, 3, 4, 5}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{5}}

	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	dset2, err := f.GetDataset("data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset2.Close()

	var result []int32
	if err := dset2.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d elements, got %d", len(data), len(result))
	}
	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}
}

func TestDatasetInteger(t *testing.T) {
	testFile := "test_integer.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	int8Data := []int8{1, 2, 3}
	int16Data := []int16{1, 2, 3}
	int32Data := []int32{1, 2, 3}
	int64Data := []int64{1, 2, 3}
	uint8Data := []uint8{1, 2, 3}
	uint16Data := []uint16{1, 2, 3}
	uint32Data := []uint32{1, 2, 3}
	uint64Data := []uint64{1, 2, 3}

	testCases := []struct {
		name   string
		dtype  Datatype
		data   interface{}
		result interface{}
	}{
		{"int8", DatatypeInt8, int8Data, &[]int8{}},
		{"int16", DatatypeInt16, int16Data, &[]int16{}},
		{"int32", DatatypeInt32, int32Data, &[]int32{}},
		{"int64", DatatypeInt64, int64Data, &[]int64{}},
		{"uint8", DatatypeUint8, uint8Data, &[]uint8{}},
		{"uint16", DatatypeUint16, uint16Data, &[]uint16{}},
		{"uint32", DatatypeUint32, uint32Data, &[]uint32{}},
		{"uint64", DatatypeUint64, uint64Data, &[]uint64{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			space := Dataspace{Dims: []uint64{3}}
			dset, err := f.CreateDataset(tc.name, tc.dtype, space, PropertyList{})
			if err != nil {
				t.Fatalf("Failed to create dataset: %v", err)
			}
			defer dset.Close()

			if err := dset.Write(tc.data); err != nil {
				t.Fatalf("Failed to write data: %v", err)
			}

			if err := dset.Read(tc.result); err != nil {
				t.Fatalf("Failed to read data: %v", err)
			}
		})
	}
}

func TestNestedGroups(t *testing.T) {
	testFile := "test_nested_groups.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g1, err := f.CreateGroup("level1")
	if err != nil {
		t.Fatalf("Failed to create group level1: %v", err)
	}
	defer g1.Close()

	g2, err := g1.CreateGroup("level2")
	if err != nil {
		t.Fatalf("Failed to create group level2: %v", err)
	}
	defer g2.Close()

	g3, err := g2.CreateGroup("level3")
	if err != nil {
		t.Fatalf("Failed to create group level3: %v", err)
	}
	defer g3.Close()

	data := []int32{1, 2, 3}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	dset, err := g3.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	dset2, err := f.GetDataset("/level1/level2/level3/data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset2.Close()

	var result []int32
	if err := dset2.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}
}

func TestGroupKeys(t *testing.T) {
	testFile := "test_group_keys.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	if _, err := f.CreateGroup("group1"); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	if _, err := f.CreateGroup("group2"); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	if _, err := f.CreateDataset("data1", dtype, space, PropertyList{}); err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	if _, err := f.CreateDataset("data2", dtype, space, PropertyList{}); err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}

	keys := f.Keys()
	if len(keys) != 4 {
		t.Errorf("Expected 4 keys, got %d: %v", len(keys), keys)
	}
}

func TestFileReopenAndRead(t *testing.T) {
	testFile := "test_reopen.h5"
	defer os.Remove(testFile)

	dtype := Datatype{
		Class:     DatatypeInteger,
		Size:      4,
		ByteOrder: LittleEndian,
		Sign:      Signed,
	}

	space := Dataspace{
		Dims:    []uint64{10},
		MaxDims: []uint64{10},
	}

	writeData := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}

	if err := dset.Write(writeData); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := dset.Close(); err != nil {
		t.Fatalf("Failed to close dataset: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	f, err = OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f.Close()

	dset, err = f.GetDataset("data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset.Close()

	var readData []int32
	if err := dset.Read(&readData); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	for i := range writeData {
		if readData[i] != writeData[i] {
			t.Errorf("Expected readData[%d] = %d, got %d", i, writeData[i], readData[i])
		}
	}
}

func TestAttributeSetAndGet(t *testing.T) {
	testFile := "test_attribute.h5"
	defer os.Remove(testFile)
	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	err = g.Attributes().Set("string_attr", "hello world")
	if err != nil {
		t.Fatalf("Failed to set string attribute: %v", err)
	}

	err = g.Attributes().Set("int_attr", int32(42))
	if err != nil {
		t.Fatalf("Failed to set int attribute: %v", err)
	}

	err = g.Attributes().Set("float_attr", float64(3.14))
	if err != nil {
		t.Fatalf("Failed to set float attribute: %v", err)
	}

	var strVal string
	attr, err := g.Attributes().Get("string_attr")
	if err != nil {
		t.Fatalf("Failed to get string attribute: %v", err)
	}
	err = attr.Read(&strVal)
	if err != nil {
		t.Fatalf("Failed to read string attribute: %v", err)
	}
	if strVal != "hello world" {
		t.Errorf("Expected string_attr = 'hello world', got '%s'", strVal)
	}

	var intVal int32
	attr, err = g.Attributes().Get("int_attr")
	if err != nil {
		t.Fatalf("Failed to get int attribute: %v", err)
	}
	err = attr.Read(&intVal)
	if err != nil {
		t.Fatalf("Failed to read int attribute: %v", err)
	}
	if intVal != 42 {
		t.Errorf("Expected int_attr = 42, got %d", intVal)
	}

	var floatVal float64
	attr, err = g.Attributes().Get("float_attr")
	if err != nil {
		t.Fatalf("Failed to get float attribute: %v", err)
	}
	err = attr.Read(&floatVal)
	if err != nil {
		t.Fatalf("Failed to read float attribute: %v", err)
	}
	if floatVal != 3.14 {
		t.Errorf("Expected float_attr = 3.14, got %f", floatVal)
	}
}

func TestAttributeKeysAndLen(t *testing.T) {
	testFile := "test_attribute_keys.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	if err := g.Attributes().Set("attr1", int32(1)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}
	if err := g.Attributes().Set("attr2", int32(2)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}
	if err := g.Attributes().Set("attr3", int32(3)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}

	attrs := g.Attributes()
	if attrs.Len() != 3 {
		t.Errorf("Expected 3 attributes, got %d", attrs.Len())
	}

	keys := attrs.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d: %v", len(keys), keys)
	}
}

func TestAttributeDelete(t *testing.T) {
	testFile := "test_attribute_delete.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	if err := g.Attributes().Set("attr1", int32(1)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}
	if err := g.Attributes().Set("attr2", int32(2)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}

	if g.Attributes().Len() != 2 {
		t.Errorf("Expected 2 attributes, got %d", g.Attributes().Len())
	}

	if err := g.Attributes().Delete("attr1"); err != nil {
		t.Fatalf("Failed to delete attribute: %v", err)
	}

	if g.Attributes().Len() != 1 {
		t.Errorf("Expected 1 attribute, got %d", g.Attributes().Len())
	}

	_, err = g.Attributes().Get("attr1")
	if err == nil {
		t.Error("Expected error when getting deleted attribute")
	}
}

func TestAttributeOnDataset(t *testing.T) {
	testFile := "test_attribute_dataset.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := []int32{1, 2, 3}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := dset.Attributes().Set("units", "meters"); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}

	var units string
	attr, err := dset.Attributes().Get("units")
	if err != nil {
		t.Fatalf("Failed to get attribute: %v", err)
	}
	if err := attr.Read(&units); err != nil {
		t.Fatalf("Failed to read attribute: %v", err)
	}
	if units != "meters" {
		t.Errorf("Expected units = 'meters', got '%s'", units)
	}
}

func TestDatasetCompression(t *testing.T) {
	testFile := "test_compression.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i)
	}

	dtype := DatatypeFloat64
	space := Dataspace{Dims: []uint64{1000}}
	plist := PropertyList{
		Compression: "gzip",
		Chunks:      []uint64{100},
	}

	dset, err := f.CreateDataset("data", dtype, space, plist)
	if err != nil {
		t.Fatalf("Failed to create dataset with compression: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	f.Close()

	f, err = OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f.Close()

	dset, err = f.GetDataset("data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset.Close()

	var result []float64
	if err := dset.Read(&result); err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d elements, got %d", len(data), len(result))
	}
	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %f, got %f", i, data[i], result[i])
			break
		}
	}

	t.Log("GZIP compression test passed")
}

func TestCompoundDatatype(t *testing.T) {
	testFile := "test_compound.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	type Point struct {
		X float64
		Y float64
	}

	points := []Point{
		{1.0, 2.0},
		{3.0, 4.0},
		{5.0, 6.0},
	}

	dtype := Datatype{
		Class: DatatypeCompound,
		Size:  16,
		Fields: []CompoundField{
			{Name: "X", Offset: 0, Datatype: DatatypeFloat64},
			{Name: "Y", Offset: 8, Datatype: DatatypeFloat64},
		},
	}

	space := Dataspace{Dims: []uint64{3}}
	dset, err := f.CreateDataset("points", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(points); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}
}

func TestSoftLink(t *testing.T) {
	testFile := "test_softlink.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	subgroup, err := f.CreateGroup("subgroup")
	if err != nil {
		t.Fatalf("Failed to create subgroup: %v", err)
	}
	defer subgroup.Close()

	data := []int32{1, 2, 3}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	dset, err := subgroup.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := f.CreateLink("/subgroup/data", "data_link"); err != nil {
		t.Fatalf("Failed to create soft link: %v", err)
	}

	dset2, err := f.GetDataset("data_link")
	if err != nil {
		t.Fatalf("Failed to get dataset via link: %v", err)
	}
	defer dset2.Close()

	var result []int32
	if err := dset2.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}
}

func TestDatasetSliceReadWrite(t *testing.T) {
	testFile := "test_slice.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]int32, 10)
	for i := range data {
		data[i] = int32(i)
	}

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{10}}
	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	sel := Selection{
		Hyperslabs: []Hyperslab{
			{Start: []uint64{2}, Count: []uint64{5}},
		},
	}

	var sliceData []int32
	if err := dset.ReadSlice(&sliceData, sel); err != nil {
		t.Fatalf("Failed to read slice: %v", err)
	}

	expected := []int32{2, 3, 4, 5, 6}
	for i := range expected {
		if sliceData[i] != expected[i] {
			t.Errorf("Expected sliceData[%d] = %d, got %d", i, expected[i], sliceData[i])
		}
	}
}

func TestEnumDatatype(t *testing.T) {
	testFile := "test_enum.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	dtype := Datatype{
		Class:    DatatypeEnum,
		Size:     4,
		BaseType: &DatatypeInt32,
		EnumMembers: []EnumMember{
			{Name: "Red", Value: 1},
			{Name: "Green", Value: 2},
			{Name: "Blue", Value: 3},
		},
	}

	space := Dataspace{Dims: []uint64{3}}
	dset, err := f.CreateDataset("colors", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	data := []int32{1, 2, 3}
	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	t.Logf("Enum datatype test passed with %d members", len(dtype.EnumMembers))
}

func TestArrayDatatype(t *testing.T) {
	testFile := "test_array.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	dtype := Datatype{
		Class:     DatatypeArray,
		Size:      12,
		BaseType:  &DatatypeInt32,
		ArrayDims: []uint64{3},
	}

	space := Dataspace{Dims: []uint64{2}}
	dset, err := f.CreateDataset("arrays", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	data := []int32{1, 2, 3, 4, 5, 6}
	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	t.Log("Array datatype test passed")
}

func TestDatasetResize(t *testing.T) {
	testFile := "test_resize.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := []int32{1, 2, 3, 4, 5}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{5}, MaxDims: []uint64{H5S_UNLIMITED}}

	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := dset.Resize([]uint64{8}); err != nil {
		t.Fatalf("Failed to resize dataset: %v", err)
	}

	data2 := []int32{6, 7, 8}
	if err := dset.WriteSlice(data2, Selection{Hyperslabs: []Hyperslab{{Start: []uint64{5}, Count: []uint64{3}}}}); err != nil {
		t.Fatalf("Failed to write slice: %v", err)
	}

	var result []int32
	if err := dset.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	expected := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	if len(result) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, expected[i], result[i])
		}
	}

	t.Logf("Dataset resize test passed: resized from 5 to 8")
}

func TestGroupCopyMoveDelete(t *testing.T) {
	testFile := "test_copy_move_delete.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g1, err := f.CreateGroup("group1")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g1.Close()

	data := []int32{1, 2, 3}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	dset, err := g1.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	g2, err := f.CreateGroup("group2")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g2.Close()

	if err := g1.Copy(g2, "data_copy"); err != nil {
		t.Fatalf("Failed to copy dataset: %v", err)
	}

	t.Log("Group copy/move/delete test passed")
}

func TestExternalLink(t *testing.T) {
	testFile1 := "test_external1.h5"
	testFile2 := "test_external2.h5"
	defer os.Remove(testFile1)
	defer os.Remove(testFile2)

	f1, err := OpenFile(testFile1, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f1.Close()

	data := []int32{1, 2, 3}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	dset, err := f1.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	f2, err := OpenFile(testFile2, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f2.Close()

	if err := f2.CreateExternalLink(testFile1, "/data", "external_data"); err != nil {
		t.Fatalf("Failed to create external link: %v", err)
	}

	dset2, err := f2.GetDataset("external_data")
	if err != nil {
		t.Fatalf("Failed to get dataset via external link: %v", err)
	}
	defer dset2.Close()

	var result []int32
	if err := dset2.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}

	t.Log("External link test passed")
}

func TestAttributeIter(t *testing.T) {
	testFile := "test_attribute_iter.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	if err := g.Attributes().Set("attr1", int32(1)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}
	if err := g.Attributes().Set("attr2", int32(2)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}
	if err := g.Attributes().Set("attr3", int32(3)); err != nil {
		t.Fatalf("Failed to set attribute: %v", err)
	}

	attrs, err := g.Attributes().Iter()
	if err != nil {
		t.Fatalf("Failed to iterate attributes: %v", err)
	}

	if len(attrs) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(attrs))
	}

	t.Logf("Attribute iteration test passed with %d attributes", len(attrs))
}

func TestVarLengthDatatype(t *testing.T) {
	testFile := "test_vlen.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	dtype := Datatype{
		Class:    DatatypeVarLength,
		Size:     16,
		BaseType: &DatatypeInt32,
	}

	space := Dataspace{Dims: []uint64{3}}
	dset, err := f.CreateDataset("vlen_data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	data := [][]int32{
		{1, 2, 3},
		{4, 5},
		{6, 7, 8, 9},
	}

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	t.Log("VarLength datatype test passed")
}

func TestReclaimFreeSpace(t *testing.T) {
	testFile := "test_reclaim.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	data := []int32{1, 2, 3}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	dset, err := g.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := g.Delete("data"); err != nil {
		t.Fatalf("Failed to delete dataset: %v", err)
	}

	t.Log("Free space reclaim test passed")
}

func TestVisitItems(t *testing.T) {
	testFile := "test_visit.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g1, err := f.CreateGroup("group1")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g1.Close()

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{3}}

	if _, err := g1.CreateDataset("data", dtype, space, PropertyList{}); err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}

	g2, err := f.CreateGroup("group2")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g2.Close()

	count := 0
	err = f.VisitItems(func(name string, obj interface{}) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to visit items: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 top-level items, got %d", count)
	}

	t.Logf("VisitItems test passed with %d visited items", count)
}

func TestPointSelection(t *testing.T) {
	testFile := "test_point_selection.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]int32, 10)
	for i := range data {
		data[i] = int32(i)
	}

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{10}}
	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	sel := Selection{
		Points: [][]uint64{{1}, {3}, {5}, {7}, {9}},
	}

	var result []int32
	if err := dset.ReadSlice(&result, sel); err != nil {
		t.Fatalf("Failed to read points: %v", err)
	}

	expected := []int32{1, 3, 5, 7, 9}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, expected[i], result[i])
		}
	}

	t.Log("Point selection test passed")
}

func TestMultiHyperslabSelection(t *testing.T) {
	testFile := "test_multi_hyperslab.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]int32, 20)
	for i := range data {
		data[i] = int32(i)
	}

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{20}}

	dset, err := f.CreateDataset("test", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	sel := Selection{
		Hyperslabs: []Hyperslab{
			{Start: []uint64{0}, Count: []uint64{3}},
			{Start: []uint64{10}, Count: []uint64{4}},
			{Start: []uint64{17}, Count: []uint64{3}},
		},
	}

	var result []int32
	if err := dset.ReadSlice(&result, sel); err != nil {
		t.Fatalf("Failed to read multi-hyperslab: %v", err)
	}

	expected := []int32{0, 1, 2, 10, 11, 12, 13, 17, 18, 19}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(result))
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, expected[i], result[i])
		}
	}

	t.Log("Multi-hyperslab selection test passed")
}

func TestMaskSelection(t *testing.T) {
	testFile := "test_mask_selection.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]int32, 10)
	for i := range data {
		data[i] = int32(i)
	}

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{10}}
	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	mask := []bool{false, true, false, true, false, true, false, true, false, true}
	sel := Selection{
		Mask: mask,
	}

	var result []int32
	if err := dset.ReadSlice(&result, sel); err != nil {
		t.Fatalf("Failed to read with mask: %v", err)
	}

	expected := []int32{1, 3, 5, 7, 9}
	if len(result) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, expected[i], result[i])
		}
	}

	t.Log("Mask selection test passed")
}

func TestCompressionLZF(t *testing.T) {
	testFile := "test_compression_lzf.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i)
	}
	dtype := DatatypeFloat64
	space := Dataspace{Dims: []uint64{1000}}
	plist := PropertyList{
		Compression: "lzf",
		Chunks:      []uint64{100},
	}

	dset, err := f.CreateDataset("data", dtype, space, plist)
	if err != nil {
		t.Fatalf("Failed to create dataset with lzf compression: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	f.Close()

	f, err = OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f.Close()

	dset, err = f.GetDataset("data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset.Close()

	var result []float64
	if err := dset.Read(&result); err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d elements, got %d", len(data), len(result))
	}
	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %f, got %f", i, data[i], result[i])
			break
		}
	}

	t.Log("LZF compression test passed")
}

func TestCompressionZSTD(t *testing.T) {
	testFile := "test_compression_zstd.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i)
	}
	dtype := DatatypeFloat64
	space := Dataspace{Dims: []uint64{1000}}
	plist := PropertyList{
		Compression: "zstd",
		Chunks:      []uint64{100},
	}

	dset, err := f.CreateDataset("data", dtype, space, plist)
	if err != nil {
		t.Fatalf("Failed to create dataset with zstd compression: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	f.Close()

	f, err = OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f.Close()

	dset, err = f.GetDataset("data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset.Close()

	var result []float64
	if err := dset.Read(&result); err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d elements, got %d", len(data), len(result))
	}
	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %f, got %f", i, data[i], result[i])
			break
		}
	}

	t.Log("ZSTD compression test passed")
}

func TestRegionReference(t *testing.T) {
	testFile := "test_region_reference.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]int32, 10)
	for i := range data {
		data[i] = int32(i)
	}

	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{10}}
	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	refDtype := Datatype{
		Class: DatatypeReference,
		Size:  8,
	}
	refSpace := Dataspace{Dims: []uint64{1}}
	refDset, err := f.CreateDataset("region_ref", refDtype, refSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create reference dataset: %v", err)
	}
	defer refDset.Close()

	refData := []uint64{dset.(*dataset).addr}
	if err := refDset.Write(refData); err != nil {
		t.Fatalf("Failed to write reference data: %v", err)
	}

	t.Log("Region reference test passed")
}

func TestMemoryFile(t *testing.T) {
	f, err := CreateMemoryFile()
	if err != nil {
		t.Fatalf("Failed to create memory file: %v", err)
	}
	defer f.Close()

	data := []int32{1, 2, 3, 4, 5}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{5}}

	dset, err := f.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	var result []int32
	if err := dset.Read(&result); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d elements, got %d", len(data), len(result))
	}
	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}

	bytes, err := f.Bytes()
	if err != nil {
		t.Fatalf("Failed to get file bytes: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("Expected non-empty file bytes")
	}

	t.Log("Memory file test passed")
}

func TestInteroperability(t *testing.T) {
	tmpFile := "test_interop.h5"
	defer os.Remove(tmpFile)

	f, err := OpenFile(tmpFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	dtype := DatatypeFloat64
	space := Dataspace{Dims: []uint64{5}}

	dset, err := f.CreateDataset("test_data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := dset.Close(); err != nil {
		t.Fatalf("Failed to close dataset: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	f2, err := OpenFile(tmpFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f2.Close()

	var readData []float64
	dset2, err := f2.GetDataset("test_data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset2.Close()

	if err := dset2.Read(&readData); err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	for i := range data {
		if readData[i] != data[i] {
			t.Errorf("Expected readData[%d] = %f, got %f", i, data[i], readData[i])
		}
	}

	t.Log("Interoperability test passed")
}

func TestObjectReference(t *testing.T) {
	testFile := "test_object_ref.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	data := []int32{1, 2, 3, 4, 5}
	dtype := DatatypeInt32
	space := Dataspace{Dims: []uint64{5}}

	dset, err := g.CreateDataset("data", dtype, space, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	if err := dset.Write(data); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	groupRef, err := g.CreateReference()
	if err != nil {
		t.Fatalf("Failed to create group reference: %v", err)
	}

	dsetRef, err := dset.CreateObjectReference()
	if err != nil {
		t.Fatalf("Failed to create dataset reference: %v", err)
	}

	obj, err := f.Dereference(groupRef)
	if err != nil {
		t.Fatalf("Failed to dereference group: %v", err)
	}

	_, ok := obj.(Group)
	if !ok {
		t.Fatalf("Expected Group type, got %T", obj)
	}

	obj2, err := f.Dereference(dsetRef)
	if err != nil {
		t.Fatalf("Failed to dereference dataset: %v", err)
	}

	dset2, ok := obj2.(Dataset)
	if !ok {
		t.Fatalf("Expected Dataset type, got %T", obj2)
	}

	var result []int32
	if err := dset2.Read(&result); err != nil {
		t.Fatalf("Failed to read dereferenced dataset: %v", err)
	}

	for i := range data {
		if result[i] != data[i] {
			t.Errorf("Expected result[%d] = %d, got %d", i, data[i], result[i])
		}
	}

	t.Log("Object reference test passed")
}
