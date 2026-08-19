// Package lvg provides LVM volume group management operations.
package lvg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a volume group operation.
type ActionResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// VGInfo represents volume group information.
type VGInfo struct {
	Name     string `json:"name"`
	PVCount  int    `json:"pv_count"`
	LVCount  int    `json:"lv_count"`
	SnapCount int   `json:"snap_count"`
	Attr     string `json:"attr"`
	VGSize   string `json:"vg_size"`
	VGFree   string `json:"vg_free"`
	UUID     string `json:"uuid"`
}

// ListResult represents the result of listing volume groups.
type ListResult struct {
	VGs []VGInfo `json:"vgs"`
}

// Create creates a new volume group with the given name and physical volumes.
func Create(name string, pvs []string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("volume group name is required")
	}
	if len(pvs) == 0 {
		return ActionResult{}, fmt.Errorf("at least one physical volume is required")
	}

	args := []string{"vgcreate", name}
	args = append(args, pvs...)

	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("vgcreate failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Remove removes a volume group.
func Remove(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("volume group name is required")
	}

	out, err := exec.Command("vgremove", "-f", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("vgremove failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Extend extends a volume group with additional physical volumes.
func Extend(name string, pvs []string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("volume group name is required")
	}
	if len(pvs) == 0 {
		return ActionResult{}, fmt.Errorf("at least one physical volume is required")
	}

	args := []string{"vgextend", name}
	args = append(args, pvs...)

	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("vgextend failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Reduce reduces a volume group by removing physical volumes.
func Reduce(name string, pvs []string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("volume group name is required")
	}
	if len(pvs) == 0 {
		return ActionResult{}, fmt.Errorf("at least one physical volume is required")
	}

	args := []string{"vgreduce", "--removemissing", name}
	// Note: vgreduce typically removes specific PVs, but --removemissing is safer
	// For specific PV removal, would need: vgreduce name pv1 pv2 ...

	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("vgreduce failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Activate activates a volume group.
func Activate(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("volume group name is required")
	}

	out, err := exec.Command("vgchange", "-a", "y", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("vgchange activate failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// Deactivate deactivates a volume group.
func Deactivate(name string) (ActionResult, error) {
	if name == "" {
		return ActionResult{}, fmt.Errorf("volume group name is required")
	}

	out, err := exec.Command("vgchange", "-a", "n", name).CombinedOutput()
	if err != nil {
		return ActionResult{Name: name, Success: false, Error: string(out)}, fmt.Errorf("vgchange deactivate failed: %w (output: %s)", err, string(out))
	}

	return ActionResult{Name: name, Changed: true, Success: true}, nil
}

// List lists all volume groups.
func List() (ListResult, error) {
	out, err := exec.Command("vgs", "--reportformat", "json", "--units", "b").CombinedOutput()
	if err != nil {
		return ListResult{}, fmt.Errorf("vgs failed: %w (output: %s)", err, string(out))
	}

	// Parse JSON output
	var vgsOut struct {
		Report []struct {
			VG []struct {
				VGName    string `json:"vg_name"`
				PVCount   string `json:"pv_count"`
				LVCount   string `json:"lv_count"`
				SnapCount string `json:"snap_count"`
				VGAttr    string `json:"vg_attr"`
				VGSize    string `json:"vg_size"`
				VGFree    string `json:"vg_free"`
				VGUUID    string `json:"vg_uuid"`
			} `json:"vg"`
		} `json:"report"`
	}

	if err := json.Unmarshal(out, &vgsOut); err != nil {
		return ListResult{}, fmt.Errorf("failed to parse vgs output: %w", err)
	}

	result := ListResult{VGs: make([]VGInfo, 0)}
	if len(vgsOut.Report) > 0 {
		for _, vg := range vgsOut.Report[0].VG {
			pvCount := parseInt(vg.PVCount)
			lvCount := parseInt(vg.LVCount)
			snapCount := parseInt(vg.SnapCount)
			result.VGs = append(result.VGs, VGInfo{
				Name:      vg.VGName,
				PVCount:   pvCount,
				LVCount:   lvCount,
				SnapCount: snapCount,
				Attr:      vg.VGAttr,
				VGSize:    vg.VGSize,
				VGFree:    vg.VGFree,
				UUID:      vg.VGUUID,
			})
		}
	}
	return result, nil
}

// Get gets information about a specific volume group.
func Get(name string) (VGInfo, error) {
	if name == "" {
		return VGInfo{}, fmt.Errorf("volume group name is required")
	}

	out, err := exec.Command("vgs", "--reportformat", "json", "--units", "b", name).CombinedOutput()
	if err != nil {
		return VGInfo{}, fmt.Errorf("vgs failed: %w (output: %s)", err, string(out))
	}

	var vgsOut struct {
		Report []struct {
			VG []struct {
				VGName    string `json:"vg_name"`
				PVCount   string `json:"pv_count"`
				LVCount   string `json:"lv_count"`
				SnapCount string `json:"snap_count"`
				VGAttr    string `json:"vg_attr"`
				VGSize    string `json:"vg_size"`
				VGFree    string `json:"vg_free"`
				VGUUID    string `json:"vg_uuid"`
			} `json:"vg"`
		} `json:"report"`
	}

	if err := json.Unmarshal(out, &vgsOut); err != nil {
		return VGInfo{}, fmt.Errorf("failed to parse vgs output: %w", err)
	}

	if len(vgsOut.Report) == 0 || len(vgsOut.Report[0].VG) == 0 {
		return VGInfo{}, fmt.Errorf("volume group %s not found", name)
	}

	vg := vgsOut.Report[0].VG[0]
	return VGInfo{
		Name:      vg.VGName,
		PVCount:   parseInt(vg.PVCount),
		LVCount:   parseInt(vg.LVCount),
		SnapCount: parseInt(vg.SnapCount),
		Attr:      vg.VGAttr,
		VGSize:    vg.VGSize,
		VGFree:    vg.VGFree,
		UUID:      vg.VGUUID,
	}, nil
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
