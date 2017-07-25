package kv

import (
	"sort"
)

// go1.9: type keys = sort.StringSlice

type KeyVal struct {
	sort.StringSlice
	Vals []string
}

func (kv KeyVal) Swap(i, j int) {
	kv.StringSlice[i], kv.StringSlice[j] = kv.StringSlice[j], kv.StringSlice[i]
	kv.Vals[i], kv.Vals[j] = kv.Vals[j], kv.Vals[i]
}

func (kv *KeyVal) Append(key, val string) {
	kv.StringSlice = append(kv.StringSlice, key)
	kv.Vals = append(kv.Vals, val)
}

func (kv *KeyVal) Index(key string) (int, bool) {
	i := kv.Search(key)

	if i < len(kv.StringSlice) && kv.StringSlice[i] == key {
		return i, true
	}

	return 0, false 
}
