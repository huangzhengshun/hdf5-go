package format

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	ObjectHeaderVersion1 = 1
	ObjectHeaderVersion2 = 2
)

const ObjectHeaderSignature = "OHDR"

type ObjectHeader struct {
	Version          uint8
	Flags            uint8
	TotalHeaderSize  uint64
	NumberOfMessages uint16
	MessageOffsets   []uint64
	Messages         []ObjectHeaderMessage
}

type ObjectHeaderMessage struct {
	Type    uint32
	Size    uint16
	Flags   uint8
	Content []byte
}

const (
	MessageNil                      = 0x00000000
	MessageDatatype                 = 0x00000001
	MessageDataspace                = 0x00000002
	MessageDataset                  = 0x00000003
	MessageFillValue                = 0x00000004
	MessageLinkInfo                 = 0x00000005
	MessageLinkMessage              = 0x00000006
	MessageGroupInfo                = 0x00000007
	MessageSymbolTable              = 0x00000008
	MessageAttribute                = 0x00000009
	MessageFilterPipeline           = 0x0000000F
	MessageObjectComment            = 0x0000000B
	MessageObjectModificationTime   = 0x0000000C
	MessageBTreeKValues             = 0x0000000D
	MessageDriverInfo               = 0x0000000E
	MessageAttributeInfo            = 0x00000010
	MessageSharedMessageTable       = 0x00000011
	MessageObjectHeaderContinuation = 0x00000012
)

func ReadObjectHeader(r io.ReadSeeker, sb *Superblock, addr uint64) (*ObjectHeader, error) {
	if _, err := r.Seek(int64(addr), io.SeekStart); err != nil {
		return nil, err
	}

	sig := make([]byte, 4)
	if _, err := r.Read(sig); err != nil {
		return nil, err
	}

	if string(sig) == ObjectHeaderSignature {
		return readObjectHeaderV2(r, sb)
	}

	if _, err := r.Seek(-4, io.SeekCurrent); err != nil {
		return nil, err
	}

	return readObjectHeaderV1(r, sb)
}

func readObjectHeaderV1(r io.ReadSeeker, sb *Superblock) (*ObjectHeader, error) {
	oh := &ObjectHeader{Version: ObjectHeaderVersion1}

	var msgType uint32
	if err := binary.Read(r, binary.LittleEndian, &msgType); err != nil {
		return nil, err
	}

	if msgType == 0 {
		return oh, nil
	}

	var msgSize uint16
	if err := binary.Read(r, binary.LittleEndian, &msgSize); err != nil {
		return nil, err
	}

	var flags uint8
	if err := binary.Read(r, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}

	if _, err := r.Seek(3, io.SeekCurrent); err != nil {
		return nil, err
	}

	content := make([]byte, msgSize-8)
	if _, err := r.Read(content); err != nil {
		return nil, err
	}

	oh.NumberOfMessages = 1
	oh.Messages = []ObjectHeaderMessage{
		{
			Type:    msgType,
			Size:    msgSize,
			Flags:   flags,
			Content: content,
		},
	}

	return oh, nil
}

func readObjectHeaderV2(r io.ReadSeeker, sb *Superblock) (*ObjectHeader, error) {
	oh := &ObjectHeader{Version: ObjectHeaderVersion2}

	if err := binary.Read(r, binary.LittleEndian, &oh.Version); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &oh.Flags); err != nil {
		return nil, err
	}

	totalSizeBytes := make([]byte, sb.LengthSize)
	if _, err := r.Read(totalSizeBytes); err != nil {
		return nil, err
	}
	oh.TotalHeaderSize = bytesToUint64(totalSizeBytes)

	if err := binary.Read(r, binary.LittleEndian, &oh.NumberOfMessages); err != nil {
		return nil, err
	}

	oh.MessageOffsets = make([]uint64, oh.NumberOfMessages)
	for i := 0; i < int(oh.NumberOfMessages); i++ {
		offsetBytes := make([]byte, sb.OffsetSize)
		if _, err := r.Read(offsetBytes); err != nil {
			return nil, err
		}
		oh.MessageOffsets[i] = bytesToUint64(offsetBytes)
	}

	currentPos, _ := r.Seek(0, io.SeekCurrent)
	sigStartPos := currentPos - (4 + 1 + 1 + int64(sb.LengthSize) + 2 + int64(sb.OffsetSize)*int64(oh.NumberOfMessages))

	oh.Messages = make([]ObjectHeaderMessage, oh.NumberOfMessages)
	for i := 0; i < int(oh.NumberOfMessages); i++ {
		if _, err := r.Seek(sigStartPos+int64(oh.MessageOffsets[i]), io.SeekStart); err != nil {
			return nil, err
		}

		var msg ObjectHeaderMessage
		if err := binary.Read(r, binary.LittleEndian, &msg.Type); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &msg.Size); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &msg.Flags); err != nil {
			return nil, err
		}

		if _, err := r.Seek(3, io.SeekCurrent); err != nil {
			return nil, err
		}

		msg.Content = make([]byte, msg.Size-10)
		if _, err := r.Read(msg.Content); err != nil {
			return nil, err
		}

		oh.Messages[i] = msg
	}

	return oh, nil
}

func WriteObjectHeader(w io.WriteSeeker, sb *Superblock, oh *ObjectHeader) (uint64, error) {
	addr, _ := w.Seek(0, io.SeekCurrent)

	if oh.Version == ObjectHeaderVersion2 {
		if _, err := w.Write([]byte(ObjectHeaderSignature)); err != nil {
			return 0, err
		}
	}

	if err := binary.Write(w, binary.LittleEndian, oh.Version); err != nil {
		return 0, err
	}

	switch oh.Version {
	case ObjectHeaderVersion1:
		if err := writeObjectHeaderV1(w, sb, oh); err != nil {
			return 0, err
		}
	case ObjectHeaderVersion2:
		if err := writeObjectHeaderV2(w, sb, oh); err != nil {
			return 0, err
		}
	default:
		return 0, errors.New("unsupported object header version")
	}

	posAfter, _ := w.Seek(0, 1)
	if uint64(posAfter) < uint64(addr)+oh.TotalHeaderSize {
		paddingSize := uint64(addr) + oh.TotalHeaderSize - uint64(posAfter)
		
		currentFileSize, _ := w.Seek(0, 2)
		if uint64(currentFileSize) < uint64(addr)+oh.TotalHeaderSize {
			padding := make([]byte, paddingSize)
			if _, err := w.Write(padding); err != nil {
				return 0, err
			}
		} else {
			w.Seek(int64(uint64(addr)+oh.TotalHeaderSize), 0)
		}
	}

	return uint64(addr), nil
}

func writeObjectHeaderV1(w io.WriteSeeker, sb *Superblock, oh *ObjectHeader) error {
	for _, msg := range oh.Messages {
		if err := binary.Write(w, binary.LittleEndian, msg.Type); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, msg.Size); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, msg.Flags); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(0)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(0)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(0)); err != nil {
			return err
		}
		if _, err := w.Write(msg.Content); err != nil {
			return err
		}
	}

	return nil
}

func writeObjectHeaderV2(w io.WriteSeeker, sb *Superblock, oh *ObjectHeader) error {
	if err := binary.Write(w, binary.LittleEndian, oh.Flags); err != nil {
		return err
	}

	totalSizeBytes := uint64ToBytes(oh.TotalHeaderSize, sb.LengthSize)
	if _, err := w.Write(totalSizeBytes); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, oh.NumberOfMessages); err != nil {
		return err
	}

	msgStart := uint64(4) + uint64(1) + uint64(1) + uint64(sb.LengthSize) + uint64(2) + uint64(sb.OffsetSize)*uint64(oh.NumberOfMessages)
	offsets := make([]uint64, len(oh.Messages))
	currentOffset := msgStart

	for i, msg := range oh.Messages {
		offsets[i] = currentOffset
		currentOffset += uint64(msg.Size)
	}

	for _, offset := range offsets {
		offsetBytes := uint64ToBytes(offset, sb.OffsetSize)
		if _, err := w.Write(offsetBytes); err != nil {
			return err
		}
	}

	for _, msg := range oh.Messages {
		if err := binary.Write(w, binary.LittleEndian, msg.Type); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, msg.Size); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, msg.Flags); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(0)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(0)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(0)); err != nil {
			return err
		}
		if _, err := w.Write(msg.Content); err != nil {
			return err
		}
	}

	return nil
}

func (oh *ObjectHeader) GetMessage(msgType uint32) ([]ObjectHeaderMessage, error) {
	var results []ObjectHeaderMessage
	for _, msg := range oh.Messages {
		if msg.Type == msgType {
			results = append(results, msg)
		}
	}
	return results, nil
}


