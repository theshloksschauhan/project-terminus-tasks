package util

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
	"strings"
)

func FacetMaterial(unit string, blend int, ord int) string {
	h := sha256.Sum256([]byte(unit + "|" + itoa(blend) + "|" + itoa(ord)))
	return hex.EncodeToString(h[:8])
}

func RunReduce(rows [][2]string, bandSpan uint16) string {
	parts := make([]string, 0, len(rows))
	for _, pair := range rows {
		parts = append(parts, pair[0]+pair[1])
	}
	sort.Strings(parts)
	payload := strings.Join(parts, "")
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, bandSpan)
	payload += string(buf)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])[:16]
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
