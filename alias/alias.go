// Package alias provides unprefixed names for the perl command and its flags.
//
//	import perl "github.com/gloo-foo/cmd-perl/alias"
//	perl.Perl(perl.Script("s/a/b/"), perl.Print)
package alias

import command "github.com/gloo-foo/cmd-perl"

// Perl re-exports the constructor.
var Perl = command.Perl

// Script re-exports the -e script-body flag type.
type Script = command.PerlScript

// -n flag: loop over input lines
const Loop = command.PerlLoop

// default: do not loop over input lines
const NoLoop = command.PerlNoLoop

// -p flag: loop and auto-print each line
const Print = command.PerlPrint

// default: do not auto-print
const NoPrint = command.PerlNoPrint

// -a flag: autosplit each line into @F
const AutoSplit = command.PerlAutoSplit

// default: do not autosplit
const NoAutoSplit = command.PerlNoAutoSplit
