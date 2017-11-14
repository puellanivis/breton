package gnuflag

import (
	"testing"
)

func TestName(t *testing.T) {
	if got := flagName("", "flag") ; got != "flag" {
		t.Errorf("Name(flag) %q != flag", got)
	}

	if got := flagName("", "Flag") ; got != "flag" {
		t.Errorf("Name(Flag) %q != flag", got)
	}

	if got := flagName("", "FlagOne") ; got != "flag-one" {
		t.Errorf("Name(FlagOne) %q != flag-one", got)
	}


	if got := flagName("", "AReallyLongFlagName") ; got != "a-really-long-flag-name" {
		t.Errorf("Name(ReallyLongFlagName) %q != a-really-long-flag-name", got)
	}

	if got := flagName("", "BaseURL") ; got != "base-url" {
		t.Errorf("Name(BaseURL) %q != base-url", got)
	}

	if got := flagName("", "URLAddress") ; got != "url-address" {
		t.Errorf("Name(URLAddress) %q != url-address", got)
	}
}
