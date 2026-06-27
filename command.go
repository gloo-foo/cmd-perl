package command

import (
	"context"
	"io"

	"github.com/destel/rill"
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
	"github.com/spf13/afero"
)

// subprocessBuilder constructs the Command that forks the external process. It
// is the seam that lets tests substitute a fake for patterns.Subprocess, so the
// argument-building and input-wiring logic is covered without a real perl
// binary on the host.
type subprocessBuilder func(name string, args ...string) gloo.Command[[]byte, []byte]

// Perl returns a Command that forks perl with a script.
//
// Two call shapes are supported:
//
//	Perl("script", opts...)              // script as required positional
//	Perl(PerlScript("script"), opts...)  // script as a flag (preferred in examples)
//
// Options: PerlLoop (-n), PerlPrint (-p), PerlAutoSplit (-a),
// PerlScript (script body), positional io.Reader / gloo.File for input.
func Perl(opts ...any) gloo.Command[[]byte, []byte] {
	return perlWith(patterns.Subprocess, afero.NewOsFs(), opts...)
}

// perlWith is Perl with an injectable subprocess builder and filesystem. The
// seams keep argument construction and file-input wiring testable without
// running perl or touching the real filesystem.
func perlWith(build subprocessBuilder, fs afero.Fs, opts ...any) gloo.Command[[]byte, []byte] {
	opts = promoteScript(opts)
	params := gloo.NewParameters[gloo.File, flags](opts...)
	sub := build("perl", perlArgs(params.Flags)...)
	if len(params.Positional) == 0 {
		return sub
	}
	return inputBoundCommand(sub, params, fs)
}

// promoteScript rewrites the first bare string option into a PerlScript flag,
// supporting the Perl("script", ...) call shape. An explicit PerlScript option
// already supplies the script, so no promotion happens and a bare string is
// then treated as a positional file path.
func promoteScript(opts []any) []any {
	if hasScript(opts) {
		return opts
	}
	for i, o := range opts {
		if s, ok := o.(string); ok {
			opts[i] = PerlScript(s)
			break
		}
	}
	return opts
}

// hasScript reports whether opts already contains an explicit PerlScript flag.
func hasScript(opts []any) bool {
	for _, o := range opts {
		if _, ok := o.(PerlScript); ok {
			return true
		}
	}
	return false
}

// perlArgs renders the perl command-line arguments from the parsed flags: the
// mode switches in GNU order followed by the -e script body.
func perlArgs(f flags) []string {
	args := switchArgs(f)
	return append(args, "-e", f.script)
}

// switchArgs collects the boolean mode switches that are enabled.
func switchArgs(f flags) []string {
	var args []string
	for _, s := range switches(f) {
		if s.on {
			args = append(args, s.flag)
		}
	}
	return args
}

// perlSwitch pairs a mode flag with whether it is enabled.
type perlSwitch struct {
	flag string
	on   bool
}

// switches lists the perl mode switches in canonical command-line order.
func switches(f flags) []perlSwitch {
	return []perlSwitch{
		{flag: "-n", on: bool(f.loop)},
		{flag: "-p", on: bool(f.print)},
		{flag: "-a", on: bool(f.autoSplit)},
	}
}

// inputBoundCommand wraps sub so that positional file/reader inputs are opened
// and streamed to perl's stdin when the command executes.
func inputBoundCommand(sub gloo.Command[[]byte, []byte], params gloo.Parameters[gloo.File, flags], fs afero.Fs) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		reader, err := params.ReaderFrom(fs, nil)
		if err != nil {
			return gloo.Wrap(rill.FromSlice([][]byte(nil), err))
		}
		input := gloo.ByteReaderSource([]io.Reader{reader}).Stream(ctx)
		return sub.Execute(ctx, input)
	})
}
