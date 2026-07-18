﻿package format

import (
	"encoding/binary"
)

type SymbolTableMessage struct {
	BTreeAddress     uint64
	NameHeapAddress  uint64
}

func EncodeSymbolTableMessage(btreeAddr, heapAddr uint64) []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:8], btreeAddr)
	binary.LittleEndian.PutUint64(buf[8:16], heapAddr)
	return buf
}

func DecodeSymbolTableMessage(data []byte) SymbolTableMessage {
	st := SymbolTableMessage{}
	st.BTreeAddress = binary.LittleEndian.Uint64(data[0:8])
	st.NameHeapAddress = binary.LittleEndian.Uint64(data[8:16])
	return st
}


