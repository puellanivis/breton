package gnuflag

import (
	"reflect"
	"testing"
	"time"
)

type mockFlag struct {
	Value string
}

func (m mockFlag) String() string {
	return m.Value
}

func (m *mockFlag) Set(s string) error {
	m.Value = s
	return nil
}

func (m mockFlag) Get() interface{} {
	return m.Value
}

var _ Value = &mockFlag{}

func TestStructVar(t *testing.T) {
	var Flags struct {
		ignored               bool
		True                  bool    `flag:"" desc:"a bool value" default:"true"`
		ZeroValueBool         bool    `flag:"" desc:"a zero value bool"`
		PiApprox              float64 `flag:",def=3.1415" desc:"a float64 with an approximation for pi"`
		NegativeFortyTwo      int     `flag:",default=-42" desc:"an int with -42"`
		SkipEvenThoughBadType uint8   `flag:"-"`
		NegativeSixtyFour     int64   `desc:"an int64 with -64" default:"-64"`
		Foo                   string  `desc:"a string with foo" default:"foo"`
		Renamed               string  `flag:"bar" desc:"a field renamed to bar"`
		FortyTwo              uint    `desc:"a uint with 42" default:"42"`
		SixtyFour             uint64  `desc:"a uint64 with 64" default:"64"`
		ANamedFlagValue       mockFlag
		AnUnnamedIntType      int

		SubFlags struct {
			Value   int
			Pointer *int
		}
	}

	var fs FlagSet

	checkedKeys := make(map[string]bool)
	if err := fs.structVar("", reflect.ValueOf(&Flags).Elem()); err != nil {
		t.Fatal("unexpected error running structVar:", err)
	}

	if len(fs.formal)+len(fs.short) == 0 {
		t.Fatal("no flags set on FlagSet")
	}

	checkFlag := func(name, value, usage string, val interface{}) {
		checkedKeys[name] = true

		f, ok := fs.formal[name]
		if !ok {
			t.Errorf("expected flag %q to exist", name)
			return
		}
		if f.DefValue != value {
			t.Errorf("flag %q has value %q, but epected %q", name, f.Value, value)
		}
		if f.Usage != usage {
			t.Errorf("flag %q has usage %q, but expected %q", name, f.Usage, usage)
		}
		ptr := reflect.ValueOf(val).Elem().Addr().Pointer()
		p := reflect.ValueOf(f.Value).Elem().Addr().Pointer()
		if ptr != p {
			t.Errorf("flag %q points to %#v, but expected %#v", name, p, ptr)
		}
	}

	checkFlag("true", "true", "a bool value", &(Flags.True))
	checkFlag("zero-value-bool", "false", "a zero value bool", &(Flags.ZeroValueBool))
	checkFlag("pi-approx", "3.1415", "a float64 with an approximation for pi", &(Flags.PiApprox))
	checkFlag("negative-forty-two", "-42", "an int with -42", &(Flags.NegativeFortyTwo))
	checkFlag("negative-sixty-four", "-64", "an int64 with -64", &(Flags.NegativeSixtyFour))
	checkFlag("foo", "foo", "a string with foo", &(Flags.Foo))
	checkFlag("bar", "", "a field renamed to bar", &(Flags.Renamed))
	checkFlag("forty-two", "42", "a uint with 42", &(Flags.FortyTwo))
	checkFlag("sixty-four", "64", "a uint64 with 64", &(Flags.SixtyFour))
	checkFlag("a-named-flag-value", "", "ANamedFlagValue `gnuflag.mockFlag`", &(Flags.ANamedFlagValue))
	checkFlag("an-unnamed-int-type", "0", "AnUnnamedIntType `int`", &(Flags.AnUnnamedIntType))
	checkFlag("sub-flags-value", "0", "Value `int`", &(Flags.SubFlags.Value))
	checkFlag("sub-flags-pointer", "0", "Pointer `*int`", Flags.SubFlags.Pointer)

	for k := range fs.formal {
		if !checkedKeys[k] {
			t.Errorf("unexpected key found: %q", k)
		}
	}
}

func TestStructVar_BadInputs(t *testing.T) {
	type test struct {
		name string
		s    interface{}
	}

	tests := []test{
		{
			name: "bad name: contains '='",
			s: &struct {
				F int `flag:"="`
			}{},
		},
		{
			name: "bad name: begins with -",
			s: &struct {
				F int `flag:"-flag"`
			}{},
		},
		{
			name: "unsupported flag type",
			s: &struct {
				F chan struct{}
			}{},
		},
		{
			name: "bad bool default",
			s: &struct {
				F bool `default:"foo"`
			}{},
		},
		{
			name: "bad time.Duration default",
			s: &struct {
				F time.Duration `default:"foo"`
			}{},
		},
		{
			name: "bad float64 default",
			s: &struct {
				F float64 `default:"foo"`
			}{},
		},
		{
			name: "bad int default",
			s: &struct {
				F int `default:"foo"`
			}{},
		},
		{
			name: "bad int64 default",
			s: &struct {
				F int64 `default:"foo"`
			}{},
		},
		{
			name: "bad uint default",
			s: &struct {
				F uint `default:"foo"`
			}{},
		},
		{
			name: "bad uint64 default",
			s: &struct {
				F uint64 `default:"foo"`
			}{},
		},
		{
			name: "uint8 default too large",
			s: &struct {
				F uint8 `default:"256"`
			}{},
		},
		{
			name: "uint16 default too large",
			s: &struct {
				F uint16 `default:"65536"`
			}{},
		},
		{
			name: "uint32 default too large",
			s: &struct {
				F uint32 `default:"4294967296"`
			}{},
		},
		{
			name: "int8 default too large",
			s: &struct {
				F int8 `default:"128"`
			}{},
		},
		{
			name: "int16 default too large",
			s: &struct {
				F int16 `default:"32768"`
			}{},
		},
		{
			name: "int32 default too large",
			s: &struct {
				F int32 `default:"2147483648"`
			}{},
		},
		{
			name: "float32 default too large",
			s: &struct {
				F int32 `default:"1e39"`
			}{},
		},
	}

	for _, tt := range tests {
		var fs FlagSet

		if err := fs.structVar("", reflect.ValueOf(tt.s).Elem()); err == nil {
			t.Error("expected error running structVar, but got none:", tt.name)
		}
	}

}

func Test_flagName(t *testing.T) {
	type test struct {
		name     string
		flagName string
		want     string
	}

	tests := []test{
		{
			name:     "create DSN variable for a country",
			flagName: "DNS",
			want:     "dns",
		},
		{
			name:     "create variable lower-casing the first character",
			flagName: "Name",
			want:     "name",
		},
		{
			name:     "one letter variable with no prefix",
			flagName: "N",
			want:     "n",
		},
		{
			name:     "split flag name depending on cases",
			flagName: "NaNo",
			want:     "na-no",
		},
		{
			name:     "split flag name, but acronym together",
			flagName: "GetMyURLFromString",
			want:     "get-my-url-from-string",
		},
		{
			name:     "split flag name, but acronym last",
			flagName: "GetMyURL",
			want:     "get-my-url",
		},
		{
			name:     "split flag name, but acronym first",
			flagName: "URLFromString",
			want:     "url-from-string",
		},
		// Legend: a = expected acronym, n = word inital capital, x = rest of non-acronym words
		{
			name:     "pathological acronym tests",
			flagName: "NxNxxANxANxxAANxAANxx",
			want:     "nx-nxx-a-nx-a-nxx-aa-nx-aa-nxx",
		},
		{
			name:     "pathological acronym tests, with initial one-letter 'acronym'",
			flagName: "ANxNxxANxANxxAANxAANxx",
			want:     "a-nx-nxx-a-nx-a-nxx-aa-nx-aa-nxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flagName(tt.flagName); got != tt.want {
				t.Errorf("flagName() = %v, want %v", got, tt.want)
			}
		})
	}
}
