package gnuflag

import (
	"reflect"
	"testing"
)

func TestStructVar_StringSlices(t *testing.T) {
	var Flags struct {
		A   []string `flag:",def=a"`
		AB  []string `flag:",def=a,b"`
		ABC []string `flag:",def=a,b,c"`
		S   []string
	}

	Flags.S = []string{"x", "y"}

	var fs FlagSet

	if err := fs.structVar("", reflect.ValueOf(&Flags).Elem()); err != nil {
		t.Fatal("unexpected error running structVar:", err)
	}

	if len(fs.formal)+len(fs.short) == 0 {
		t.Fatal("no flags set on FlagSet")
	}

	checkedKeys := make(map[string]bool)
	checkFlag := func(name, value, usage string, val, expect interface{}) {
		checkedKeys[name] = true

		f, ok := fs.formal[name]
		if !ok {
			t.Errorf("expected flag %q to exist", name)
			return
		}
		if f.DefValue != value {
			t.Errorf("flag %q has default value %q, but epected %q", name, f.Value, value)
		}
		if f.Usage != usage {
			t.Errorf("flag %q has usage %q, but expected %q", name, f.Usage, usage)
		}
		if !reflect.DeepEqual(val, expect) {
			t.Errorf("flag %q is %#v, but expected %#v", name, val, expect)
		}
	}

	checkFlag("a", "a", "A `[]string`", Flags.A, []string{"a"})
	checkFlag("ab", "a,b", "AB `[]string`", Flags.AB, []string{"a", "b"})
	checkFlag("abc", "a,b,c", "ABC `[]string`", Flags.ABC, []string{"a", "b", "c"})
	checkFlag("s", "x,y", "S `[]string`", Flags.S, []string{"x", "y"})

	for k := range fs.formal {
		if !checkedKeys[k] {
			t.Errorf("unexpected key found: %q", k)
		}
	}

	if err := fs.Set("a", "d"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("ab", "d"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("abc", "d"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	checkFlag("a", "a", "A `[]string`", Flags.A, []string{"a", "d"})
	checkFlag("ab", "a,b", "AB `[]string`", Flags.AB, []string{"a", "b", "d"})
	checkFlag("abc", "a,b,c", "ABC `[]string`", Flags.ABC, []string{"a", "b", "c", "d"})
}

func TestStructVar_IntSlices(t *testing.T) {
	var Flags struct {
		A   []int `flag:",def=2"`
		AB  []int `flag:",def=2,3"`
		ABC []int `flag:",def=2,3,5"`
		S   []int
	}

	Flags.S = []int{42, 13}

	var fs FlagSet

	if err := fs.structVar("", reflect.ValueOf(&Flags).Elem()); err != nil {
		t.Fatal("unexpected error running structVar:", err)
	}

	if len(fs.formal)+len(fs.short) == 0 {
		t.Fatal("no flags set on FlagSet")
	}

	checkedKeys := make(map[string]bool)
	checkFlag := func(name, value, usage string, val, expect interface{}) {
		checkedKeys[name] = true

		f, ok := fs.formal[name]
		if !ok {
			t.Errorf("expected flag %q to exist", name)
			return
		}
		if f.DefValue != value {
			t.Errorf("flag %q has default value %q, but epected %q", name, f.Value, value)
		}
		if f.Usage != usage {
			t.Errorf("flag %q has usage %q, but expected %q", name, f.Usage, usage)
		}
		if !reflect.DeepEqual(val, expect) {
			t.Errorf("flag %q is %#v, but expected %#v", name, val, expect)
		}
	}

	checkFlag("a", "2", "A `[]int`", Flags.A, []int{2})
	checkFlag("ab", "2,3", "AB `[]int`", Flags.AB, []int{2, 3})
	checkFlag("abc", "2,3,5", "ABC `[]int`", Flags.ABC, []int{2, 3, 5})
	checkFlag("s", "42,13", "S `[]int`", Flags.S, []int{42, 13})

	for k := range fs.formal {
		if !checkedKeys[k] {
			t.Errorf("unexpected key found: %q", k)
		}
	}

	if err := fs.Set("a", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("ab", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("abc", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	checkFlag("a", "2", "A `[]int`", Flags.A, []int{2, 7})
	checkFlag("ab", "2,3", "AB `[]int`", Flags.AB, []int{2, 3, 7})
	checkFlag("abc", "2,3,5", "ABC `[]int`", Flags.ABC, []int{2, 3, 5, 7})
}

func TestStructVar_Int64Slices(t *testing.T) {
	var Flags struct {
		A   []int64 `flag:",def=2"`
		AB  []int64 `flag:",def=2,3"`
		ABC []int64 `flag:",def=2,3,5"`
		S   []int64
	}

	Flags.S = []int64{42, 13}

	var fs FlagSet

	if err := fs.structVar("", reflect.ValueOf(&Flags).Elem()); err != nil {
		t.Fatal("unexpected error running structVar:", err)
	}

	if len(fs.formal)+len(fs.short) == 0 {
		t.Fatal("no flags set on FlagSet")
	}

	checkedKeys := make(map[string]bool)
	checkFlag := func(name, value, usage string, val, expect interface{}) {
		checkedKeys[name] = true

		f, ok := fs.formal[name]
		if !ok {
			t.Errorf("expected flag %q to exist", name)
			return
		}
		if f.DefValue != value {
			t.Errorf("flag %q has default value %q, but epected %q", name, f.Value, value)
		}
		if f.Usage != usage {
			t.Errorf("flag %q has usage %q, but expected %q", name, f.Usage, usage)
		}
		if !reflect.DeepEqual(val, expect) {
			t.Errorf("flag %q is %#v, but expected %#v", name, val, expect)
		}
	}

	checkFlag("a", "2", "A `[]int64`", Flags.A, []int64{2})
	checkFlag("ab", "2,3", "AB `[]int64`", Flags.AB, []int64{2, 3})
	checkFlag("abc", "2,3,5", "ABC `[]int64`", Flags.ABC, []int64{2, 3, 5})
	checkFlag("s", "42,13", "S `[]int64`", Flags.S, []int64{42, 13})

	for k := range fs.formal {
		if !checkedKeys[k] {
			t.Errorf("unexpected key found: %q", k)
		}
	}

	if err := fs.Set("a", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("ab", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("abc", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	checkFlag("a", "2", "A `[]int64`", Flags.A, []int64{2, 7})
	checkFlag("ab", "2,3", "AB `[]int64`", Flags.AB, []int64{2, 3, 7})
	checkFlag("abc", "2,3,5", "ABC `[]int64`", Flags.ABC, []int64{2, 3, 5, 7})
}

func TestStructVar_UintSlices(t *testing.T) {
	var Flags struct {
		A   []uint `flag:",def=2"`
		AB  []uint `flag:",def=2,3"`
		ABC []uint `flag:",def=2,3,5"`
		S   []uint
	}

	Flags.S = []uint{42, 13}

	var fs FlagSet

	if err := fs.structVar("", reflect.ValueOf(&Flags).Elem()); err != nil {
		t.Fatal("unexpected error running structVar:", err)
	}

	if len(fs.formal)+len(fs.short) == 0 {
		t.Fatal("no flags set on FlagSet")
	}

	checkedKeys := make(map[string]bool)
	checkFlag := func(name, value, usage string, val, expect interface{}) {
		checkedKeys[name] = true

		f, ok := fs.formal[name]
		if !ok {
			t.Errorf("expected flag %q to exist", name)
			return
		}
		if f.DefValue != value {
			t.Errorf("flag %q has default value %q, but epected %q", name, f.Value, value)
		}
		if f.Usage != usage {
			t.Errorf("flag %q has usage %q, but expected %q", name, f.Usage, usage)
		}
		if !reflect.DeepEqual(val, expect) {
			t.Errorf("flag %q is %#v, but expected %#v", name, val, expect)
		}
	}

	checkFlag("a", "2", "A `[]uint`", Flags.A, []uint{2})
	checkFlag("ab", "2,3", "AB `[]uint`", Flags.AB, []uint{2, 3})
	checkFlag("abc", "2,3,5", "ABC `[]uint`", Flags.ABC, []uint{2, 3, 5})
	checkFlag("s", "42,13", "S `[]uint`", Flags.S, []uint{42, 13})

	for k := range fs.formal {
		if !checkedKeys[k] {
			t.Errorf("unexpected key found: %q", k)
		}
	}

	if err := fs.Set("a", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("ab", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("abc", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	checkFlag("a", "2", "A `[]uint`", Flags.A, []uint{2, 7})
	checkFlag("ab", "2,3", "AB `[]uint`", Flags.AB, []uint{2, 3, 7})
	checkFlag("abc", "2,3,5", "ABC `[]uint`", Flags.ABC, []uint{2, 3, 5, 7})
}

func TestStructVar_Uint64Slices(t *testing.T) {
	var Flags struct {
		A   []uint64 `flag:",def=2"`
		AB  []uint64 `flag:",def=2,3"`
		ABC []uint64 `flag:",def=2,3,5"`
		S   []uint64
	}

	Flags.S = []uint64{42, 13}

	var fs FlagSet

	if err := fs.structVar("", reflect.ValueOf(&Flags).Elem()); err != nil {
		t.Fatal("unexpected error running structVar:", err)
	}

	if len(fs.formal)+len(fs.short) == 0 {
		t.Fatal("no flags set on FlagSet")
	}

	checkedKeys := make(map[string]bool)
	checkFlag := func(name, value, usage string, val, expect interface{}) {
		checkedKeys[name] = true

		f, ok := fs.formal[name]
		if !ok {
			t.Errorf("expected flag %q to exist", name)
			return
		}
		if f.DefValue != value {
			t.Errorf("flag %q has default value %q, but epected %q", name, f.Value, value)
		}
		if f.Usage != usage {
			t.Errorf("flag %q has usage %q, but expected %q", name, f.Usage, usage)
		}
		if !reflect.DeepEqual(val, expect) {
			t.Errorf("flag %q is %#v, but expected %#v", name, val, expect)
		}
	}

	checkFlag("a", "2", "A `[]uint64`", Flags.A, []uint64{2})
	checkFlag("ab", "2,3", "AB `[]uint64`", Flags.AB, []uint64{2, 3})
	checkFlag("abc", "2,3,5", "ABC `[]uint64`", Flags.ABC, []uint64{2, 3, 5})
	checkFlag("s", "42,13", "S `[]uint64`", Flags.S, []uint64{42, 13})

	for k := range fs.formal {
		if !checkedKeys[k] {
			t.Errorf("unexpected key found: %q", k)
		}
	}

	if err := fs.Set("a", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("ab", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	if err := fs.Set("abc", "7"); err != nil {
		t.Fatal("unexpected error setting flag:", err)
	}

	checkFlag("a", "2", "A `[]uint64`", Flags.A, []uint64{2, 7})
	checkFlag("ab", "2,3", "AB `[]uint64`", Flags.AB, []uint64{2, 3, 7})
	checkFlag("abc", "2,3,5", "ABC `[]uint64`", Flags.ABC, []uint64{2, 3, 5, 7})
}
