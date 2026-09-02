//go:build !opssec

package main

// Security hooks are implemented in sec_hooks_sec.go for every build. This
// compatibility file intentionally contains no declarations so -tags opssec
// does not create duplicate symbols.
