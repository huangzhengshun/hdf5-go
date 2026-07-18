package hdf5

import (
	"math"
	"os"
	"sync"
	"testing"
	"time"
)

func TestEmptyDataset(t *testing.T) {
	testFile := "test_empty_dataset.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	emptySpace := Dataspace{Dims: []uint64{0}}
	dset, err := f.CreateDataset("empty", DatatypeInt32, emptySpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create empty dataset: %v", err)
	}
	defer dset.Close()

	var readData []int32
	err = dset.Read(&readData)
	if err != nil {
		t.Fatalf("Failed to read empty dataset: %v", err)
	}
	if len(readData) != 0 {
		t.Errorf("Expected empty slice, got %d elements", len(readData))
	}
}

func TestZeroDimensionDataset(t *testing.T) {
	testFile := "test_zero_dim.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	scalarSpace := Dataspace{Dims: []uint64{}}
	dset, err := f.CreateDataset("scalar", DatatypeInt32, scalarSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create scalar dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write(int32(42))
	if err != nil {
		t.Fatalf("Failed to write scalar value: %v", err)
	}

	var readValue int32
	err = dset.Read(&readValue)
	if err != nil {
		t.Fatalf("Failed to read scalar value: %v", err)
	}
	if readValue != 42 {
		t.Errorf("Expected 42, got %d", readValue)
	}
}

func TestLargeDataset(t *testing.T) {
	testFile := "test_large_dataset.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	largeSize := uint64(10000)
	largeSpace := Dataspace{Dims: []uint64{largeSize}}
	dset, err := f.CreateDataset("large", DatatypeFloat64, largeSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create large dataset: %v", err)
	}
	defer dset.Close()

	largeData := make([]float64, largeSize)
	for i := uint64(0); i < largeSize; i++ {
		largeData[i] = float64(i) * 0.1
	}

	err = dset.Write(largeData)
	if err != nil {
		t.Fatalf("Failed to write large dataset: %v", err)
	}

	var readData []float64
	err = dset.Read(&readData)
	if err != nil {
		t.Fatalf("Failed to read large dataset: %v", err)
	}

	if uint64(len(readData)) != largeSize {
		t.Errorf("Expected length %d, got %d", largeSize, len(readData))
	}

	for i := uint64(0); i < 100 && i < largeSize; i++ {
		if readData[i] != float64(i)*0.1 {
			t.Errorf("Data[%d]: expected %f, got %f", i, float64(i)*0.1, readData[i])
			break
		}
	}
}

func TestSpecialCharacterNames(t *testing.T) {
	testFile := "test_special_names.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	specialNames := []string{
		"group/with/slashes",
		"group with spaces",
		"group.with.dots",
		"group-with-dashes",
		"group_with_underscores",
		"GROUP_UPPERCASE",
		"group123numbers",
	}

	for _, name := range specialNames {
		g, err := f.CreateGroup(name)
		if err != nil {
			t.Fatalf("Failed to create group '%s': %v", name, err)
		}

		dset, err := g.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{5}}, PropertyList{})
		if err != nil {
			t.Fatalf("Failed to create dataset in group '%s': %v", name, err)
		}
		dset.Write([]int32{1, 2, 3, 4, 5})
		dset.Close()
		g.Close()
	}

	for _, name := range specialNames {
		g, err := f.GetGroup(name)
		if err != nil {
			t.Fatalf("Failed to get group '%s': %v", name, err)
		}

		dset, err := g.GetDataset("data")
		if err != nil {
			t.Fatalf("Failed to get dataset from group '%s': %v", name, err)
		}

		var data []int32
		err = dset.Read(&data)
		if err != nil {
			t.Fatalf("Failed to read dataset from group '%s': %v", name, err)
		}

		if len(data) != 5 {
			t.Errorf("Group '%s': expected 5 elements, got %d", name, len(data))
		}

		dset.Close()
		g.Close()
	}
}

func TestDeepNestedGroups(t *testing.T) {
	testFile := "test_deep_nested.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	current, err := f.CreateGroup("level0")
	if err != nil {
		t.Fatalf("Failed to create level0 group: %v", err)
	}

	for i := 1; i < 10; i++ {
		groupName := "level" + string(rune('0'+i))
		g, err := current.CreateGroup(groupName)
		if err != nil {
			t.Fatalf("Failed to create group at level %d: %v", i, err)
		}
		current.Close()
		current = g
	}

	current.Close()

	f.Close()

	f, err = OpenFile(testFile, ReadWrite)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f.Close()

	current, err = f.GetGroup("level0")
	if err != nil {
		t.Fatalf("Failed to get level0: %v", err)
	}

	for i := 1; i < 10; i++ {
		groupName := "level" + string(rune('0'+i))
		g, err := current.GetGroup(groupName)
		if err != nil {
			t.Fatalf("Failed to get %s: %v", groupName, err)
		}
		current.Close()
		current = g
	}

	dset, err := current.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{1}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset in deep group: %v", err)
	}
	dset.Write(int32(42))
	dset.Close()
	current.Close()
}

func TestConcurrentAccess(t *testing.T) {
	testFile := "test_concurrent.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	dset, err := f.CreateDataset("concurrent", DatatypeInt32, Dataspace{Dims: []uint64{100}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}

	data := make([]int32, 100)
	for i := range data {
		data[i] = int32(i)
	}
	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}
	dset.Close()

	err = f.Close()
	if err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			f, err := OpenFile(testFile, ReadOnly)
			if err != nil {
				errors <- err
				return
			}
			defer f.Close()

			dset, err := f.GetDataset("concurrent")
			if err != nil {
				errors <- err
				return
			}
			defer dset.Close()

			var data []int32
			err = dset.Read(&data)
			if err != nil {
				errors <- err
				return
			}

			if len(data) != 100 {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestAttributeAllTypes(t *testing.T) {
	testFile := "test_attr_all_types.h5"
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

	err = g.Attributes().Set("int_attr", int32(42))
	if err != nil {
		t.Fatalf("Failed to set int attribute: %v", err)
	}

	err = g.Attributes().Set("float_attr", float64(3.14))
	if err != nil {
		t.Fatalf("Failed to set float attribute: %v", err)
	}

	err = g.Attributes().Set("string_attr", "hello")
	if err != nil {
		t.Fatalf("Failed to set string attribute: %v", err)
	}

	err = g.Attributes().Set("int_array", []int32{1, 2, 3})
	if err != nil {
		t.Fatalf("Failed to set int array attribute: %v", err)
	}

	err = g.Attributes().Set("float_array", []float64{1.1, 2.2, 3.3})
	if err != nil {
		t.Fatalf("Failed to set float array attribute: %v", err)
	}

	var intVal int32
	attr, err := g.Attributes().Get("int_attr")
	if err != nil {
		t.Fatalf("Failed to get int attribute: %v", err)
	}
	err = attr.Read(&intVal)
	if err != nil {
		t.Fatalf("Failed to read int attribute: %v", err)
	}
	if intVal != 42 {
		t.Errorf("Expected 42, got %d", intVal)
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
		t.Errorf("Expected 3.14, got %f", floatVal)
	}

	var stringVal string
	attr, err = g.Attributes().Get("string_attr")
	if err != nil {
		t.Fatalf("Failed to get string attribute: %v", err)
	}
	err = attr.Read(&stringVal)
	if err != nil {
		t.Fatalf("Failed to read string attribute: %v", err)
	}
	if stringVal != "hello" {
		t.Errorf("Expected 'hello', got '%s'", stringVal)
	}

	var intArr []int32
	attr, err = g.Attributes().Get("int_array")
	if err != nil {
		t.Fatalf("Failed to get int array attribute: %v", err)
	}
	err = attr.Read(&intArr)
	if err != nil {
		t.Fatalf("Failed to read int array attribute: %v", err)
	}
	if len(intArr) != 3 || intArr[0] != 1 || intArr[1] != 2 || intArr[2] != 3 {
		t.Errorf("Expected [1,2,3], got %v", intArr)
	}

	var floatArr []float64
	attr, err = g.Attributes().Get("float_array")
	if err != nil {
		t.Fatalf("Failed to get float array attribute: %v", err)
	}
	err = attr.Read(&floatArr)
	if err != nil {
		t.Fatalf("Failed to read float array attribute: %v", err)
	}
	if len(floatArr) != 3 || floatArr[0] != 1.1 || floatArr[1] != 2.2 || floatArr[2] != 3.3 {
		t.Errorf("Expected [1.1,2.2,3.3], got %v", floatArr)
	}
}

func TestCompressionEdgeCases(t *testing.T) {
	testFile := "test_compression_edge.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	plist1 := NewPropertyList(PropertyListDatasetCreate)
	plist1.SetDeflate(9)

	plist2 := NewPropertyList(PropertyListDatasetCreate)
	plist2.SetLZF()

	plist3 := NewPropertyList(PropertyListDatasetCreate)
	plist3.SetZSTD(22)

	plists := []PropertyList{plist1, plist2, plist3}
	names := []string{"deflate_max", "lzf", "zstd_max"}

	for i, plist := range plists {
		data := make([]int32, 1000)
		for j := range data {
			data[j] = int32(j)
		}

		dset, err := f.CreateDataset(names[i], DatatypeInt32, Dataspace{Dims: []uint64{1000}}, plist)
		if err != nil {
			t.Fatalf("Failed to create compressed dataset %s: %v", names[i], err)
		}

		err = dset.Write(data)
		if err != nil {
			t.Fatalf("Failed to write compressed dataset %s: %v", names[i], err)
		}
		dset.Close()
	}

	for _, name := range names {
		dset, err := f.GetDataset(name)
		if err != nil {
			t.Fatalf("Failed to get compressed dataset %s: %v", name, err)
		}

		var readData []int32
		err = dset.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read compressed dataset %s: %v", name, err)
		}

		if len(readData) != 1000 {
			t.Errorf("Dataset %s: expected 1000 elements, got %d", name, len(readData))
		}

		for j := 0; j < 10; j++ {
			if readData[j] != int32(j) {
				t.Errorf("Dataset %s: data[%d] = %d, expected %d", name, j, readData[j], j)
				break
			}
		}

		dset.Close()
	}
}

func TestSelectionEdgeCases(t *testing.T) {
	testFile := "test_selection_edge.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	data := make([]int32, 100)
	for i := range data {
		data[i] = int32(i)
	}

	dset, err := f.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{100}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write dataset: %v", err)
	}

	sel := Selection{
		Hyperslabs: []Hyperslab{
			{Start: []uint64{10}, Count: []uint64{20}, Stride: []uint64{1}},
		},
	}

	var readData []int32
	err = dset.ReadSlice(&readData, sel)
	if err != nil {
		t.Fatalf("Failed to read selection: %v", err)
	}

	if len(readData) != 20 {
		t.Errorf("Expected 20 elements, got %d", len(readData))
	}

	for i := 0; i < 20; i++ {
		if readData[i] != int32(10+i) {
			t.Errorf("Data[%d] = %d, expected %d", i, readData[i], 10+i)
			break
		}
	}

	dset.Close()
}

func TestResizeMultipleTimes(t *testing.T) {
	testFile := "test_resize_multiple.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	plist := NewPropertyList(PropertyListDatasetCreate)
	plist.SetChunk([]uint64{10})
	dset, err := f.CreateDataset("resizeable", DatatypeInt32, Dataspace{Dims: []uint64{5}, MaxDims: []uint64{H5S_UNLIMITED}}, plist)
	if err != nil {
		t.Fatalf("Failed to create resizeable dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write([]int32{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("Failed to write initial data: %v", err)
	}

	err = dset.Resize([]uint64{10})
	if err != nil {
		t.Fatalf("Failed to resize to 10: %v", err)
	}

	sel1 := Selection{Hyperslabs: []Hyperslab{{Start: []uint64{5}, Count: []uint64{5}}}}
	err = dset.WriteSlice([]int32{6, 7, 8, 9, 10}, sel1)
	if err != nil {
		t.Fatalf("Failed to write at position 5: %v", err)
	}

	err = dset.Resize([]uint64{15})
	if err != nil {
		t.Fatalf("Failed to resize to 15: %v", err)
	}

	sel2 := Selection{Hyperslabs: []Hyperslab{{Start: []uint64{10}, Count: []uint64{5}}}}
	err = dset.WriteSlice([]int32{11, 12, 13, 14, 15}, sel2)
	if err != nil {
		t.Fatalf("Failed to write at position 10: %v", err)
	}

	var readData []int32
	err = dset.Read(&readData)
	if err != nil {
		t.Fatalf("Failed to read dataset: %v", err)
	}

	if len(readData) != 15 {
		t.Errorf("Expected 15 elements, got %d", len(readData))
	}

	for i := 0; i < 15; i++ {
		if readData[i] != int32(i+1) {
			t.Errorf("Data[%d] = %d, expected %d", i, readData[i], i+1)
			break
		}
	}
}

func TestFileReopenMultiple(t *testing.T) {
	testFile := "test_reopen_multiple.h5"
	defer os.Remove(testFile)

	for i := 0; i < 5; i++ {
		f, err := OpenFile(testFile, Create)
		if err != nil {
			t.Fatalf("Attempt %d: Failed to create file: %v", i, err)
		}

		dset, err := f.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{10}}, PropertyList{})
		if err != nil {
			t.Fatalf("Attempt %d: Failed to create dataset: %v", i, err)
		}

		data := make([]int32, 10)
		for j := range data {
			data[j] = int32(i*10 + j)
		}

		err = dset.Write(data)
		if err != nil {
			t.Fatalf("Attempt %d: Failed to write dataset: %v", i, err)
		}

		dset.Close()
		f.Close()
	}

	f, err := OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to open file for final read: %v", err)
	}
	defer f.Close()

	dset, err := f.GetDataset("data")
	if err != nil {
		t.Fatalf("Failed to get dataset: %v", err)
	}
	defer dset.Close()

	var readData []int32
	err = dset.Read(&readData)
	if err != nil {
		t.Fatalf("Failed to read dataset: %v", err)
	}

	if len(readData) != 10 {
		t.Errorf("Expected 10 elements, got %d", len(readData))
	}

	expectedBase := 4 * 10
	for j := 0; j < 10; j++ {
		if readData[j] != int32(expectedBase+j) {
			t.Errorf("Data[%d] = %d, expected %d", j, readData[j], expectedBase+j)
		}
	}
}

func TestLargeStringDataset(t *testing.T) {
	testFile := "test_large_string.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	strings := []string{
		"hello world",
		"test string",
		"hdf5 go implementation",
		"this is a longer string for testing purposes",
		"short",
	}

	stringDtype := Datatype{Class: DatatypeString, Size: 100}
	dset, err := f.CreateDataset("strings", stringDtype, Dataspace{Dims: []uint64{5}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create string dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write(strings)
	if err != nil {
		t.Fatalf("Failed to write string dataset: %v", err)
	}

	var readStrings []string
	err = dset.Read(&readStrings)
	if err != nil {
		t.Fatalf("Failed to read string dataset: %v", err)
	}

	if len(readStrings) != len(strings) {
		t.Errorf("Expected %d strings, got %d", len(strings), len(readStrings))
	}

	for i, s := range strings {
		if readStrings[i] != s {
			t.Errorf("String[%d]: expected '%s', got '%s'", i, s, readStrings[i])
		}
	}
}

func TestCompoundWithStrings(t *testing.T) {
	testFile := "test_compound_strings.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	nameDtype := Datatype{Class: DatatypeString, Size: 50}
	fields := []CompoundField{
		{Name: "ID", Offset: 0, Datatype: DatatypeInt32},
		{Name: "Name", Offset: 4, Datatype: nameDtype},
	}

	dtype := Datatype{Class: DatatypeCompound, Size: 54, Fields: fields}

	type Record struct {
		ID   int32
		Name string
	}

	records := []Record{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	dset, err := f.CreateDataset("compound_strings", dtype, Dataspace{Dims: []uint64{3}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create compound dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write(records)
	if err != nil {
		t.Fatalf("Failed to write compound dataset: %v", err)
	}

	var readRecords []Record
	err = dset.Read(&readRecords)
	if err != nil {
		t.Fatalf("Failed to read compound dataset: %v", err)
	}

	if len(readRecords) != 3 {
		t.Errorf("Expected 3 records, got %d", len(readRecords))
	}

	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for i := 0; i < 3; i++ {
		if readRecords[i].ID != int32(i+1) {
			t.Errorf("Record[%d].ID: expected %d, got %d", i, i+1, readRecords[i].ID)
		}
		if readRecords[i].Name != expectedNames[i] {
			t.Errorf("Record[%d].Name: expected '%s', got '%s'", i, expectedNames[i], readRecords[i].Name)
		}
	}
}

func TestEnumWithIntegerValues(t *testing.T) {
	testFile := "test_enum_values.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	members := []EnumMember{
		{Name: "RED", Value: int64(1)},
		{Name: "GREEN", Value: int64(2)},
		{Name: "BLUE", Value: int64(4)},
	}

	dtype := Datatype{Class: DatatypeEnum, Size: 4, EnumMembers: members, BaseType: &DatatypeInt32}

	dset, err := f.CreateDataset("colors", dtype, Dataspace{Dims: []uint64{3}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create enum dataset: %v", err)
	}
	defer dset.Close()

	values := []int32{1, 2, 4}
	err = dset.Write(values)
	if err != nil {
		t.Fatalf("Failed to write enum dataset: %v", err)
	}

	var readValues []int32
	err = dset.Read(&readValues)
	if err != nil {
		t.Fatalf("Failed to read enum dataset: %v", err)
	}

	if len(readValues) != 3 {
		t.Errorf("Expected 3 values, got %d", len(readValues))
	}

	for i, v := range values {
		if readValues[i] != v {
			t.Errorf("Value[%d]: expected %d, got %d", i, v, readValues[i])
		}
	}
}

func TestArrayDatatype2D(t *testing.T) {
	testFile := "test_array_2d.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	arrayDtype := Datatype{Class: DatatypeArray, Size: 9 * 8, ArrayDims: []uint64{3, 3}, BaseType: &DatatypeFloat64}

	dset, err := f.CreateDataset("2d_arrays", arrayDtype, Dataspace{Dims: []uint64{2}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create 2D array dataset: %v", err)
	}
	defer dset.Close()

	data := [][]float64{
		{1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3},
		{4.1, 4.2, 4.3, 5.1, 5.2, 5.3, 6.1, 6.2, 6.3},
	}

	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write 2D array dataset: %v", err)
	}

	var readData [][]float64
	err = dset.Read(&readData)
	if err != nil {
		t.Fatalf("Failed to read 2D array dataset: %v", err)
	}

	if len(readData) != 2 {
		t.Errorf("Expected 2 arrays, got %d", len(readData))
	}

	for i, arr := range data {
		if len(readData[i]) != len(arr) {
			t.Errorf("Array[%d]: expected length %d, got %d", i, len(arr), len(readData[i]))
			continue
		}
		for j, v := range arr {
			if math.Abs(readData[i][j]-v) > 1e-9 {
				t.Errorf("Array[%d][%d]: expected %f, got %f", i, j, v, readData[i][j])
			}
		}
	}
}

func TestVarLengthData(t *testing.T) {
	testFile := "test_varlength.h5"
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
	defer dset.Close()

	data := [][]int32{
		{1, 2, 3},
	}

	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write varlength dataset: %v", err)
	}

	f.Close()

	f, err = OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f.Close()

	dset2, err := f.GetDataset("varlength")
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

	var readData [][]int32
	err = dset2.Read(&readData)
	if err != nil {
		t.Fatalf("Failed to read varlength dataset: %v", err)
	}

	if len(readData) != 1 {
		t.Errorf("Expected 1 array, got %d", len(readData))
		return
	}

	if len(readData[0]) != 3 {
		t.Errorf("Expected length 3, got %d", len(readData[0]))
		return
	}

	for i, v := range []int32{1, 2, 3} {
		if readData[0][i] != v {
			t.Errorf("Expected %d, got %d", v, readData[0][i])
		}
	}
}

func TestMaskSelectionOnLargeDataset(t *testing.T) {
	testFile := "test_mask_large.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	size := 1000
	data := make([]int32, size)
	mask := make([]bool, size)
	expected := []int32{}

	for i := 0; i < size; i++ {
		data[i] = int32(i)
		if i%3 == 0 {
			mask[i] = true
			expected = append(expected, int32(i))
		} else {
			mask[i] = false
		}
	}

	dset, err := f.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{uint64(size)}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write dataset: %v", err)
	}

	sel := Selection{
		Mask: mask,
	}

	var readData []int32
	err = dset.ReadSlice(&readData, sel)
	if err != nil {
		t.Fatalf("Failed to read masked selection: %v", err)
	}

	if len(readData) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(readData))
	}

	for i, v := range expected {
		if readData[i] != v {
			t.Errorf("Data[%d]: expected %d, got %d", i, v, readData[i])
		}
	}
}

func TestPointSelectionScattered(t *testing.T) {
	testFile := "test_point_scattered.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	size := 100
	data := make([]int32, size)
	for i := 0; i < size; i++ {
		data[i] = int32(i)
	}

	dset, err := f.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{uint64(size)}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write dataset: %v", err)
	}

	points := [][]uint64{{5}, {15}, {25}, {50}, {75}, {99}}
	expected := []int32{5, 15, 25, 50, 75, 99}

	sel := Selection{
		Points: points,
	}

	var readData []int32
	err = dset.ReadSlice(&readData, sel)
	if err != nil {
		t.Fatalf("Failed to read point selection: %v", err)
	}

	if len(readData) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(readData))
	}

	for i, v := range expected {
		if readData[i] != v {
			t.Errorf("Data[%d]: expected %d, got %d", i, v, readData[i])
		}
	}
}

func TestMultiHyperslabNonContiguous(t *testing.T) {
	testFile := "test_multi_hyperslab_noncontiguous.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	size := 100
	data := make([]int32, size)
	for i := 0; i < size; i++ {
		data[i] = int32(i)
	}

	dset, err := f.CreateDataset("data", DatatypeInt32, Dataspace{Dims: []uint64{uint64(size)}}, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create dataset: %v", err)
	}
	defer dset.Close()

	err = dset.Write(data)
	if err != nil {
		t.Fatalf("Failed to write dataset: %v", err)
	}

	sel := Selection{
		Hyperslabs: []Hyperslab{
			{Start: []uint64{0}, Count: []uint64{5}, Stride: []uint64{1}},
			{Start: []uint64{50}, Count: []uint64{10}, Stride: []uint64{1}},
			{Start: []uint64{90}, Count: []uint64{10}, Stride: []uint64{1}},
		},
	}

	var readData []int32
	err = dset.ReadSlice(&readData, sel)
	if err != nil {
		t.Fatalf("Failed to read multi-hyperslab selection: %v", err)
	}

	expectedLen := 5 + 10 + 10
	if len(readData) != expectedLen {
		t.Errorf("Expected %d elements, got %d", expectedLen, len(readData))
	}

	expected := []int32{0, 1, 2, 3, 4, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99}
	for i, v := range expected {
		if readData[i] != v {
			t.Errorf("Data[%d]: expected %d, got %d", i, v, readData[i])
		}
	}
}

func TestCompressionOnSmallDataset(t *testing.T) {
	testFile := "test_compression_small.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	plist := NewPropertyList(PropertyListDatasetCreate)
	plist.SetDeflate(5)

	sizes := []uint64{1, 10, 100}
	names := []string{"tiny", "small", "medium"}

	for i, size := range sizes {
		data := make([]int32, size)
		for j := uint64(0); j < size; j++ {
			data[j] = int32(j)
		}

		dset, err := f.CreateDataset(names[i], DatatypeInt32, Dataspace{Dims: []uint64{size}}, plist)
		if err != nil {
			t.Fatalf("Failed to create compressed dataset %s (size=%d): %v", names[i], size, err)
		}

		err = dset.Write(data)
		if err != nil {
			t.Fatalf("Failed to write compressed dataset %s (size=%d): %v", names[i], size, err)
		}
		dset.Close()

		dset, err = f.GetDataset(names[i])
		if err != nil {
			t.Fatalf("Failed to get compressed dataset %s: %v", names[i], err)
		}

		var readData []int32
		err = dset.Read(&readData)
		if err != nil {
			t.Fatalf("Failed to read compressed dataset %s: %v", names[i], err)
		}

		if uint64(len(readData)) != size {
			t.Errorf("Dataset %s: expected %d elements, got %d", names[i], size, len(readData))
		}

		for j := uint64(0); j < size; j++ {
			if readData[j] != int32(j) {
				t.Errorf("Dataset %s: data[%d] = %d, expected %d", names[i], j, readData[j], j)
				break
			}
		}

		dset.Close()
	}
}

func TestEdgeCasesSummary(t *testing.T) {
	testFile := "test_edge_cases.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	err = f.Attributes().Set("global_attr", "test_value")
	if err != nil {
		t.Fatalf("Failed to set global attribute: %v", err)
	}

	var attrValue string
	attr, err := f.Attributes().Get("global_attr")
	if err != nil {
		t.Fatalf("Failed to get global attribute: %v", err)
	}
	err = attr.Read(&attrValue)
	if err != nil {
		t.Fatalf("Failed to read global attribute: %v", err)
	}
	if attrValue != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", attrValue)
	}

	rootGroup, err := f.GetGroup("/")
	if err != nil {
		t.Fatalf("Failed to get root group: %v", err)
	}

	keys := rootGroup.Keys()
	t.Logf("Root group has %d keys", len(keys))

	emptyGroup, err := f.CreateGroup("empty_group")
	if err != nil {
		t.Fatalf("Failed to create empty group: %v", err)
	}
	emptyGroup.Close()

	emptyGroup, err = f.GetGroup("empty_group")
	if err != nil {
		t.Fatalf("Failed to get empty group: %v", err)
	}

	emptyKeys := emptyGroup.Keys()
	if len(emptyKeys) != 0 {
		t.Errorf("Expected empty group to have 0 keys, got %d", len(emptyKeys))
	}
	emptyGroup.Close()

	t.Log("All edge case tests passed!")
}
