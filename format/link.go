﻿package format

import (
	"encoding/binary"
	"errors"
)

const (
	LinkTypeHard     = 0
	LinkTypeSoft     = 1
	LinkTypeExternal = 64
)

type LinkMessage struct {
	Version         uint8
	LinkType        uint8
	CreationOrder   uint8
	Flags           uint8
	Name            string
	TargetAddress   uint64
	TargetPath      string
	ExternalFile    string
}

func EncodeLinkMessage(link LinkMessage) ([]byte, error) {
	nameLen := uint32(len(link.Name))
	var contentSize int

	switch link.LinkType {
	case LinkTypeHard:
		contentSize = 1 + 1 + 1 + 1 + 4 + int(nameLen) + 8
	case LinkTypeSoft:
		targetLen := uint32(len(link.TargetPath))
		contentSize = 1 + 1 + 1 + 1 + 4 + int(nameLen) + 4 + int(targetLen)
	case LinkTypeExternal:
		fileLen := uint32(len(link.ExternalFile))
		pathLen := uint32(len(link.TargetPath))
		contentSize = 1 + 1 + 1 + 1 + 4 + int(nameLen) + 4 + int(fileLen) + 4 + int(pathLen)
	default:
		return nil, errors.New("unsupported link type")
	}

	buf := make([]byte, contentSize)
	buf[0] = link.Version
	buf[1] = link.LinkType
	buf[2] = link.CreationOrder
	buf[3] = link.Flags

	binary.LittleEndian.PutUint32(buf[4:8], nameLen)
	copy(buf[8:8+nameLen], []byte(link.Name))

	offset := 8 + int(nameLen)

	switch link.LinkType {
	case LinkTypeHard:
		for i := offset; i < offset+8; i++ {
			buf[i] = 0
		}
		binary.LittleEndian.PutUint64(buf[offset:offset+8], link.TargetAddress)
	case LinkTypeSoft:
		targetLen := uint32(len(link.TargetPath))
		binary.LittleEndian.PutUint32(buf[offset:offset+4], targetLen)
		copy(buf[offset+4:offset+4+int(targetLen)], []byte(link.TargetPath))
	case LinkTypeExternal:
		fileLen := uint32(len(link.ExternalFile))
		pathLen := uint32(len(link.TargetPath))
		binary.LittleEndian.PutUint32(buf[offset:offset+4], fileLen)
		copy(buf[offset+4:offset+4+int(fileLen)], []byte(link.ExternalFile))
		binary.LittleEndian.PutUint32(buf[offset+4+int(fileLen):], pathLen)
		copy(buf[offset+8+int(fileLen):], []byte(link.TargetPath))
	}

	return buf, nil
}

func DecodeLinkMessage(data []byte) (LinkMessage, error) {
	if len(data) < 8 {
		return LinkMessage{}, errors.New("link message too short")
	}

	var link LinkMessage
	link.Version = data[0]
	link.LinkType = data[1]
	link.CreationOrder = data[2]
	link.Flags = data[3]

	nameLen := binary.LittleEndian.Uint32(data[4:8])
	if int(nameLen)+8 > len(data) {
		return link, errors.New("link message truncated")
	}
	link.Name = string(data[8 : 8+nameLen])

	offset := 8 + int(nameLen)

	switch link.LinkType {
	case LinkTypeHard:
		if offset+8 > len(data) {
			return link, errors.New("link message truncated")
		}
		link.TargetAddress = binary.LittleEndian.Uint64(data[offset : offset+8])
	case LinkTypeSoft:
		if offset+4 > len(data) {
			return link, errors.New("link message truncated")
		}
		targetLen := binary.LittleEndian.Uint32(data[offset : offset+4])
		if offset+4+int(targetLen) > len(data) {
			return link, errors.New("link message truncated")
		}
		link.TargetPath = string(data[offset+4 : offset+4+int(targetLen)])
	case LinkTypeExternal:
		if offset+4 > len(data) {
			return link, errors.New("link message truncated")
		}
		fileLen := binary.LittleEndian.Uint32(data[offset : offset+4])
		if offset+4+int(fileLen)+4 > len(data) {
			return link, errors.New("link message truncated")
		}
		link.ExternalFile = string(data[offset+4 : offset+4+int(fileLen)])

		pathOffset := offset + 4 + int(fileLen)
		pathLen := binary.LittleEndian.Uint32(data[pathOffset : pathOffset+4])
		if pathOffset+4+int(pathLen) > len(data) {
			return link, errors.New("link message truncated")
		}
		link.TargetPath = string(data[pathOffset+4 : pathOffset+4+int(pathLen)])
	}

	return link, nil
}

