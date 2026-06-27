package command

// PerlScript holds the perl source passed to perl -e.
// Use either Perl(PerlScript("..."), opts...) or Perl("...", opts...).
type PerlScript string

func (s PerlScript) Configure(flags *flags) { flags.script = string(s) }

type perlLoopFlag bool

const (
	PerlLoop   perlLoopFlag = true
	PerlNoLoop perlLoopFlag = false
)

type perlPrintFlag bool

const (
	PerlPrint   perlPrintFlag = true
	PerlNoPrint perlPrintFlag = false
)

type perlAutoSplitFlag bool

const (
	PerlAutoSplit   perlAutoSplitFlag = true
	PerlNoAutoSplit perlAutoSplitFlag = false
)

type flags struct {
	script    string
	loop      perlLoopFlag
	print     perlPrintFlag
	autoSplit perlAutoSplitFlag
}

func (l perlLoopFlag) Configure(flags *flags)      { flags.loop = l }
func (p perlPrintFlag) Configure(flags *flags)     { flags.print = p }
func (a perlAutoSplitFlag) Configure(flags *flags) { flags.autoSplit = a }
