﻿package format

import (
	"fmt"
	"strings"
)

func HexDump(data []byte, offset int) string {
	var sb strings.Builder
	addr := offset

	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}

		sb.WriteString(fmt.Sprintf("%08x  ", addr))

		for j := i; j < end; j++ {
			sb.WriteString(fmt.Sprintf("%02x ", data[j]))
			if j == i+7 {
				sb.WriteString(" ")
			}
		}

		if end-i < 16 {
			pad := (16 - (end - i)) * 3
			if end-i <= 8 {
				pad += 1
			}
			sb.WriteString(strings.Repeat(" ", pad))
		}

		sb.WriteString("|")
		for j := i; j < end; j++ {
			c := data[j]
			if c >= 32 && c <= 126 {
				sb.WriteByte(c)
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteString("|\n")

		addr += 16
	}

	return sb.String()
}

func HexDumpBytes(data []byte) string {
	return HexDump(data, 0)
}

