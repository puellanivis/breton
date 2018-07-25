package gnuflag

import (
	"testing"
)

func TestName(t *testing.T) {
	if got := flagName("", "flag"); got != "flag" {
		t.Errorf("Name(flag) %q != flag", got)
	}

	if got := flagName("", "Flag"); got != "flag" {
		t.Errorf("Name(Flag) %q != flag", got)
	}

	if got := flagName("", "FlagOne"); got != "flag-one" {
		t.Errorf("Name(FlagOne) %q != flag-one", got)
	}

	if got := flagName("", "AReallyLongFlagName"); got != "a-really-long-flag-name" {
		t.Errorf("Name(ReallyLongFlagName) %q != a-really-long-flag-name", got)
	}

	if got := flagName("", "BaseURL"); got != "base-url" {
		t.Errorf("Name(BaseURL) %q != base-url", got)
	}

	if got := flagName("", "URLAddress"); got != "url-address" {
		t.Errorf("Name(URLAddress) %q != url-address", got)
	}
}

type structTest struct{
	Alpha string `flag:",default=10" desc:"a first element with a default"`

	Beta  string `                desc:"a flag with name from field and no default"`
	Gamma string `flag:",short=g" desc:"a flag with a short version"`
	Delta string `flag:",short=Δ" desc:"a flag with a utf8 rune as short version"`
	Eta   string `flag:"echo"     desc:"a flag that is being renamed"`
}

func TestStruct(t *testing.T) {
	f := structTest{}
	fs := NewFlagSet("name", ContinueOnError)

	if err := fs.Struct("test", &f); err != nil {
		t.Fatalf("gnuflag.FlagSet.Struct returned an unexpected error: %v", err)
	}

	expectFlags := []string{
		"test-alpha",
		"test-beta",
		"test-gamma",
		"test-delta",
		"test-echo",
	}

	for _, testFor := range expectFlags {
		if _, ok := fs.formal[testFor]; !ok {
			t.Errorf("field %s did not appear", testFor)
		}
	}

	expectShorts := []rune{
		'g',
		'Δ',
	}

	for _, testFor := range expectShorts {
		if _, ok := fs.short[testFor]; !ok {
			t.Errorf("short field %c did not appear", testFor)
		}
	}

	if fs.formal["test-gamma"] != fs.short['g'] {
		t.Errorf("--test-gamma and -g do not point to the same flag")
	}

	if fs.formal["test-delta"] != fs.short['Δ'] {
		t.Errorf("--test-delta and -Δ do not point to the same flag")
	}

	expectDefaults := map[string]string{
		"test-alpha": "10",
	}

	for k, v := range expectDefaults {
		f, ok := fs.formal[k]
		if !ok {
			t.Fatalf("tried to look up fs.format[%q] which should exist, but didn‘t", k)
		}

		if f.DefValue != v {
			t.Errorf("flag --%s did not have the correct default value: expected %q, but got %q", k, v, f.DefValue)
		}
	}
}
