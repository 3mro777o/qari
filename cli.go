package main

// Command-line parsing and Tab completion.

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	"github.com/3mro777o/qari/internal/ui"
)

// parseArgs parses flags anywhere among the arguments and returns the rest
// in order; everything after "--" is kept as is.
func parseArgs(fs *flag.FlagSet, args []string) []string {
	var pos []string
	for {
		fs.Parse(args) // ExitOnError: a bad flag exits with the usage
		rest := fs.Args()
		if len(rest) == 0 {
			return pos
		}
		if n := len(args) - len(rest); n > 0 && args[n-1] == "--" {
			return append(pos, rest...)
		}
		pos, args = append(pos, rest[0]), rest[1:]
	}
}

var commands = []ui.Cand{
	{Value: "play", Desc: "play a qari's surah"},
	{Value: "radio", Desc: "play a live radio station"},
	{Value: "random", Desc: "random play from a recipe"},
	{Value: "completion", Desc: "print the Tab-completion script"},
	{Value: "help", Desc: "show usage"},
}

// complete prints the candidates for the last of words (the word being
// typed), one "value<tab>description" per line. The shell scripts call it as
// qari __complete.
func complete(w io.Writer, words []string) {
	for _, c := range completions(words, cachedNames) {
		fmt.Fprintf(w, "%s\t%s\n", c.Value, c.Desc)
	}
}

func completions(words []string, names func(what, qari string) []ui.Cand) []ui.Cand {
	if len(words) == 0 {
		words = []string{""}
	}
	cur, prev := unquote(words[len(words)-1]), words[:len(words)-1]
	if n := len(prev); n > 0 {
		if f := valueFlag(prev[n-1]); f != nil {
			var out []ui.Cand
			if f.Name == "theme" {
				for _, t := range ui.Themes() {
					out = append(out, ui.Cand{Value: t.ID, Desc: t.Name})
				}
			}
			return filter(out, cur)
		}
	}
	if strings.HasPrefix(cur, "-") {
		var out []ui.Cand
		flag.VisitAll(func(f *flag.Flag) { out = append(out, ui.Cand{Value: "--" + f.Name, Desc: f.Usage}) })
		return filter(out, cur)
	}
	var pos []string
	for i := 0; i < len(prev); i++ {
		if a := prev[i]; strings.HasPrefix(a, "-") {
			if valueFlag(a) != nil {
				i++ // its value
			}
		} else {
			pos = append(pos, unquote(a))
		}
	}
	var out []ui.Cand
	switch strings.Join(pos, " ") {
	case "":
		out = commands
	case "completion":
		for _, s := range []string{"bash", "zsh", "fish", "powershell"} {
			out = append(out, ui.Cand{Value: s, Desc: s + " script"})
		}
	case "play":
		out = names("qari", "")
	case "radio":
		out = names("station", "")
	case "random":
		out = names("recipe", "")
	default:
		if len(pos) != 2 || pos[0] != "play" {
			return nil
		}
		out = names("surah", pos[1])
		if cur != "" && (cur[0] < '0' || cur[0] > '9') { // complete the name, not the number
			for i, c := range out {
				out[i].Value, out[i].Desc = strings.TrimSpace(c.Desc), c.Value
			}
		}
	}
	return filter(out, cur)
}

// valueFlag is the flag a word sets when its value is the next word.
func valueFlag(word string) *flag.Flag {
	if !strings.HasPrefix(word, "-") || strings.Contains(word, "=") {
		return nil
	}
	f := flag.Lookup(strings.TrimLeft(word, "-"))
	if f == nil {
		return nil
	}
	if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
		return nil
	}
	return f
}

// filter keeps the candidates that start with cur; when none do, those
// whose names contain it (so "alafasi" finds "Mishary Alafasi").
func filter(cands []ui.Cand, cur string) []ui.Cand {
	var pre, fuzzy []ui.Cand
	lc := strings.ToLower(cur)
	for _, c := range cands {
		switch {
		case strings.HasPrefix(strings.ToLower(c.Value), lc):
			pre = append(pre, c)
		case c.Search != "" && norm.Match(c.Search, cur):
			fuzzy = append(fuzzy, c)
		}
	}
	if len(pre) > 0 {
		return pre
	}
	return fuzzy
}

// unquote undoes the shell quoting a word may still carry: 'x', "x", an
// unclosed quote, or backslash escapes.
func unquote(s string) string {
	if s != "" && (s[0] == '\'' || s[0] == '"') {
		q := s[:1]
		s = strings.TrimSuffix(s[1:], q)
		if q == "'" {
			s = strings.ReplaceAll(s, "''", "'") // PowerShell
		}
		return s
	}
	return strings.ReplaceAll(s, `\`, "")
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// cachedNames reads names from the cached catalog only: completion must be
// instant and never touch the network.
func cachedNames(what, qari string) []ui.Cand {
	cacheDir, stateFile, _ := paths()
	offline := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") })}
	cat, err := api.LoadCatalog(context.Background(), api.New(offline, cacheDir), "eng")
	if err != nil {
		return nil
	}
	st, err := store.Open(stateFile)
	if err != nil {
		return nil
	}
	return ui.Names(ui.Options{Store: st, Catalog: cat, Popular: popular.Load()}, what, qari)
}

var scripts = map[string]string{
	"bash": `# bash completion for qari
_qari() {
	local IFS=$'\n'
	COMPREPLY=($("$1" __complete "${COMP_WORDS[@]:1:COMP_CWORD}" 2>/dev/null | cut -f1))
	compopt -o filenames 2>/dev/null # bash then quotes names with spaces
}
complete -o nosort -F _qari qari 2>/dev/null || complete -F _qari qari
`,
	"zsh": `#compdef qari
# zsh completion for qari
_qari() {
	local -a lines vals descs
	local line w=0
	lines=("${(@f)$("${words[1]}" __complete "${(@Q)words[2,CURRENT]}" 2>/dev/null)}")
	for line in $lines; do
		vals+=("${line%%$'\t'*}")
		(( ${#vals[-1]} > w )) && w=${#vals[-1]}
	done
	for line in $lines; do
		descs+=("${(r:w:)${line%%$'\t'*}}  -- ${line#*$'\t'}")
	done
	compadd -U -V qari -l -d descs -a vals
}
if [ "$funcstack[1]" = "_qari" ]; then _qari "$@"; else compdef _qari qari; fi
`,
	"fish": `# fish completion for qari
function __qari_complete
    set -l cur (commandline -ct)
    set -l words (commandline -opc) "$cur"
    $words[1] __complete $words[2..-1] 2>/dev/null
end
complete -c qari -f -k -a '(__qari_complete)'
`,
	"powershell": `# PowerShell completion for qari
Register-ArgumentCompleter -Native -CommandName qari, qari.exe -ScriptBlock {
	param($wordToComplete, $commandAst, $cursorPosition)
	$words = @($commandAst.CommandElements | Where-Object { $_.Extent.StartOffset -lt $cursorPosition } | Select-Object -Skip 1 | ForEach-Object { "$_" })
	if ($wordToComplete -eq '') { $words += '""' } # Windows PowerShell drops empty arguments
	qari __complete @words 2>$null | ForEach-Object {
		$v, $d = $_ -split "` + "`" + `t", 2
		$text = if ($v -match '[\s''"` + "`" + `$;,(){}@&|<>#]') { "'" + ($v -replace "'", "''") + "'" } else { $v }
		[System.Management.Automation.CompletionResult]::new($text, $v, 'ParameterValue', $(if ($d) { $d } else { $v }))
	}
}
`,
}
