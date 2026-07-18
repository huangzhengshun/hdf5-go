package format

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	SuperblockSignature = "\x89HDF\r\n\x1a\n"
	SuperblockVersion0  = 0
	SuperblockVersion2  = 2
)

type Superblock struct {
	Version              uint8
	FreeSpaceManagerAddr uint64
	RootGroupAddr        uint64
	FileConsistencyFlags uint32
	OffsetSize           uint8
	LengthSize           uint8
	GroupLeafNodeK       uint16
	GroupInternalNodeK   uint16
}

func ReadSuperblock(r io.ReadSeeker) (*Superblock, error) {
	sig := make([]byte, 8)
	if _, err := r.Read(sig); err != nil {
		return nil, err
	}
	if string(sig) != SuperblockSignature {
		return nil, errors.New("invalid HDF5 file signature")
	}

	var version uint8
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, err
	}

	switch version {
	case SuperblockVersion0:
		return readSuperblockV0(r)
	case SuperblockVersion2:
		return readSuperblockV2(r)
	default:
		return nil, errors.New("unsupported superblock version")
	}
}

func readSuperblockV0(r io.ReadSeeker) (*Superblock, error) {
	sb := &Superblock{Version: SuperblockVersion0}

	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &sb.FreeSpaceManagerAddr); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.RootGroupAddr); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.FileConsistencyFlags); err != nil {
		return nil, err
	}

	sb.OffsetSize = 8
	sb.LengthSize = 8
	sb.GroupLeafNodeK = 16
	sb.GroupInternalNodeK = 16

	return sb, nil
}

func readSuperblockV2(r io.ReadSeeker) (*Superblock, error) {
	sb := &Superblock{Version: SuperblockVersion2}

	if err := binary.Read(r, binary.LittleEndian, &sb.OffsetSize); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.LengthSize); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.GroupLeafNodeK); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.GroupInternalNodeK); err != nil {
		return nil, err
	}

	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return nil, err
	}

	rootAddr := make([]byte, sb.OffsetSize)
	if _, err := r.Read(rootAddr); err != nil {
		return nil, err
	}
	sb.RootGroupAddr = bytesToUint64(rootAddr)

	freeAddr := make([]byte, sb.OffsetSize)
	if _, err := r.Read(freeAddr); err != nil {
		return nil, err
	}
	sb.FreeSpaceManagerAddr = bytesToUint64(freeAddr)

	if err := binary.Read(r, binary.LittleEndian, &sb.FileConsistencyFlags); err != nil {
		return nil, err
	}

	return sb, nil
}

func WriteSuperblock(w io.WriteSeeker, sb *Superblock) error {
	if _, err := w.Write([]byte(SuperblockSignature)); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, sb.Version); err != nil {
		return err
	}

	switch sb.Version {
	case SuperblockVersion0:
		return writeSuperblockV0(w, sb)
	case SuperblockVersion2:
		return writeSuperblockV2(w, sb)
	default:
		return errors.New("unsupported superblock version")
	}
}

func writeSuperblockV0(w io.WriteSeeker, sb *Superblock) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sb.FreeSpaceManagerAddr); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sb.RootGroupAddr); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sb.FileConsistencyFlags); err != nil {
		return err
	}
	return nil
}

func writeSuperblockV2(w io.WriteSeeker, sb *Superblock) error {
	if err := binary.Write(w, binary.LittleEndian, sb.OffsetSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sb.LengthSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sb.GroupLeafNodeK); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sb.GroupInternalNodeK); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}

	rootAddr := uint64ToBytes(sb.RootGroupAddr, sb.OffsetSize)
	if _, err := w.Write(rootAddr); err != nil {
		return err
	}

	freeAddr := uint64ToBytes(sb.FreeSpaceManagerAddr, sb.OffsetSize)
	if _, err := w.Write(freeAddr); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, sb.FileConsistencyFlags); err != nil {
		return err
	}

	return nil
}

func bytesToUint64(b []byte) uint64 {
	var val uint64
	for i := len(b) - 1; i >= 0; i-- {
		val = (val << 8) | uint64(b[i])
	}
	return val
}

func uint64ToBytes(val uint64, size uint8) []byte {
	b := make([]byte, size)
	for i := uint8(0); i < size; i++ {
		b[i] = byte(val >> (8 * i))
	}
	return b
}
