package runner

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

func testKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	return pub, priv
}

func testPackage() *InstructionPackage {
	return &InstructionPackage{
		Version:   "1.0",
		TaskID:    "sign-test",
		DryRun:    false,
		Privilege: string(ast.PrivilegeReadOnly),
		Instructions: []Instruction{
			{Op: "sys.hostname", Args: map[string]interface{}{}, Assign: "h"},
			{Op: "log", Args: map[string]interface{}{"message": "$h"}},
		},
	}
}

func TestSignAndVerifyRoundTrip(t *testing.T) {
	pub, priv := testKeyPair(t)
	pkg := testPackage()

	if err := SignPackage(pkg, priv); err != nil {
		t.Fatalf("SignPackage: %v", err)
	}
	if pkg.Signature == "" {
		t.Fatal("signature was not set")
	}

	ok, err := VerifyPackage(pkg, pub)
	if err != nil {
		t.Fatalf("VerifyPackage: %v", err)
	}
	if !ok {
		t.Fatal("valid signature failed verification")
	}
}

func TestVerifyDetectsTampering(t *testing.T) {
	pub, priv := testKeyPair(t)
	pkg := testPackage()
	if err := SignPackage(pkg, priv); err != nil {
		t.Fatalf("SignPackage: %v", err)
	}

	// Each mutation below must break verification; the signature covers
	// the whole payload, not just the instructions.
	mutations := map[string]func(*InstructionPackage){
		"flip dry_run":     func(p *InstructionPackage) { p.DryRun = true },
		"change privilege": func(p *InstructionPackage) { p.Privilege = "root" },
		"add instruction": func(p *InstructionPackage) {
			p.Instructions = append(p.Instructions,
				Instruction{Op: "file.delete", Args: map[string]interface{}{"path": "/"}})
		},
		"change task id": func(p *InstructionPackage) { p.TaskID = "evil" },
	}
	for name, mutate := range mutations {
		pkg2 := testPackage()
		if err := SignPackage(pkg2, priv); err != nil {
			t.Fatalf("%s: SignPackage: %v", name, err)
		}
		mutate(pkg2)
		ok, verr := VerifyPackage(pkg2, pub)
		if verr != nil {
			t.Fatalf("%s: unexpected verify error: %v", name, verr)
		}
		if ok {
			t.Fatalf("%s: tampered package passed verification", name)
		}
	}
}

func TestVerifyRejectsUnsignedAndWrongKey(t *testing.T) {
	pub, _ := testKeyPair(t)
	otherPub, otherPriv := testKeyPair(t)

	unsigned := testPackage()
	if _, err := VerifyPackage(unsigned, pub); err == nil {
		t.Fatal("unsigned package must produce an error")
	} else if !strings.Contains(err.Error(), "no signature") {
		t.Fatalf("error should mention the missing signature, got: %v", err)
	}

	signed := testPackage()
	if err := SignPackage(signed, otherPriv); err != nil {
		t.Fatalf("SignPackage: %v", err)
	}
	ok, err := VerifyPackage(signed, pub)
	if err != nil {
		t.Fatalf("VerifyPackage with wrong key: %v", err)
	}
	if ok {
		t.Fatal("signature from a different key must not verify")
	}
	_ = otherPub
}

func TestSignPackageRejectsBadKeySize(t *testing.T) {
	pkg := testPackage()
	if err := SignPackage(pkg, ed25519.PrivateKey("too-short")); err == nil {
		t.Fatal("expected error for undersized private key")
	}
	pub, _ := testKeyPair(t)
	if _, err := VerifyPackage(pkg, pub[:10]); err == nil {
		t.Fatal("expected error for undersized public key")
	}
}

func TestSignatureIsDeterministicJSONPayload(t *testing.T) {
	// The signed payload must be the marshaled package minus Signature;
	// re-signing an identical package twice yields two valid signatures
	// over the same bytes (Ed25519 is deterministic).
	pub, priv := testKeyPair(t)
	pkg1 := testPackage()
	pkg2 := testPackage()
	if err := SignPackage(pkg1, priv); err != nil {
		t.Fatal(err)
	}
	if err := SignPackage(pkg2, priv); err != nil {
		t.Fatal(err)
	}
	b1, _ := json.Marshal(pkg1)
	b2, _ := json.Marshal(pkg2)
	if !bytes.Equal(b1, b2) {
		t.Fatal("identical packages must serialize identically after signing")
	}
	if ok, err := VerifyPackage(pkg2, pub); err != nil || !ok {
		t.Fatalf("second package must verify: (%v, %v)", ok, err)
	}
}

func TestVerifyDoesNotMutatePackage(t *testing.T) {
	_, priv := testKeyPair(t)
	pkg := testPackage()
	if err := SignPackage(pkg, priv); err != nil {
		t.Fatal(err)
	}
	sigBefore := pkg.Signature
	pub, _ := testKeyPair(t)
	// Wrong key on purpose: verification must fail without clearing or
	// otherwise touching pkg.Signature.
	_, _ = VerifyPackage(pkg, pub)
	if pkg.Signature != sigBefore {
		t.Fatal("VerifyPackage must not modify the package signature field")
	}
}
