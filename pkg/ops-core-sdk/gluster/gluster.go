package gluster

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// VolumeInfo represents GlusterFS volume information
type VolumeInfo struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Bricks     int    `json:"bricks"`
	Transport  string `json:"transport"`
	Snapshot   int    `json:"snapshot"`
}

// PeerInfo represents GlusterFS peer information
type PeerInfo struct {
	UUID    string `json:"uuid"`
	Host    string `json:"host"`
	State   string `json:"state"`
	Connected bool  `json:"connected"`
}

// GlusterResult represents GlusterFS operation result
type GlusterResult struct {
	Changed    bool   `json:"changed"`
	Name       string `json:"name,omitempty"`
	Message    string `json:"message,omitempty"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// VolumeList lists all volumes
func VolumeList() ([]VolumeInfo, error) {
	start := time.Now()
	cmd := exec.Command("gluster", "volume", "info")
	output, err := cmd.CombinedOutput()
	_ = time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("gluster volume info failed: %v, output: %s", err, string(output))
	}

	var volumes []VolumeInfo
	lines := strings.Split(string(output), "\n")
	var current VolumeInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Volume Name:") {
			if current.Name != "" {
				volumes = append(volumes, current)
			}
			current = VolumeInfo{Name: strings.TrimPrefix(line, "Volume Name: ")}
		} else if strings.HasPrefix(line, "Volume ID:") {
			current.ID = strings.TrimPrefix(line, "Volume ID: ")
		} else if strings.HasPrefix(line, "Status:") {
			current.Status = strings.TrimPrefix(line, "Status: ")
		} else if strings.HasPrefix(line, "Type:") {
			current.Type = strings.TrimPrefix(line, "Type: ")
		} else if strings.HasPrefix(line, "Transport-type:") {
			current.Transport = strings.TrimPrefix(line, "Transport-type: ")
		}
	}

	if current.Name != "" {
		volumes = append(volumes, current)
	}

	return volumes, nil
}

// VolumeCreate creates a new volume
func VolumeCreate(name string, bricks []string, replica int, stripe int, transport string) (GlusterResult, error) {
	start := time.Now()

	if name == "" {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("volume name required")
	}
	if len(bricks) == 0 {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("at least one brick required")
	}

	args := []string{"volume", "create", name}

	if replica > 0 {
		args = append(args, "replica", fmt.Sprintf("%d", replica))
	}
	if stripe > 0 {
		args = append(args, "stripe", fmt.Sprintf("%d", stripe))
	}

	if transport != "" {
		args = append(args, "transport", transport)
	}

	args = append(args, bricks...)
	args = append(args, "force")

	cmd := exec.Command("gluster", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := GlusterResult{
		Changed:    true,
		Name:       name,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("gluster volume create failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Volume created successfully"
	return result, nil
}

// VolumeDelete deletes a volume
func VolumeDelete(name string) (GlusterResult, error) {
	start := time.Now()

	if name == "" {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("volume name required")
	}

	// Stop volume first
	stopCmd := exec.Command("gluster", "volume", "stop", name, "force")
	_, _ = stopCmd.CombinedOutput()

	cmd := exec.Command("gluster", "volume", "delete", name)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := GlusterResult{
		Changed:    true,
		Name:       name,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("gluster volume delete failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Volume deleted successfully"
	return result, nil
}

// VolumeStart starts a volume
func VolumeStart(name string) (GlusterResult, error) {
	start := time.Now()

	if name == "" {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("volume name required")
	}

	cmd := exec.Command("gluster", "volume", "start", name)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := GlusterResult{
		Changed:    true,
		Name:       name,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("gluster volume start failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Volume started successfully"
	return result, nil
}

// VolumeStop stops a volume
func VolumeStop(name string) (GlusterResult, error) {
	start := time.Now()

	if name == "" {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("volume name required")
	}

	cmd := exec.Command("gluster", "volume", "stop", name, "force")
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := GlusterResult{
		Changed:    true,
		Name:       name,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("gluster volume stop failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Volume stopped successfully"
	return result, nil
}

// PeerList lists peers
func PeerList() ([]PeerInfo, error) {
	start := time.Now()
	cmd := exec.Command("gluster", "peer", "status")
	output, err := cmd.CombinedOutput()
	_ = time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("gluster peer status failed: %v, output: %s", err, string(output))
	}

	var peers []PeerInfo
	lines := strings.Split(string(output), "\n")
	var current PeerInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "UUID:") {
			if current.UUID != "" {
				peers = append(peers, current)
			}
			current = PeerInfo{UUID: strings.TrimPrefix(line, "UUID: ")}
		} else if strings.HasPrefix(line, "Hostname:") {
			current.Host = strings.TrimPrefix(line, "Hostname: ")
		} else if strings.HasPrefix(line, "State:") {
			state := strings.TrimPrefix(line, "State: ")
			current.State = state
			current.Connected = strings.Contains(state, "Peer in Cluster")
		}
	}

	if current.UUID != "" {
		peers = append(peers, current)
	}

	return peers, nil
}

// PeerProbe adds a new peer
func PeerProbe(host string) (GlusterResult, error) {
	start := time.Now()

	if host == "" {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("host required")
	}

	cmd := exec.Command("gluster", "peer", "probe", host)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := GlusterResult{
		Changed:    true,
		Name:       host,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("gluster peer probe failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Peer probed successfully"
	return result, nil
}

// PeerDetach detaches a peer
func PeerDetach(host string) (GlusterResult, error) {
	start := time.Now()

	if host == "" {
		return GlusterResult{DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("host required")
	}

	cmd := exec.Command("gluster", "peer", "detach", host, "force")
	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	result := GlusterResult{
		Changed:    true,
		Name:       host,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = fmt.Sprintf("gluster peer detach failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("%s", result.Error)
	}

	result.Message = "Peer detached successfully"
	return result, nil
}
