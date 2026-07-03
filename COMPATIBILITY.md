# Perl Command Compatibility

## Summary

cmd-perl is a thin wrapper that forks the system `perl` binary: it maps its own `-n`/`-p`/`-a` switches onto perl's, appends `-e SCRIPT`, and streams stdin through. Because the real interpreter runs the script, output is byte-identical to invoking `perl` directly with the equivalent switches (verified in the Docker integration harness against a Debian `perl` install). It exposes a deliberate subset of perl's command line — the per-line filter switches — not the full perl CLI.

## Key Behaviors

```bash
# -p: run SCRIPT for each input line, then auto-print $_ (perl -p -e)
$ printf 'alpha\nbeta\n' | perl -p '$_=uc'
ALPHA
BETA

# -p with a substitution
$ printf 'banana\n' | perl -p 's/a/A/g'
bAnAnA

# -n: loop over input without auto-printing; SCRIPT prints explicitly (perl -n -e)
$ printf 'hello\nworld\n' | perl -n 'print uc($_)'
HELLO
WORLD

# -a autosplit with -p / -n: @F holds the whitespace-split fields (perl -a)
$ printf 'one two three\n' | perl -p -a '$_=join(":",@F)."\n"'
one:two:three

# No mode switch: SCRIPT runs once; stdin is not looped over (perl -e)
$ printf 'ignored\n' | perl 'print "constant\n"'
constant
```

## Intentional Divergences

- cmd-perl exposes a subset of perl's command line: the SCRIPT (perl `-e`) plus the three per-line filter switches `-n` (`--loop`), `-p` (`--print`), and `-a` (`--autosplit`). The script is supplied as a positional operand rather than via an explicit `-e`. The many other perl options (`-i`, `-l`, `-0`, `-F`, `-M`, multiple `-e`, `@ARGV` file operands, `-w`/`-W`, etc.) are not exposed by the wrapper.
- The SCRIPT itself is full Perl, executed by the real `perl` binary — there is no reimplemented or limited expression language. For the switches that are exposed, output matches `perl` exactly.
- Input is read from stdin (or a positional file/reader at the framework level); perl's own trailing `@ARGV` file-argument handling is not surfaced through the CLI.
