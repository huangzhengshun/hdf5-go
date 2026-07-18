﻿package format

import (
	"encoding/binary"
	"errors"
)

const (
	FilterDeflate    = 1
	FilterShuffle    = 2
	FilterFletcher32 = 3
	FilterLZF        = 32000
	FilterScaleOffset = 32001
	FilterBitshuffle = 32008
	FilterZSTD       = 32015
)

type FilterPipelineMessage struct {
	Version         uint8
	NumberOfFilters uint8
	Filters         []FilterInfo
}

type FilterInfo struct {
	FilterID       uint32
	Flags          uint16
	NumberOfParams uint16
	Params         []uint32
	ClientData     []byte
}

func EncodeFilterPipelineMessage(filters []FilterInfo) ([]byte, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	totalSize := 2
	for _, f := range filters {
		totalSize += 8 + 4*len(f.Params) + len(f.ClientData)
	}

	buf := make([]byte, totalSize)
	buf[0] = 1
	buf[1] = uint8(len(filters))

	offset := 2
	for _, f := range filters {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], f.FilterID)
		binary.LittleEndian.PutUint16(buf[offset+4:offset+6], f.Flags)
		binary.LittleEndian.PutUint16(buf[offset+6:offset+8], f.NumberOfParams)
		offset += 8

		for _, p := range f.Params {
			binary.LittleEndian.PutUint32(buf[offset:offset+4], p)
			offset += 4
		}

		copy(buf[offset:offset+len(f.ClientData)], f.ClientData)
		offset += len(f.ClientData)
	}

	return buf, nil
}

func DecodeFilterPipelineMessage(data []byte) (FilterPipelineMessage, error) {
	if len(data) < 2 {
		return FilterPipelineMessage{}, errors.New("filter pipeline message too short")
	}

	fpm := FilterPipelineMessage{
		Version:         data[0],
		NumberOfFilters: data[1],
		Filters:         make([]FilterInfo, 0, data[1]),
	}

	offset := 2
	for i := uint8(0); i < fpm.NumberOfFilters; i++ {
		if offset+8 > len(data) {
			return fpm, errors.New("filter pipeline message truncated")
		}

		fi := FilterInfo{
			FilterID:       binary.LittleEndian.Uint32(data[offset : offset+4]),
			Flags:          binary.LittleEndian.Uint16(data[offset+4 : offset+6]),
			NumberOfParams: binary.LittleEndian.Uint16(data[offset+6 : offset+8]),
		}
		offset += 8

		if fi.NumberOfParams > 0 {
			if offset+4*int(fi.NumberOfParams) > len(data) {
				return fpm, errors.New("filter pipeline message truncated")
			}
			fi.Params = make([]uint32, fi.NumberOfParams)
			for j := uint16(0); j < fi.NumberOfParams; j++ {
				fi.Params[j] = binary.LittleEndian.Uint32(data[offset : offset+4])
				offset += 4
			}
		}

		fpm.Filters = append(fpm.Filters, fi)
	}

	return fpm, nil
}


