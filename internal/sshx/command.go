package sshx

import (
	"regexp"
	"strings"
)

var safeShellWord = regexp.MustCompile(`^[A-Za-z0-9_./:@%+=,-]+$`)

// ShellQuote quotes one argument for the command interpreter used by an SSH
// exec channel. SSH servers commonly invoke that channel through a shell, so
// paths and operator-controlled values must not be concatenated raw.
func ShellQuote(value string) string {
	if safeShellWord.MatchString(value) {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

// JoinCommand builds a safely quoted command line from argv-style arguments.
func JoinCommand(args ...string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = ShellQuote(arg)
	}
	return strings.Join(quoted, " ")
}
