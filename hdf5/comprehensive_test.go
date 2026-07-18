package hdf5

import (
	"os"
	"testing"
)

func TestSimpleMultiDataset(t *testing.T) {
	testFile := "test_simple_multi.h5"
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

	intData := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	intDtype := DatatypeInt32
	intSpace := Dataspace{Dims: []uint64{10}}
	intDset, err := g.CreateDataset("int_data", intDtype, intSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create integer dataset: %v", err)
	}
	err = intDset.Write(intData)
	if err != nil {
		t.Fatalf("Failed to write integer dataset: %v", err)
	}
	intDset.Close()

	floatData := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
	floatDtype := DatatypeFloat64
	floatSpace := Dataspace{Dims: []uint64{5}}
	floatDset, err := g.CreateDataset("float_data", floatDtype, floatSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create float dataset: %v", err)
	}
	err = floatDset.Write(floatData)
	if err != nil {
		t.Fatalf("Failed to write float dataset: %v", err)
	}
	floatDset.Close()

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	f2, err := OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f2.Close()

	g2, err := f2.GetGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}
	defer g2.Close()

	intDset2, err := g2.GetDataset("int_data")
	if err != nil {
		t.Fatalf("Failed to get integer dataset: %v", err)
	}
	var readIntData []int32
	err = intDset2.Read(&readIntData)
	if err != nil {
		t.Fatalf("Failed to read integer dataset: %v", err)
	}
	for i := range readIntData {
		if readIntData[i] != intData[i] {
			t.Errorf("Integer data[%d]: expected %d, got %d", i, intData[i], readIntData[i])
		}
	}
	intDset2.Close()

	floatDset2, err := g2.GetDataset("float_data")
	if err != nil {
		t.Fatalf("Failed to get float dataset: %v", err)
	}
	var readFloatData []float64
	err = floatDset2.Read(&readFloatData)
	if err != nil {
		t.Fatalf("Failed to read float dataset: %v", err)
	}
	for i := range readFloatData {
		if readFloatData[i] != floatData[i] {
			t.Errorf("Float data[%d]: expected %f, got %f", i, floatData[i], readFloatData[i])
		}
	}
	floatDset2.Close()

	t.Log("Simple multi-dataset test passed!")
}

func TestComprehensiveHDF5Operations(t *testing.T) {
	testFile := "test_comprehensive.h5"
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

	subGroup, err := g.CreateGroup("sub_group")
	if err != nil {
		t.Fatalf("Failed to create sub group: %v", err)
	}
	defer subGroup.Close()

	err = g.Attributes().Set("group_attr_string", "hello world")
	if err != nil {
		t.Fatalf("Failed to set group attribute string: %v", err)
	}

	err = g.Attributes().Set("group_attr_int", int32(42))
	if err != nil {
		t.Fatalf("Failed to set group attribute int: %v", err)
	}

	err = g.Attributes().Set("group_attr_float", float64(3.14))
	if err != nil {
		t.Fatalf("Failed to set group attribute float: %v", err)
	}

	t.Log("Testing integer datasets...")
	intData := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	intDtype := DatatypeInt32
	intSpace := Dataspace{Dims: []uint64{10}}
	intDset, err := g.CreateDataset("int_data", intDtype, intSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create integer dataset: %v", err)
	}
	err = intDset.Write(intData)
	if err != nil {
		t.Fatalf("Failed to write integer dataset: %v", err)
	}
	err = intDset.Attributes().Set("data_attr", "integer data")
	if err != nil {
		t.Fatalf("Failed to set integer dataset attribute: %v", err)
	}
	intDset.Close()

	t.Log("Testing float datasets...")
	floatData := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
	floatDtype := DatatypeFloat64
	floatSpace := Dataspace{Dims: []uint64{5}}
	floatDset, err := g.CreateDataset("float_data", floatDtype, floatSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create float dataset: %v", err)
	}
	err = floatDset.Write(floatData)
	if err != nil {
		t.Fatalf("Failed to write float dataset: %v", err)
	}
	floatDset.Close()

	t.Log("Testing 2D datasets...")
	twoDData := []float32{1, 2, 3, 4, 5, 6}
	twoDDtype := DatatypeFloat32
	twoDSpace := Dataspace{Dims: []uint64{2, 3}}
	twoDDset, err := g.CreateDataset("2d_data", twoDDtype, twoDSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create 2D dataset: %v", err)
	}
	err = twoDDset.Write(twoDData)
	if err != nil {
		t.Fatalf("Failed to write 2D dataset: %v", err)
	}
	twoDDset.Close()

	t.Log("Testing compound datasets...")
	compoundDtype := Datatype{
		Class: DatatypeCompound,
		Fields: []CompoundField{
			{Name: "ID", Offset: 0, Datatype: DatatypeInt32},
			{Name: "Value", Offset: 4, Datatype: DatatypeFloat64},
		},
	}
	compoundSpace := Dataspace{Dims: []uint64{3}}
	compoundDset, err := g.CreateDataset("compound_data", compoundDtype, compoundSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create compound dataset: %v", err)
	}
	compoundData := []struct {
		ID    int32
		Value float64
	}{
		{1, 10.5},
		{2, 20.5},
		{3, 30.5},
	}
	err = compoundDset.Write(compoundData)
	if err != nil {
		t.Fatalf("Failed to write compound dataset: %v", err)
	}
	compoundDset.Close()

	t.Log("Testing string datasets...")
	stringData := []string{"hello", "world", "hdf5"}
	stringDtype := Datatype{
		Class:         DatatypeString,
		Size:          32,
		StringPadding: NullTerminated,
		StringSize:    32,
	}
	stringSpace := Dataspace{Dims: []uint64{3}}
	stringDset, err := g.CreateDataset("string_data", stringDtype, stringSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create string dataset: %v", err)
	}
	err = stringDset.Write(stringData)
	if err != nil {
		t.Fatalf("Failed to write string dataset: %v", err)
	}
	stringDset.Close()

	t.Log("Testing chunked datasets...")
	chunkedData := make([]int32, 100)
	for i := range chunkedData {
		chunkedData[i] = int32(i)
	}
	chunkedDtype := DatatypeInt32
	chunkedSpace := Dataspace{Dims: []uint64{100}}
	chunkedPlist := PropertyList{
		Chunks: []uint64{10},
	}
	chunkedDset, err := g.CreateDataset("chunked_data", chunkedDtype, chunkedSpace, chunkedPlist)
	if err != nil {
		t.Fatalf("Failed to create chunked dataset: %v", err)
	}
	err = chunkedDset.Write(chunkedData)
	if err != nil {
		t.Fatalf("Failed to write chunked dataset: %v", err)
	}
	chunkedDset.Close()

	t.Log("Testing compressed datasets (GZIP)...")
	gzipData := make([]float64, 50)
	for i := range gzipData {
		gzipData[i] = float64(i) * 0.1
	}
	gzipDtype := DatatypeFloat64
	gzipSpace := Dataspace{Dims: []uint64{50}}
	gzipPlist := PropertyList{
		Chunks:      []uint64{10},
		Compression: "gzip",
	}
	gzipDset, err := g.CreateDataset("gzip_data", gzipDtype, gzipSpace, gzipPlist)
	if err != nil {
		t.Fatalf("Failed to create gzip dataset: %v", err)
	}
	err = gzipDset.Write(gzipData)
	if err != nil {
		t.Fatalf("Failed to write gzip dataset: %v", err)
	}
	gzipDset.Close()

	t.Log("Testing compressed datasets (LZF)...")
	lzfData := make([]int64, 50)
	for i := range lzfData {
		lzfData[i] = int64(i) * 100
	}
	lzfDtype := DatatypeInt64
	lzfSpace := Dataspace{Dims: []uint64{50}}
	lzfPlist := PropertyList{
		Chunks:      []uint64{10},
		Compression: "lzf",
	}
	lzfDset, err := g.CreateDataset("lzf_data", lzfDtype, lzfSpace, lzfPlist)
	if err != nil {
		t.Fatalf("Failed to create lzf dataset: %v", err)
	}
	err = lzfDset.Write(lzfData)
	if err != nil {
		t.Fatalf("Failed to write lzf dataset: %v", err)
	}
	lzfDset.Close()

	t.Log("Testing compressed datasets (ZSTD)...")
	zstdData := make([]uint32, 50)
	for i := range zstdData {
		zstdData[i] = uint32(i) * 256
	}
	zstdDtype := DatatypeUint32
	zstdSpace := Dataspace{Dims: []uint64{50}}
	zstdPlist := PropertyList{
		Chunks:      []uint64{10},
		Compression: "zstd",
	}
	zstdDset, err := g.CreateDataset("zstd_data", zstdDtype, zstdSpace, zstdPlist)
	if err != nil {
		t.Fatalf("Failed to create zstd dataset: %v", err)
	}
	err = zstdDset.Write(zstdData)
	if err != nil {
		t.Fatalf("Failed to write zstd dataset: %v", err)
	}
	zstdDset.Close()

	t.Log("Testing soft links...")
	err = g.CreateLink("/test_group/int_data", "link_to_int")
	if err != nil {
		t.Fatalf("Failed to create soft link: %v", err)
	}

	t.Log("Testing dataset resize...")
	resizeDtype := DatatypeInt32
	resizeSpace := Dataspace{Dims: []uint64{5}, MaxDims: []uint64{10}}
	resizePlist := PropertyList{
		Chunks: []uint64{5},
	}
	resizeDset, err := g.CreateDataset("resize_data", resizeDtype, resizeSpace, resizePlist)
	if err != nil {
		t.Fatalf("Failed to create resize dataset: %v", err)
	}
	resizeData := []int32{1, 2, 3, 4, 5}
	err = resizeDset.Write(resizeData)
	if err != nil {
		t.Fatalf("Failed to write resize dataset: %v", err)
	}
	err = resizeDset.Resize([]uint64{8})
	if err != nil {
		t.Fatalf("Failed to resize dataset: %v", err)
	}
	resizeFullData := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	err = resizeDset.Write(resizeFullData)
	if err != nil {
		t.Fatalf("Failed to write full data: %v", err)
	}
	resizeDset.Close()

	t.Log("Testing point selection...")
	pointData := []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	pointDtype := DatatypeInt32
	pointSpace := Dataspace{Dims: []uint64{10}}
	pointDset, err := g.CreateDataset("point_data", pointDtype, pointSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create point dataset: %v", err)
	}
	err = pointDset.Write(pointData)
	if err != nil {
		t.Fatalf("Failed to write point dataset: %v", err)
	}
	pointSel := Selection{
		Points: [][]uint64{{1}, {3}, {5}, {7}, {9}},
	}
	var pointResult []int32
	err = pointDset.ReadSlice(&pointResult, pointSel)
	if err != nil {
		t.Fatalf("Failed to read point selection: %v", err)
	}
	expectedPoint := []int32{1, 3, 5, 7, 9}
	for i := range pointResult {
		if pointResult[i] != expectedPoint[i] {
			t.Errorf("Point selection: expected %d, got %d", expectedPoint[i], pointResult[i])
		}
	}
	pointDset.Close()

	t.Log("Testing hyperslab selection...")
	slabData := make([]float32, 20)
	for i := range slabData {
		slabData[i] = float32(i)
	}
	slabDtype := DatatypeFloat32
	slabSpace := Dataspace{Dims: []uint64{5, 4}}
	slabDset, err := g.CreateDataset("slab_data", slabDtype, slabSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create slab dataset: %v", err)
	}
	err = slabDset.Write(slabData)
	if err != nil {
		t.Fatalf("Failed to write slab dataset: %v", err)
	}
	slabSel := Selection{
		Hyperslabs: []Hyperslab{
			{Start: []uint64{1, 1}, Count: []uint64{2, 2}},
		},
	}
	var slabResult []float32
	err = slabDset.ReadSlice(&slabResult, slabSel)
	if err != nil {
		t.Fatalf("Failed to read hyperslab selection: %v", err)
	}
	expectedSlab := []float32{5, 6, 9, 10}
	for i := range slabResult {
		if slabResult[i] != expectedSlab[i] {
			t.Errorf("Hyperslab selection: expected %f, got %f", expectedSlab[i], slabResult[i])
		}
	}
	slabDset.Close()

	t.Log("Testing nested group access...")
	nestedData := []int16{100, 200, 300}
	nestedDtype := DatatypeInt16
	nestedSpace := Dataspace{Dims: []uint64{3}}
	nestedDset, err := subGroup.CreateDataset("nested_data", nestedDtype, nestedSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create nested dataset: %v", err)
	}
	err = nestedDset.Write(nestedData)
	if err != nil {
		t.Fatalf("Failed to write nested dataset: %v", err)
	}
	nestedDset.Close()

	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	t.Log("Reopening file to verify data...")
	f2, err := OpenFile(testFile, ReadOnly)
	if err != nil {
		t.Fatalf("Failed to reopen file: %v", err)
	}
	defer f2.Close()

	g2, err := f2.GetGroup("test_group")
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}
	defer g2.Close()

	t.Log("Verifying group attributes...")
	groupAttrStringAttr, err := g2.Attributes().Get("group_attr_string")
	if err != nil {
		t.Fatalf("Failed to get group attribute string: %v", err)
	}
	var groupAttrString string
	err = groupAttrStringAttr.Read(&groupAttrString)
	if err != nil {
		t.Fatalf("Failed to read group attribute string: %v", err)
	}
	if groupAttrString != "hello world" {
		t.Errorf("Group attribute string: expected 'hello world', got '%s'", groupAttrString)
	}

	groupAttrIntAttr, err := g2.Attributes().Get("group_attr_int")
	if err != nil {
		t.Fatalf("Failed to get group attribute int: %v", err)
	}
	var groupAttrInt int32
	err = groupAttrIntAttr.Read(&groupAttrInt)
	if err != nil {
		t.Fatalf("Failed to read group attribute int: %v", err)
	}
	if groupAttrInt != 42 {
		t.Errorf("Group attribute int: expected 42, got %d", groupAttrInt)
	}

	groupAttrFloatAttr, err := g2.Attributes().Get("group_attr_float")
	if err != nil {
		t.Fatalf("Failed to get group attribute float: %v", err)
	}
	var groupAttrFloat float64
	err = groupAttrFloatAttr.Read(&groupAttrFloat)
	if err != nil {
		t.Fatalf("Failed to read group attribute float: %v", err)
	}
	if groupAttrFloat != 3.14 {
		t.Errorf("Group attribute float: expected 3.14, got %f", groupAttrFloat)
	}

	t.Log("Verifying integer dataset...")
	intDset2, err := g2.GetDataset("int_data")
	if err != nil {
		t.Fatalf("Failed to get integer dataset: %v", err)
	}
	var readIntData []int32
	err = intDset2.Read(&readIntData)
	if err != nil {
		t.Fatalf("Failed to read integer dataset: %v", err)
	}
	for i := range readIntData {
		if readIntData[i] != intData[i] {
			t.Errorf("Integer data: expected %d, got %d", intData[i], readIntData[i])
		}
	}
	intDset2.Close()

	t.Log("Verifying float dataset...")
	floatDset2, err := g2.GetDataset("float_data")
	if err != nil {
		t.Fatalf("Failed to get float dataset: %v", err)
	}
	var readFloatData []float64
	err = floatDset2.Read(&readFloatData)
	if err != nil {
		t.Fatalf("Failed to read float dataset: %v", err)
	}
	for i := range readFloatData {
		if readFloatData[i] != floatData[i] {
			t.Errorf("Float data: expected %f, got %f", floatData[i], readFloatData[i])
		}
	}
	floatDset2.Close()

	t.Log("Verifying 2D dataset...")
	twoDDset2, err := g2.GetDataset("2d_data")
	if err != nil {
		t.Fatalf("Failed to get 2D dataset: %v", err)
	}
	var readTwoDData []float32
	err = twoDDset2.Read(&readTwoDData)
	if err != nil {
		t.Fatalf("Failed to read 2D dataset: %v", err)
	}
	for i := range readTwoDData {
		if readTwoDData[i] != twoDData[i] {
			t.Errorf("2D data: expected %f, got %f", twoDData[i], readTwoDData[i])
		}
	}
	twoDDset2.Close()

	t.Log("Verifying compound dataset...")
	compoundDset2, err := g2.GetDataset("compound_data")
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
	for i := range readCompoundData {
		if readCompoundData[i].ID != compoundData[i].ID || readCompoundData[i].Value != compoundData[i].Value {
			t.Errorf("Compound data: expected (%d, %f), got (%d, %f)",
				compoundData[i].ID, compoundData[i].Value,
				readCompoundData[i].ID, readCompoundData[i].Value)
		}
	}
	compoundDset2.Close()

	t.Log("Verifying string dataset...")
	stringDset2, err := g2.GetDataset("string_data")
	if err != nil {
		t.Fatalf("Failed to get string dataset: %v", err)
	}
	var readStringData []string
	err = stringDset2.Read(&readStringData)
	if err != nil {
		t.Fatalf("Failed to read string dataset: %v", err)
	}
	if len(readStringData) != len(stringData) {
		t.Errorf("String data length: expected %d, got %d", len(stringData), len(readStringData))
	} else {
		for i := range readStringData {
			expectedStr := stringData[i]
			gotStr := readStringData[i]
			if expectedStr != gotStr {
				t.Errorf("String data: expected '%s' (len=%d), got '%s' (len=%d)", expectedStr, len(expectedStr), gotStr, len(gotStr))
			}
		}
	}
	stringDset2.Close()

	t.Log("Verifying chunked dataset...")
	chunkedDset2, err := g2.GetDataset("chunked_data")
	if err != nil {
		t.Fatalf("Failed to get chunked dataset: %v", err)
	}
	var readChunkedData []int32
	err = chunkedDset2.Read(&readChunkedData)
	if err != nil {
		t.Fatalf("Failed to read chunked dataset: %v", err)
	}
	for i := range readChunkedData {
		if readChunkedData[i] != chunkedData[i] {
			t.Errorf("Chunked data: expected %d, got %d", chunkedData[i], readChunkedData[i])
		}
	}
	chunkedDset2.Close()

	t.Log("Verifying GZIP compressed dataset...")
	gzipDset2, err := g2.GetDataset("gzip_data")
	if err != nil {
		t.Fatalf("Failed to get gzip dataset: %v", err)
	}
	var readGzipData []float64
	err = gzipDset2.Read(&readGzipData)
	if err != nil {
		t.Fatalf("Failed to read gzip dataset: %v", err)
	}
	for i := range readGzipData {
		if readGzipData[i] != gzipData[i] {
			t.Errorf("GZIP data: expected %f, got %f", gzipData[i], readGzipData[i])
		}
	}
	gzipDset2.Close()

	t.Log("Verifying LZF compressed dataset...")
	lzfDset2, err := g2.GetDataset("lzf_data")
	if err != nil {
		t.Fatalf("Failed to get lzf dataset: %v", err)
	}
	var readLzfData []int64
	err = lzfDset2.Read(&readLzfData)
	if err != nil {
		t.Fatalf("Failed to read lzf dataset: %v", err)
	}
	for i := range readLzfData {
		if readLzfData[i] != lzfData[i] {
			t.Errorf("LZF data: expected %d, got %d", lzfData[i], readLzfData[i])
		}
	}
	lzfDset2.Close()

	t.Log("Verifying ZSTD compressed dataset...")
	zstdDset2, err := g2.GetDataset("zstd_data")
	if err != nil {
		t.Fatalf("Failed to get zstd dataset: %v", err)
	}
	var readZstdData []uint32
	err = zstdDset2.Read(&readZstdData)
	if err != nil {
		t.Fatalf("Failed to read zstd dataset: %v", err)
	}
	for i := range readZstdData {
		if readZstdData[i] != zstdData[i] {
			t.Errorf("ZSTD data: expected %d, got %d", zstdData[i], readZstdData[i])
		}
	}
	zstdDset2.Close()

	t.Log("Verifying soft link...")
	linkDset, err := g2.GetDataset("link_to_int")
	if err != nil {
		t.Fatalf("Failed to get dataset through soft link: %v", err)
	}
	var linkData []int32
	err = linkDset.Read(&linkData)
	if err != nil {
		t.Fatalf("Failed to read dataset through soft link: %v", err)
	}
	for i := range linkData {
		if linkData[i] != intData[i] {
			t.Errorf("Soft link data: expected %d, got %d", intData[i], linkData[i])
		}
	}
	linkDset.Close()

	t.Log("Verifying resized dataset...")
	resizeDset2, err := g2.GetDataset("resize_data")
	if err != nil {
		t.Fatalf("Failed to get resize dataset: %v", err)
	}
	var readResizeData []int32
	err = resizeDset2.Read(&readResizeData)
	if err != nil {
		t.Fatalf("Failed to read resize dataset: %v", err)
	}
	expectedResize := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	if len(readResizeData) != len(expectedResize) {
		t.Errorf("Resize data length: expected %d, got %d", len(expectedResize), len(readResizeData))
	} else {
		for i := range expectedResize {
			if readResizeData[i] != expectedResize[i] {
				t.Errorf("Resize data: expected %d, got %d", expectedResize[i], readResizeData[i])
			}
		}
	}
	resizeDset2.Close()

	t.Log("Verifying nested group dataset...")
	subGroup2, err := g2.GetGroup("sub_group")
	if err != nil {
		t.Fatalf("Failed to get sub group: %v", err)
	}
	nestedDset2, err := subGroup2.GetDataset("nested_data")
	if err != nil {
		t.Fatalf("Failed to get nested dataset: %v", err)
	}
	var readNestedData []int16
	err = nestedDset2.Read(&readNestedData)
	if err != nil {
		t.Fatalf("Failed to read nested dataset: %v", err)
	}
	for i := range readNestedData {
		if readNestedData[i] != nestedData[i] {
			t.Errorf("Nested data: expected %d, got %d", nestedData[i], readNestedData[i])
		}
	}
	nestedDset2.Close()
	subGroup2.Close()

	t.Log("Testing group keys...")
	keys := g2.Keys()
	if len(keys) == 0 {
		t.Error("Expected group keys, got empty")
	}

	t.Log("Testing visit items...")
	visitCount := 0
	err = g2.VisitItems(func(name string, obj interface{}) error {
		visitCount++
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to visit items: %v", err)
	}
	t.Logf("Visited %d items in group", visitCount)

	t.Log("Comprehensive HDF5 operations test passed!")
}

func TestHDF5InteroperabilityWithH5PY(t *testing.T) {
	testFile := "test_interop_h5py.h5"
	defer os.Remove(testFile)

	f, err := OpenFile(testFile, Create)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	g, err := f.CreateGroup("go_data")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	defer g.Close()

	intArray := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	err = g.Attributes().Set("array_attr", intArray)
	if err != nil {
		t.Fatalf("Failed to set array attribute: %v", err)
	}

	err = g.Attributes().Set("string_attr", "created by go-hdf5")
	if err != nil {
		t.Fatalf("Failed to set string attribute: %v", err)
	}

	err = g.Attributes().Set("float_attr", float64(3.14159))
	if err != nil {
		t.Fatalf("Failed to set float attribute: %v", err)
	}

	err = g.Attributes().Set("int_attr", int32(42))
	if err != nil {
		t.Fatalf("Failed to set int attribute: %v", err)
	}

	floatData := []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.1}
	floatDtype := DatatypeFloat64
	floatSpace := Dataspace{Dims: []uint64{10}}
	floatDset, err := g.CreateDataset("float_array", floatDtype, floatSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create float dataset: %v", err)
	}
	err = floatDset.Write(floatData)
	if err != nil {
		t.Fatalf("Failed to write float dataset: %v", err)
	}
	floatDset.Close()

	intData := []int64{100, 200, 300, 400, 500}
	intDtype := DatatypeInt64
	intSpace := Dataspace{Dims: []uint64{5}}
	intDset, err := g.CreateDataset("int_array", intDtype, intSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create int dataset: %v", err)
	}
	err = intDset.Write(intData)
	if err != nil {
		t.Fatalf("Failed to write int dataset: %v", err)
	}
	intDset.Close()

	twoDIntData := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}
	twoDIntDtype := DatatypeInt32
	twoDIntSpace := Dataspace{Dims: []uint64{3, 3}}
	twoDIntDset, err := g.CreateDataset("2d_int_array", twoDIntDtype, twoDIntSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create 2D int dataset: %v", err)
	}
	err = twoDIntDset.Write(twoDIntData)
	if err != nil {
		t.Fatalf("Failed to write 2D int dataset: %v", err)
	}
	twoDIntDset.Close()

	chunkedFloatData := make([]float32, 100)
	for i := range chunkedFloatData {
		chunkedFloatData[i] = float32(i) * 0.1
	}
	chunkedFloatDtype := DatatypeFloat32
	chunkedFloatSpace := Dataspace{Dims: []uint64{100}}
	chunkedFloatPlist := PropertyList{
		Chunks:      []uint64{20},
		Compression: "gzip",
	}
	chunkedFloatDset, err := g.CreateDataset("chunked_gzip_float", chunkedFloatDtype, chunkedFloatSpace, chunkedFloatPlist)
	if err != nil {
		t.Fatalf("Failed to create chunked gzip float dataset: %v", err)
	}
	err = chunkedFloatDset.Write(chunkedFloatData)
	if err != nil {
		t.Fatalf("Failed to write chunked gzip float dataset: %v", err)
	}
	chunkedFloatDset.Close()

	stringData := []string{"go", "hdf5", "library", "test"}
	stringDtype := Datatype{
		Class:         DatatypeString,
		Size:          32,
		StringPadding: NullTerminated,
		StringSize:    32,
	}
	stringSpace := Dataspace{Dims: []uint64{4}}
	stringDset, err := g.CreateDataset("string_array", stringDtype, stringSpace, PropertyList{})
	if err != nil {
		t.Fatalf("Failed to create string dataset: %v", err)
	}
	err = stringDset.Write(stringData)
	if err != nil {
		t.Fatalf("Failed to write string dataset: %v", err)
	}
	stringDset.Close()

	t.Log("HDF5 interoperability test file created. Use h5py to verify:")
	t.Log("import h5py")
	t.Log("f = h5py.File('test_interop_h5py.h5', 'r')")
	t.Log("print(f['go_data'].attrs)")
	t.Log("print(f['go_data/float_array'][:])")
	t.Log("print(f['go_data/int_array'][:])")
	t.Log("print(f['go_data/2d_int_array'][:])")
	t.Log("print(f['go_data/chunked_gzip_float'][:])")
	t.Log("print(f['go_data/string_array'][:])")
	t.Log("f.close()")
}
