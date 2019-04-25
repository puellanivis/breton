package gnuflag

// -- string Value
type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s stringValue) Get() interface{} { return string(s) }

func (s stringValue) String() string { return string(s) }

// String defines a string flag with specified name, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *FlagSet) String(name string, usage string, options ...Option) *string {
	p := new(string)
	if err := f.Var((*stringValue)(p), name, usage, options...); err != nil {
		panic(err)
	}
	return p
}

// String defines a string flag with specified name, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, usage string, options ...Option) *string {
	return CommandLine.String(name, usage, options...)
}
