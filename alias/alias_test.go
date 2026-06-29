package alias_test

import (
	"reflect"
	"testing"

	command "github.com/gloo-foo/cmd-perl"
	perl "github.com/gloo-foo/cmd-perl/alias"
)

// The alias package re-exports the Perl constructor and its flag constants
// under unprefixed names. A mis-wired re-export (Loop bound to the disabled
// PerlNoLoop, or Perl bound to the wrong function) compiles cleanly, so the
// wiring must be proven. Executing the returned Command would fork real perl —
// non-hermetic and dependent on a working install — so instead each re-export
// is proven to point at the exact same value as the command package: identical
// identity means identical forking behavior.

func TestAlias_PerlReExportsConstructor(t *testing.T) {
	got := reflect.ValueOf(perl.Perl).Pointer()
	want := reflect.ValueOf(command.Perl).Pointer()
	if got != want {
		t.Fatalf("alias.Perl is not wired to command.Perl (%v != %v)", got, want)
	}
}

// Script must alias the underlying flag type, so perl.Script("x") is assignable
// to a command.PerlScript and configures the same field. A wrong type alias
// would fail to compile; a value round-trip proves the identity at runtime.
func TestAlias_ScriptAliasesFlagType(t *testing.T) {
	// The explicit command.PerlScript element type forces a compile-time
	// assignability check: perl.Script must alias command.PerlScript, else this
	// literal fails to build.
	scripts := []command.PerlScript{perl.Script("s/a/b/")}
	if scripts[0] != command.PerlScript("s/a/b/") {
		t.Fatalf("alias.Script did not alias command.PerlScript, got %q", string(scripts[0]))
	}
}

func TestAlias_SwitchConstantsWireToEnabledForms(t *testing.T) {
	cases := []struct {
		alias any
		cmd   any
		name  string
	}{
		{name: "Loop", alias: perl.Loop, cmd: command.PerlLoop},
		{name: "Print", alias: perl.Print, cmd: command.PerlPrint},
		{name: "AutoSplit", alias: perl.AutoSplit, cmd: command.PerlAutoSplit},
	}
	for _, c := range cases {
		if c.alias != c.cmd {
			t.Errorf("alias.%s is not wired to its enabled command constant", c.name)
		}
	}
}

func TestAlias_DisabledConstantsWireToDisabledForms(t *testing.T) {
	cases := []struct {
		alias any
		cmd   any
		name  string
	}{
		{name: "NoLoop", alias: perl.NoLoop, cmd: command.PerlNoLoop},
		{name: "NoPrint", alias: perl.NoPrint, cmd: command.PerlNoPrint},
		{name: "NoAutoSplit", alias: perl.NoAutoSplit, cmd: command.PerlNoAutoSplit},
	}
	for _, c := range cases {
		if c.alias != c.cmd {
			t.Errorf("alias.%s is not wired to its disabled command constant", c.name)
		}
	}
}

// The re-exported constructor must still build a usable Command from the
// re-exported flags, including the bare (no-flag) case.
func TestAlias_PerlBuildsCommand(t *testing.T) {
	if perl.Perl(perl.Script(`print "x"`), perl.Loop, perl.Print) == nil {
		t.Fatal("alias.Perl(...) returned a nil Command")
	}
	if perl.Perl(perl.Script(`print "x"`)) == nil {
		t.Fatal("alias.Perl(script) returned a nil Command")
	}
}
