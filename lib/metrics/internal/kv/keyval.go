package kv

import (
	"sort"
)

// A KeyVal defines an equivalent store to map[string]string that can be used for small maps where traversing in sorted order is important.
// (Small Go maps enforce random traversal which can impose a lot of overhead.)
type KeyVal struct {
	Keys []string
	Vals []string
}

// Len returns how many entries are in the KeyVal map.
func (kv KeyVal) Len() int {
	return len(kv.Keys)
}

// Less returns true iff the i-th element should sort before the j-th element.
func (kv KeyVal) Less(i, j int) bool {
	return kv.Keys[i] < kv.Keys[j]
}

// Swap swaps the i-th and j-th keys and values.
func (kv KeyVal) Swap(i, j int) {
	kv.Keys[i], kv.Keys[j] = kv.Keys[j], kv.Keys[i]
	kv.Vals[i], kv.Vals[j] = kv.Vals[j], kv.Vals[i]
}

// Sort sorts the KeyVal set.
func (kv KeyVal) Sort() {
	sort.Sort(kv)
}

// Append adds a (key:value) pair to the end of the KeyVal.
func (kv *KeyVal) Append(key, val string) {
	kv.Keys = append(kv.Keys, key)
	kv.Vals = append(kv.Vals, val)
}

// Index returns the index a given key has into the KeyVal store, and true if the key is present.
// Otherwise, it returns (0, false).
// (Caller MUST ensure that the KeyVal has been sorted. One may use kv.Sort() from the sort.Keys)
func (kv *KeyVal) Index(key string) (int, bool) {
	i := sort.Search(len(kv.Keys), func(i int) bool {
		return kv.Keys[i] >= key
	})

	if i < len(kv.Keys) && kv.Keys[i] == key {
		return i, true
	}

	return 0, false
}
