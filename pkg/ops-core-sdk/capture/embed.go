package capture

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// The "local:" prefix marks a pcap path the OPERATOR wants materialized on
// the machine running the CONTROLLER, not on the host performing the
// capture. Execution context decides the actual strategy:
//
//   - interpreter (runs where opsctl runs): treated as a plain absolute
//     path - same machine, so the file is simply written there.
//   - ops-runner / remote execution: the frame buffer is written to a
//     temporary file on the capturing host, base64-embedded into the
//     operation result under reserved keys, and the temp file removed.
//     The controller decodes it back to disk before results are shown.
//
// Diagnostic-sized captures are expected; anything beyond MaxEmbedBytes
// refuses the round-trip rather than silently blowing up the control
// channel.
const (
	// PcapLocalPrefix marks operator-local pcap delivery.
	PcapLocalPrefix = "local:"
	// KeyPCapB64 carries the base64 payload inside an op result map.
	KeyPCapB64 = "__pcap_b64"
	// KeyPCapLocalPath carries the requested controller-side path.
	KeyPCapLocalPath = "__pcap_local_path"
	// MaxEmbedBytes bounds what may travel through the control channel.
	MaxEmbedBytes = 16 << 20 // 16 MiB
)

// SplitPcapTarget classifies a requested pcap path.
func SplitPcapTarget(path string) (local bool, fsPath string) {
	if strings.HasPrefix(path, PcapLocalPrefix) {
		return true, strings.TrimPrefix(path, PcapLocalPrefix)
	}
	return false, path
}

// MaterializeLocal reads a freshly written pcap destined for the operator
// workstation, base64-encodes it, and removes the on-disk copy. tempFile is
// the staged path on the capturing host; userTarget is the workstation path
// the controller should write. Result fields PCapB64 / PCapLocalPath are
// populated; PcapPath stays empty since nothing remains on this host.
func MaterializeLocal(res *Result, tempFile, userTarget string) error {
	raw, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("read staged pcap %s: %w", tempFile, err)
	}
	if len(raw) > MaxEmbedBytes {
		os.Remove(tempFile)
		return fmt.Errorf("pcap %s (%d bytes) exceeds local-transfer limit %d bytes; capture fewer packets or lower snaplen",
			tempFile, len(raw), MaxEmbedBytes)
	}
	res.PCapB64 = base64.StdEncoding.EncodeToString(raw)
	res.PCapLocalPath = userTarget
	res.PcapPath = ""
	if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove staged pcap: %w", err)
	}
	return nil
}

// SaveEmbedded scans a result tree for capture payloads deposited by a
// remote net.capture and writes them to their requested LOCAL locations,
// stripping the reserved keys so callers never surface raw blobs in JSON.
// It returns the saved paths. Non-fatal problems are reported per key via
// the returned warnings slice.
func SaveEmbedded(node interface{}) (saved []string, warnings []string) {
	switch v := node.(type) {
	case map[string]interface{}:
		b64, hasB64 := v[KeyPCapB64]
		lp, hasLP := v[KeyPCapLocalPath]
		if hasB64 && hasLP {
			target, _ := lp.(string)
			payload, ok := b64.(string)
			if !ok || target == "" {
				warnings = append(warnings, fmt.Sprintf("malformed %s entry dropped", KeyPCapB64))
			} else {
				raw, derr := base64.StdEncoding.DecodeString(payload)
				if derr != nil {
					warnings = append(warnings, "pcap payload is not valid base64: "+derr.Error())
				} else if werr := os.WriteFile(target, raw, 0600); werr != nil {
					warnings = append(warnings, fmt.Sprintf("save %s: %v", target, werr))
				} else {
					saved = append(saved, target)
				}
			}
			delete(v, KeyPCapB64)
			delete(v, KeyPCapLocalPath)
		}
		for _, child := range v {
			s, w := SaveEmbedded(child)
			saved = append(saved, s...)
			warnings = append(warnings, w...)
		}
	case []interface{}:
		for _, child := range v {
			s, w := SaveEmbedded(child)
			saved = append(saved, s...)
			warnings = append(warnings, w...)
		}
	}
	return saved, warnings
}

// marshalStableToMap converts a typed result into the generic map shape the
// runner protocol transports.
func marshalStableToMap(r any) (map[string]interface{}, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// MarshalStable converts a typed Result into the generic map shape used by
// the runner protocol.
func MarshalStable(r Result) (map[string]interface{}, error) {
	return marshalStableToMap(&r)
}
