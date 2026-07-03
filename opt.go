package command

// PerlScript holds the perl source passed to perl -e.
// Use either Perl(PerlScript("..."), opts...) or Perl("...", opts...).
type PerlScript string

// perlLoopFlag toggles perl's -n switch (loop over input lines without printing).
type perlLoopFlag bool

const (
	PerlLoop   perlLoopFlag = true
	PerlNoLoop perlLoopFlag = false
)

// perlPrintFlag toggles perl's -p switch (loop over input lines and auto-print each).
type perlPrintFlag bool

const (
	PerlPrint   perlPrintFlag = true
	PerlNoPrint perlPrintFlag = false
)

// perlAutoSplitFlag toggles perl's -a switch (autosplit each line into @F).
type perlAutoSplitFlag bool

const (
	PerlAutoSplit   perlAutoSplitFlag = true
	PerlNoAutoSplit perlAutoSplitFlag = false
)

// flags holds the parsed perl invocation options.
type flags struct {
	script           string
	loopEnabled      perlLoopFlag
	printEnabled     perlPrintFlag
	autoSplitEnabled perlAutoSplitFlag
}

// with folds one option value into the flags, returning the updated copy and
// whether the argument was one of this command's option types (false leaves it
// for positional classification).
func (f flags) with(o any) (flags, bool) {
	switch v := o.(type) {
	case PerlScript:
		f.script = string(v)
	case perlLoopFlag:
		f.loopEnabled = v
	case perlPrintFlag:
		f.printEnabled = v
	case perlAutoSplitFlag:
		f.autoSplitEnabled = v
	default:
		return f, false
	}
	return f, true
}

// foldOptions folds the command's own option values into a flags value and
// returns the remaining arguments (positional inputs) in their original order.
func foldOptions(opts []any) (flags, []any) {
	var f flags
	var rest []any
	for _, o := range opts {
		next, isOption := f.with(o)
		if !isOption {
			rest = append(rest, o)
			continue
		}
		f = next
	}
	return f, rest
}
