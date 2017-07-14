// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*	Package flag implements command-line flag parsing compatible with GNU longopts.

	Usage:

	Define flags using flag.String(), Bool(), Int(), etc.

	This declares an integer flag, --flagname, with short flagname -f stored in the pointer ip, with type *int.
		import "flag"
		var ip = flag.Int("flagname", 1234, "help message for flagname", flag.WithShort('f'))
	If you like, you can bind the flag to a variable using the Var() functions.
		var flagvar int
		func init() {
			flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname", flag.WithShort('f'))
		}
	Or you can create custom flags that satisfy the Value interface (with
	pointer receivers) and couple them to flag parsing by
		flag.Var(&flagVal, "name", "help message for flagname")
	For such flags, the default value is just the initial value of the variable.

	After all flags are defined, call
		flag.Parse()
	to parse the command line into the defined flags.

	Flags may then be used directly. If you're using the flags themselves,
	they are all pointers; if you bind to variables, they're values.
		fmt.Println("ip has value ", *ip)
		fmt.Println("flagvar has value ", flagvar)

	After parsing, the arguments following the flags are available as the
	slice flag.Args() or individually as flag.Arg(i).
	The arguments are indexed from 0 through flag.NArg()-1.

	Command line flag syntax:
		-f    --flag
		-f=x  --flag=x
		-f x  --flag x  // non-boolean flags only
	One minus sign signifies a short flag, while two indicates a long name.
	The last form is not permitted for boolean flags because the
	meaning of the commands
		cmd --flag *
		cmd -f *
	will change if there is a file called 0, false, etc.  You must
	use the --flag=false or -f=false form to turn off a boolean flag.

	Flag parsing stops just before the first non-flag argument
	("-" is a non-flag argument) or just after the terminator "--".

	Integer flags accept 1234, 0664, 0x1234 and may be negative.
	Boolean flags may be:
		1, 0, t, f, T, F, true, false, TRUE, FALSE, True, False
	Duration flags accept any input valid for time.ParseDuration.

	The default set of command-line flags is controlled by
	top-level functions.  The FlagSet type allows one to define
	independent sets of flags, such as to implement subcommands
	in a command-line interface. The methods of FlagSet are
	analogous to the top-level functions for the command-line
	flag set.
*/
package flag

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
)

// ErrHelp is the error returned if the --help or -? flag is invoked
// but no such flag is defined.
var ErrHelp = errors.New("flag: help requested")

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes --name equivalent to --name=true
// rather than using the next command-line argument.
//
// Set is called once, in command line order, for each flag present.
// The flag package may call the String method with a zero-valued receiver,
// such as a nil pointer.
type Value interface {
	String() string

	Set(string) error
	Get() interface{}
}

// ErrorHandling defines how FlagSet.Parse behaves if the parse fails.
type ErrorHandling int

// These constants cause FlagSet.Parse to behave as described if the parse fails.
const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2).
	PanicOnError                         // Call panic with a descriptive error.
)

// A FlagSet represents a set of defined flags. The zero value of a FlagSet
// has no name and has ContinueOnError error handling.
type FlagSet struct {
	// Usage is the function called when an error occurs while parsing flags.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler.
	Usage func()

	name   string
	parsed bool

	actual map[string]*Flag
	formal map[string]*Flag
	short  map[rune]*Flag

	args          []string // arguments after flags
	errorHandling ErrorHandling
	output        io.Writer // nil means stderr; use out() accessor
}

// A Flag represents the state of a flag.
type Flag struct {
	Name     string // name as it appears on command line
	Short    rune   // short flag as it appears on command line
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message
}

// sortFlags returns the flags as a slice in lexicographical sorted order.
func sortFlags(flags map[string]*Flag) []*Flag {
	list := make(sort.StringSlice, len(flags))
	i := 0
	for _, f := range flags {
		list[i] = f.Name
		i++
	}
	list.Sort()
	result := make([]*Flag, len(list))
	for i, name := range list {
		result[i] = flags[name]
	}
	return result
}

func (f *FlagSet) out() io.Writer {
	if f.output == nil {
		return os.Stderr
	}
	return f.output
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (f *FlagSet) SetOutput(output io.Writer) {
	f.output = output
}

// VisitAll visits the flags in lexicographical order, calling fn for each.
// It visits all flags, even those not set.
func (f *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range sortFlags(f.formal) {
		fn(flag)
	}
}

// VisitAll visits the command-line flags in lexicographical order, calling
// fn for each. It visits all flags, even those not set.
func VisitAll(fn func(*Flag)) {
	CommandLine.VisitAll(fn)
}

// Visit visits the flags in lexicographical order, calling fn for each.
// It visits only those flags that have been set.
func (f *FlagSet) Visit(fn func(*Flag)) {
	for _, flag := range sortFlags(f.actual) {
		fn(flag)
	}
}

// Visit visits the command-line flags in lexicographical order, calling fn
// for each. It visits only those flags that have been set.
func Visit(fn func(*Flag)) {
	CommandLine.Visit(fn)
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func (f *FlagSet) Lookup(name string) *Flag {
	// len(string) gives the length in bytes, we want the length in runes.
	r := []rune(name)
	switch len(r) {
	case 0:
		return nil
	case 1:
		return f.short[r[0]]
	}
	return f.formal[name]
}

// Lookup returns the Flag structure of the named command-line flag,
// returning nil if none exists.
func Lookup(name string) *Flag {
	return CommandLine.Lookup(name)
}

// Set sets the value of the named flag.
func (f *FlagSet) Set(name, value string) error {
	flag, ok := f.formal[name]
	if !ok {
		return fmt.Errorf("no such flag -%v", name)
	}
	if err := flag.Value.Set(value); err != nil {
		return err
	}
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return nil
}

// Set sets the value of the named command-line flag.
func Set(name, value string) error {
	return CommandLine.Set(name, value)
}

// isZeroValue guesses whether the string represents the zero
// value for a flag. It is not accurate but in practice works OK.
func isZeroValue(flag *Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(flag.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	if value == z.Interface().(Value).String() {
		return true
	}

	switch value {
	case "false":
		return true
	case "":
		return true
	case "0":
		return true
	}
	return false
}

type valueTyper interface {
	ValueType() string
}

// UnquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func UnquoteUsage(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}

	// No explicit name, so use type if we can find one.
	name = "value"
	switch f := flag.Value.(type) {
	case boolFlag:
		name = ""
	case valueTyper:
		name = f.ValueType()
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}

// PrintDefaults prints to standard error the default values of all
// defined command-line flags in the set. See the documentation for
// the global function PrintDefaults for more information.
func (f *FlagSet) PrintDefaults() {
	f.VisitAll(func(flag *Flag) {
		// This function is rarely ever called, as such, we can afford to
		// keep just += a string, and keep things simple.
		s := " --" + flag.Name
		if flag.Short != 0 {
			s += " | -" + string(flag.Short)
		}

		name, usage := UnquoteUsage(flag)
		if len(name) > 0 {
			s += " <" + name + ">"
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if len(s) <= 7 { // space, space, '-', 'x'.
			s += "\t"
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			s += "\n  \t"
		}

		s += usage

		if !isZeroValue(flag, flag.DefValue) {
			if _, ok := flag.Value.(*stringValue); ok {
				// put quotes on the value
				s += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				s += fmt.Sprintf(" (default %v)", flag.DefValue)
			}
		}
		fmt.Fprint(f.out(), s, "\n")
	})
}

// PrintDefaults prints, to standard error unless configured otherwise,
// a usage message showing the default settings of all defined
// command-line flags.
// For an integer valued flag x, the default output has the form
//	-x int
//		usage-message-for-x (default 7)
// The usage message will appear on a separate line for anything but
// a bool flag with a one-byte name. For bool flags, the type is
// omitted and if the flag name is one byte the usage message appears
// on the same line. The parenthetical default is omitted if the
// default is the zero value for the type. The listed type, here int,
// can be changed by placing a back-quoted name in the flag's usage
// string; the first such item in the message is taken to be a parameter
// name to show in the message and the back quotes are stripped from
// the message when displayed. For instance, given
//	flag.String("I", "", "search `directory` for include files")
// the output will be
//	-I directory
//		search directory for include files.
func PrintDefaults() {
	CommandLine.PrintDefaults()
}

// defaultUsage is the default function to print a usage message.
func (f *FlagSet) defaultUsage() {
	if f.name == "" {
		fmt.Fprintf(f.out(), "Usage:\n")
	} else {
		fmt.Fprintf(f.out(), "Usage of %s:\n", f.name)
	}
	f.PrintDefaults()
}

// NOTE: Usage is not just defaultUsage(CommandLine)
// because it serves (via godoc flag Usage) as the example
// for how to write your own usage function.

// Usage prints to standard error a usage message documenting all defined command-line flags.
// It is called when an error occurs while parsing flags.
// The function is a variable that may be changed to point to a custom function.
// By default it prints a simple header and calls PrintDefaults; for details about the
// format of the output and how to control it, see the documentation for PrintDefaults.
var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	PrintDefaults()
}

// NFlag returns the number of flags that have been set.
func (f *FlagSet) NFlag() int { return len(f.actual) }

// NFlag returns the number of command-line flags that have been set.
func NFlag() int { return len(CommandLine.actual) }

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// Arg returns the i'th command-line argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func Arg(i int) string {
	return CommandLine.Arg(i)
}

// NArg is the number of arguments remaining after flags have been processed.
func (f *FlagSet) NArg() int { return len(f.args) }

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int { return len(CommandLine.args) }

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string { return f.args }

// Args returns the non-flag command-line arguments.
func Args() []string { return CommandLine.args }

func (f *FlagSet) set(flag *Flag, name string) {
	if len(name) < 1 {
		return
	}

	if f.formal == nil {
		f.formal = make(map[string]*Flag)
	}
	_, alreadythere := f.formal[name]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = fmt.Sprintf("longflag redefined: %q", name)
		} else {
			msg = fmt.Sprintf("%s longflag redefined: %q", f.name, name)
		}
		fmt.Fprintln(f.out(), msg)
		panic(msg) // Happens only if flags are declared with identical names
	}
	f.formal[name] = flag
}

func (f *FlagSet) setShort(flag *Flag, name rune) {
	if name < 1 {
		return
	}

	if f.short == nil {
		f.short = make(map[rune]*Flag)
	}
	_, alreadythere := f.short[name]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = fmt.Sprintf("shortflag redefined: %q", name)
		} else {
			msg = fmt.Sprintf("%s shortflag redefined: %q", f.name, name)
		}
		fmt.Fprintln(f.out(), msg)
		panic(msg) // Happens only if flags are declared with identical names
	}

	f.short[name] = flag
}

// Copy defines a flag as a copy of an existing flag. Using the already given name, short, usage, and default. Will obviously panic if you attempt to copy it into a FlagSet where it is already defined.
//	fs := flag.NewFlagSet("example", flag.ExitOnError)
//	fs.Copy(flag.Lookup("output"))
func (f *FlagSet) Copy(flag *Flag) {
	f.set(flag, flag.Name)
	f.setShort(flag, flag.Short)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (f *FlagSet) Var(value Value, name string, usage string, options ...Option) {
	// Remember the default value as a string; it won't change.
	flag := &Flag{
		Name:     name,
		Usage:    usage,
		Value:    value,
		DefValue: value.String(),
	}

	for _, opt := range options {
		// during initialization we discard all reversing functionality
		_ = opt(flag)
	}

	f.set(flag, name)
	f.setShort(flag, flag.Short)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func Var(value Value, name string, usage string, options ...Option) {
	CommandLine.Var(value, name, usage, options...)
}

// failf prints to standard error a formatted error and usage message and
// returns the error.
func (f *FlagSet) failf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Fprintln(f.out(), err)
	f.usage()
	return err
}

// usage calls the Usage method for the flag set if one is specified,
// or the appropriate default usage function otherwise.
func (f *FlagSet) usage() {
	if f.Usage == nil {
		f.defaultUsage()
	} else {
		f.Usage()
	}
}

func (f *FlagSet) undefinedFlag(name string) (bool, error) {
	if name == "help" || name == "?" {
		// special case for nice help message.
		f.usage()
		return false, ErrHelp
	}

	return false, f.failf("flag provided but not defined: -%s", name)
}

// parseOne parses one flag. It reports whether a flag was seen.
func (f *FlagSet) parseOne() (bool, error) {
	if len(f.args) < 1 {
		return false, nil
	}

	r := []rune(f.args[0])

	if len(r) < 2 {
		return false, nil
	}

	var long bool

	switch r[0] {
	case '–', '—':
		long = true
	case '-':
		// short form flag
	default:
		return false, nil
	}

	numMinuses := 1

	if r[1] == '-' {
		long = true
		numMinuses++
	}

	// if we're all minuses, terminate the flags
	if len(r) <= numMinuses {
		f.args = f.args[1:]
		return false, nil
	}

	switch r[numMinuses] {
	case '-', '–', '—', '=':
		return false, f.failf("bad flag syntax: %s", string(r))
	}

	name := string(r[numMinuses:])

	// it's a flag. does it have an argument?
	f.args = f.args[1:]

	var hasValue bool
	var value string

	if i := strings.IndexByte(name, '='); i > 0 {
		value = name[i+1:]
		hasValue = true
		name = name[0:i]
	}

	if long {
		flag, alreadythere := f.formal[name] // BUG
		if !alreadythere {
			return f.undefinedFlag(name)
		}
		return f.doFlag(flag, name, value, hasValue)
	}

	r = []rune(name)

	last := len(r) - 1
	if last > 0 {
		for _, c := range r[:last] {
			flag, ok := f.short[c] // BUG
			if !ok {
				return f.undefinedFlag(string(c))
			}

			// only the last short arg can have a value
			_, err := f.doFlag(flag, string(c), "", false)
			if err != nil {
				return false, err
			}
		}
	}

	c := r[last]

	flag, ok := f.short[c] // BUG
	if !ok {
		return f.undefinedFlag(string(c))
	}
	return f.doFlag(flag, string(c), value, hasValue)
}

func (f *FlagSet) doFlag(flag *Flag, name string, value string, hasValue bool) (bool, error) {
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag

	if fv, ok := flag.Value.(boolFlag); ok && fv.IsBoolFlag() {
		// special case: doesn't need an arg
		if !hasValue {
			value = "true"
		}

		if err := fv.Set(value); err != nil {
			return false, f.failf("invalid boolean value %q for -%s: %v", value, name, err)
		}

		return true, nil
	}

	// It must have a value, which might be the next argument.
	if !hasValue {
		if len(f.args) < 1 {
			return false, f.failf("flag needs an argument: -%s", name)
		}

		// value is the next arg
		value, f.args = f.args[0], f.args[1:]
	}

	if err := flag.Value.Set(value); err != nil {
		return false, f.failf("invalid value %q for flag -%s: %v", value, name, err)
	}

	return true, nil
}

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = arguments

	if len(f.args) == 0 {
		return nil
	}

	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	// Ignore errors; CommandLine is set for ExitOnError.
	CommandLine.Parse(os.Args[1:])
}

// Parsed reports whether the command-line flags have been parsed.
func Parsed() bool {
	return CommandLine.Parsed()
}

// CommandLine is the default set of command-line flags, parsed from os.Args.
// The top-level functions such as BoolVar, Arg, and so on are wrappers for the
// methods of CommandLine.
var CommandLine = NewFlagSet(os.Args[0], ExitOnError)

func init() {
	// Override generic FlagSet default Usage with call to global Usage.
	// Note: This is not CommandLine.Usage = Usage,
	// because we want any eventual call to use any updated value of Usage,
	// not the value it has when this line is run.
	CommandLine.Usage = commandLineUsage
}

func commandLineUsage() {
	Usage()
}

// NewFlagSet returns a new, empty flag set with the specified name and
// error handling property.
func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	f := &FlagSet{
		name:          name,
		errorHandling: errorHandling,
	}
	f.Usage = f.defaultUsage
	return f
}

// Init sets the name and error handling property for a flag set.
// By default, the zero FlagSet uses an empty name and the
// ContinueOnError error handling policy.
func (f *FlagSet) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}
