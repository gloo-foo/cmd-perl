package command

import (
	"context"
	"slices"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
	"github.com/spf13/afero"
)

// captured records what a fake subprocess builder was handed, so a test can
// assert exactly the command Perl would have forked — proving the contract
// without executing perl (which would be non-hermetic and depend on a working
// perl install).
type captured struct {
	name patterns.ProcessName
	args []patterns.ProcessArg
}

// fakeBuilder returns a subprocessBuilder that records its inputs into got and
// yields a pass-through Command, so no real process is ever spawned. The
// pass-through lets input-wiring branches run end to end without perl.
func fakeBuilder(got *captured) subprocessBuilder {
	return func(name patterns.ProcessName, args ...patterns.ProcessArg) gloo.Command[[]byte, []byte] {
		got.name = name
		got.args = args
		return gloo.FuncCommand[[]byte, []byte](
			func(_ context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] { return in },
		)
	}
}

// memFs returns an in-memory filesystem seeded with one file, so File-positional
// input is exercised without touching the real disk.
func memFs(t *testing.T, name, content string) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, name, []byte(content), 0o644); err != nil {
		t.Fatalf("seeding %q: %v", name, err)
	}
	return fs
}

// run executes cmd against the given input lines and collects its output lines.
func run(t *testing.T, cmd gloo.Command[[]byte, []byte], input ...string) ([]string, error) {
	t.Helper()
	src := gloo.StreamOf(toBytes(input)...)
	out := cmd.Execute(context.Background(), src)
	return collect(out)
}

// toBytes converts string lines to byte lines for stream input.
func toBytes(lines []string) [][]byte {
	out := make([][]byte, len(lines))
	for i, l := range lines {
		out[i] = []byte(l)
	}
	return out
}

// collect drains a byte stream into string lines, surfacing the first error.
func collect(s gloo.Stream[[]byte]) ([]string, error) {
	var lines []string
	for item := range s.Chan() {
		if item.Error != nil {
			return nil, item.Error
		}
		lines = append(lines, string(item.Value))
	}
	return lines, nil
}

func TestPerlWith_ScriptPositionalPromotedToFlag(t *testing.T) {
	var got captured
	perlWith(fakeBuilder(&got), afero.NewMemMapFs(), `print "hi"`)
	if got.name != "perl" {
		t.Errorf("forked %q, want \"perl\"", got.name)
	}
	if !slices.Equal(got.args, []patterns.ProcessArg{"-e", `print "hi"`}) {
		t.Errorf("got args %q, want [-e print \"hi\"]", got.args)
	}
}

func TestPerlWith_ExplicitScriptFlag(t *testing.T) {
	var got captured
	perlWith(fakeBuilder(&got), afero.NewMemMapFs(), PerlScript(`s/a/b/`))
	if !slices.Equal(got.args, []patterns.ProcessArg{"-e", `s/a/b/`}) {
		t.Errorf("got args %q, want [-e s/a/b/]", got.args)
	}
}

func TestPerlWith_AllSwitchesInCanonicalOrder(t *testing.T) {
	var got captured
	perlWith(fakeBuilder(&got), afero.NewMemMapFs(),
		PerlScript(`X`), PerlLoop, PerlPrint, PerlAutoSplit)
	if !slices.Equal(got.args, []patterns.ProcessArg{"-n", "-p", "-a", "-e", "X"}) {
		t.Errorf("got args %q, want [-n -p -a -e X]", got.args)
	}
}

func TestPerlWith_DisabledSwitchesOmitted(t *testing.T) {
	var got captured
	perlWith(fakeBuilder(&got), afero.NewMemMapFs(),
		PerlScript(`X`), PerlNoLoop, PerlNoPrint, PerlNoAutoSplit)
	if !slices.Equal(got.args, []patterns.ProcessArg{"-e", "X"}) {
		t.Errorf("got args %q, want [-e X]", got.args)
	}
}

// With an explicit PerlScript flag present, a bare string is no longer promoted
// to the script — it is classified as a positional File path instead.
func TestPerlWith_ExplicitScriptKeepsBareStringAsFile(t *testing.T) {
	var got captured
	fs := memFs(t, "in.txt", "from-file\n")
	cmd := perlWith(fakeBuilder(&got), fs, PerlScript(`pass`), "in.txt")
	if !slices.Equal(got.args, []patterns.ProcessArg{"-e", `pass`}) {
		t.Errorf("got args %q, want [-e pass]", got.args)
	}
	lines, err := run(t, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(lines, []string{"from-file"}) {
		t.Errorf("got %q, want [from-file]", lines)
	}
}

// No positional input: Perl returns the bare subprocess Command, which (via the
// fake) passes the pipeline stream straight through.
func TestPerlWith_NoInputReturnsBareSubprocess(t *testing.T) {
	var got captured
	cmd := perlWith(fakeBuilder(&got), afero.NewMemMapFs(), PerlScript(`pass`))
	lines, err := run(t, cmd, "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(lines, []string{"a", "b"}) {
		t.Errorf("got %q, want [a b]", lines)
	}
}

// A File positional opens the file and streams its lines to perl's stdin; the
// pass-through fake echoes them back, proving the wiring.
func TestPerlWith_FileInputStreamsToStdin(t *testing.T) {
	var got captured
	fs := memFs(t, "data.txt", "one\ntwo\n")
	cmd := perlWith(fakeBuilder(&got), fs, PerlScript(`pass`), "data.txt")
	lines, err := run(t, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(lines, []string{"one", "two"}) {
		t.Errorf("got %q, want [one two]", lines)
	}
}

// A missing File positional surfaces an error through the stream rather than
// forking perl with unusable input.
func TestPerlWith_MissingFileInputYieldsError(t *testing.T) {
	var got captured
	cmd := perlWith(fakeBuilder(&got), afero.NewMemMapFs(), PerlScript(`pass`), "absent.txt")
	_, err := run(t, cmd)
	if err == nil {
		t.Fatal("expected an error for a missing input file, got nil")
	}
}

// Perl wires the production builder (patterns.Subprocess) and the OS filesystem.
// Constructing the Command must not fork perl, so this asserts only that a
// usable Command is returned; the argument and wiring contracts are proven
// against the fake builder above.
func TestPerl_ReturnsCommand(t *testing.T) {
	if Perl(PerlScript(`print "x"`)) == nil {
		t.Fatal("Perl returned a nil Command")
	}
}
