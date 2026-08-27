package main

import "time"

// Shared local types so deploy/exec keep their signatures while the
// side-effect half of internal/security (audit/approval/cleanup/signature)
// stays out of the default build. Real behavior is provided by
// sec_hooks_sec.go (-tags opssec); sec_hooks_nosec.go holds the no-op build.

// approvalSource mirrors security.ApprovalSource without linking the
// side-effect half of internal/security.
type approvalSource string

const (
	approveInteractive    approvalSource = "interactive"
	approveAutoFlag       approvalSource = "auto-approve-flag"
	approveAutoEnv        approvalSource = "auto-approve-env"
	approveNonInteractive approvalSource = "non-interactive"
)

// approvalRecord mirrors the auditable outcome of an approval decision in
// both build flavors; opssec fills every field, the no-op build leaves it
// nil or zero-valued.
type approvalRecord struct {
	Required     bool
	Decision     string // "approved" | "denied"
	Source       string
	User         string
	Privilege    string
	MutatingOps  []string
	ProdTargets  []string
	TotalTargets int
	DecidedAt    time.Time
}
