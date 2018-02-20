package ts

import (
	"fmt"
	"sort"
	"strings"
)

type PAT map[uint16]uint16

func (pat PAT) Unmarshal(b []byte) error {
	for i := 0; i < len(b); i += 4 {
		pnum := (uint16(b[i]) << 8) | uint16(b[i+1])
		pid := (uint16(b[i+2]&0x1F) << 8) | uint16(b[i+3])

		pat[pnum] = pid
	}

	return nil
}

func (pat PAT) String() string {
	var keys []uint16
	for key := range pat {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var out []string
	for _, key := range keys {
		out = append(out, fmt.Sprintf("x%x:x%x", key, pat[key]))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, ", "))
}
