// Package kubernetes provides Kubernetes resource management operations via kubectl.
package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result is returned by mutating operations (apply/delete/scale/etc).
type Result struct {
	Changed  bool   `json:"changed"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ResourceResult represents a single Kubernetes resource returned by Get.
type ResourceResult struct {
	Name       string                 `json:"name"`
	Kind       string                 `json:"kind"`
	Namespace  string                 `json:"namespace"`
	APIVersion string                 `json:"api_version"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
	Raw        map[string]interface{} `json:"raw,omitempty"`
}

// PodInfo represents a Kubernetes pod.
type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Ready     string            `json:"ready"`
	Restarts  int               `json:"restarts"`
	Age       string            `json:"age"`
	Node      string            `json:"node"`
	IP        string            `json:"ip"`
	Labels    map[string]string `json:"labels,omitempty"`
	Containers []string         `json:"containers,omitempty"`
}

// ServiceInfo represents a Kubernetes service.
type ServiceInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Type       string `json:"type"`
	ClusterIP  string `json:"cluster_ip"`
	ExternalIP string `json:"external_ip,omitempty"`
	Ports      string `json:"ports"`
	Age        string `json:"age"`
	Selector   map[string]string `json:"selector,omitempty"`
}

// DeploymentInfo represents a Kubernetes deployment.
type DeploymentInfo struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Ready             string `json:"ready"`
	UpToDate          int    `json:"up_to_date"`
	Available         int    `json:"available"`
	Age               string `json:"age"`
	Containers        []string `json:"containers,omitempty"`
	Images            []string `json:"images,omitempty"`
	Replicas          int    `json:"replicas"`
	ReadyReplicas     int    `json:"ready_replicas"`
	AvailableReplicas int    `json:"available_replicas"`
}

// ExecResult is returned by Exec.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Pod      string `json:"pod"`
	Container string `json:"container,omitempty"`
}

// LogsResult is returned by Logs.
type LogsResult struct {
	Logs      string `json:"logs"`
	Pod       string `json:"pod"`
	Container string `json:"container,omitempty"`
	Lines     int    `json:"lines"`
}

// kubectl checks that kubectl is available and returns its path.
func kubectl() (string, error) {
	path, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl not found in PATH: install kubectl to use kubernetes functions")
	}
	return path, nil
}

// runKubectl executes kubectl with the given args, returning stdout and stderr.
func runKubectl(args ...string) (string, string, error) {
	bin, err := kubectl()
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

// isInlineYAML reports whether manifest looks like inline YAML rather than a file path.
func isInlineYAML(s string) bool {
	return strings.Contains(s, "\n") || strings.Contains(s, "apiVersion:") || strings.Contains(s, "kind:")
}

// manifestArg returns the kubectl -f argument for a manifest.
// If it looks like inline YAML, writes to a temp file and returns its path;
// otherwise treats it as a file path. Caller must clean up the temp file when
// returned clean is non-nil.
func manifestArg(manifest string) (string, func(), error) {
	if isInlineYAML(manifest) {
		tmp, err := os.CreateTemp("", "ops-k8s-*.yaml")
		if err != nil {
			return "", nil, fmt.Errorf("failed to create temp manifest: %w", err)
		}
		if _, err := tmp.WriteString(manifest); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", nil, fmt.Errorf("failed to write temp manifest: %w", err)
		}
		tmp.Close()
		return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
	}
	if _, err := os.Stat(manifest); err != nil {
		return "", nil, fmt.Errorf("manifest file not found: %w", err)
	}
	abs, err := filepath.Abs(manifest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve manifest path: %w", err)
	}
	return abs, func() {}, nil
}

// Apply applies a YAML manifest (inline or file path) to the cluster.
func Apply(manifest string, namespace string, dryRun bool) (Result, error) {
	farg, cleanup, err := manifestArg(manifest)
	if err != nil {
		return Result{}, fmt.Errorf("kubernetes.apply: %w", err)
	}
	defer cleanup()

	args := []string{"apply", "-f", farg}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if dryRun {
		args = append(args, "--dry-run=client")
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return Result{Error: stderr}, fmt.Errorf("kubectl apply failed: %s: %w", stderr, err)
	}
	return Result{Changed: true, Status: "applied", Message: strings.TrimSpace(stdout)}, nil
}

// Delete deletes resources described by the YAML manifest (inline or file path).
func Delete(manifest string, namespace string) (Result, error) {
	farg, cleanup, err := manifestArg(manifest)
	if err != nil {
		return Result{}, fmt.Errorf("kubernetes.delete: %w", err)
	}
	defer cleanup()

	args := []string{"delete", "-f", farg}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return Result{Error: stderr}, fmt.Errorf("kubectl delete failed: %s: %w", stderr, err)
	}
	return Result{Changed: true, Status: "deleted", Message: strings.TrimSpace(stdout)}, nil
}

// Get retrieves a single resource by type and name.
func Get(resourceType string, name string, namespace string) (ResourceResult, error) {
	args := []string{"get", resourceType, name, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return ResourceResult{}, fmt.Errorf("kubectl get failed: %s: %w", stderr, err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		return ResourceResult{}, fmt.Errorf("failed to parse kubectl output: %w", err)
	}
	return parseResource(raw), nil
}

// List retrieves resources of a given type, optionally filtered by labels (e.g. "app=web,env=prod").
func List(resourceType string, namespace string, labels string) ([]ResourceResult, error) {
	args := []string{"get", resourceType, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	if labels != "" {
		args = append(args, "-l", labels)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get failed: %s: %w", stderr, err)
	}
	var wrapper struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse kubectl output: %w", err)
	}
	out := make([]ResourceResult, 0, len(wrapper.Items))
	for _, item := range wrapper.Items {
		out = append(out, parseResource(item))
	}
	return out, nil
}

// parseResource extracts ResourceResult fields from raw kubectl JSON.
func parseResource(raw map[string]interface{}) ResourceResult {
	r := ResourceResult{Raw: raw}
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		r.Name, _ = meta["name"].(string)
		r.Namespace, _ = meta["namespace"].(string)
		if lbls, ok := meta["labels"].(map[string]interface{}); ok {
			r.Labels = toStringMap(lbls)
		}
		if anns, ok := meta["annotations"].(map[string]interface{}); ok {
			r.Annotations = toStringMap(anns)
		}
	}
	r.Kind, _ = raw["kind"].(string)
	r.APIVersion, _ = raw["apiVersion"].(string)
	return r
}

func toStringMap(m map[string]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

// CreateNamespace creates a new namespace.
func CreateNamespace(name string) (Result, error) {
	args := []string{"create", "namespace", name}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		if strings.Contains(stderr, "already exists") {
			return Result{Changed: false, Status: "exists", Message: "namespace already exists"}, nil
		}
		return Result{Error: stderr}, fmt.Errorf("kubectl create namespace failed: %s: %w", stderr, err)
	}
	return Result{Changed: true, Status: "created", Message: strings.TrimSpace(stdout)}, nil
}

// DeleteNamespace deletes a namespace.
func DeleteNamespace(name string) (Result, error) {
	args := []string{"delete", "namespace", name, "--ignore-not-found"}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return Result{Error: stderr}, fmt.Errorf("kubectl delete namespace failed: %s: %w", stderr, err)
	}
	return Result{Changed: true, Status: "deleted", Message: strings.TrimSpace(stdout)}, nil
}

// GetPods lists pods in a namespace, optionally filtered by labels.
func GetPods(namespace string, labels string) ([]PodInfo, error) {
	args := []string{"get", "pods", "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	if labels != "" {
		args = append(args, "-l", labels)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get pods failed: %s: %w", stderr, err)
	}
	var wrapper struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse pod list: %w", err)
	}
	pods := make([]PodInfo, 0, len(wrapper.Items))
	for _, item := range wrapper.Items {
		pods = append(pods, parsePod(item))
	}
	return pods, nil
}

func parsePod(raw map[string]interface{}) PodInfo {
	p := PodInfo{}
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		p.Name, _ = meta["name"].(string)
		p.Namespace, _ = meta["namespace"].(string)
		if lbls, ok := meta["labels"].(map[string]interface{}); ok {
			p.Labels = toStringMap(lbls)
		}
	}
	if status, ok := raw["status"].(map[string]interface{}); ok {
		p.Status, _ = status["phase"].(string)
		p.IP, _ = status["podIP"].(string)
	}
	if spec, ok := raw["spec"].(map[string]interface{}); ok {
		p.Node, _ = spec["nodeName"].(string)
		if cs, ok := spec["containers"].([]interface{}); ok {
			for _, c := range cs {
				if cm, ok := c.(map[string]interface{}); ok {
					if n, ok := cm["name"].(string); ok {
						p.Containers = append(p.Containers, n)
					}
				}
			}
		}
	}
	return p
}

// GetServices lists services in a namespace.
func GetServices(namespace string) ([]ServiceInfo, error) {
	args := []string{"get", "svc", "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get svc failed: %s: %w", stderr, err)
	}
	var wrapper struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse service list: %w", err)
	}
	svcs := make([]ServiceInfo, 0, len(wrapper.Items))
	for _, item := range wrapper.Items {
		svcs = append(svcs, parseService(item))
	}
	return svcs, nil
}

func parseService(raw map[string]interface{}) ServiceInfo {
	s := ServiceInfo{}
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		s.Name, _ = meta["name"].(string)
		s.Namespace, _ = meta["namespace"].(string)
	}
	if spec, ok := raw["spec"].(map[string]interface{}); ok {
		s.Type, _ = spec["type"].(string)
		s.ClusterIP, _ = spec["clusterIP"].(string)
		if ips, ok := spec["externalIPs"].([]interface{}); ok && len(ips) > 0 {
			if ip, ok := ips[0].(string); ok {
				s.ExternalIP = ip
			}
		}
		if sel, ok := spec["selector"].(map[string]interface{}); ok {
			s.Selector = toStringMap(sel)
		}
	}
	return s
}

// GetDeployments lists deployments in a namespace.
func GetDeployments(namespace string) ([]DeploymentInfo, error) {
	args := []string{"get", "deploy", "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return nil, fmt.Errorf("kubectl get deploy failed: %s: %w", stderr, err)
	}
	var wrapper struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse deployment list: %w", err)
	}
	deploys := make([]DeploymentInfo, 0, len(wrapper.Items))
	for _, item := range wrapper.Items {
		deploys = append(deploys, parseDeployment(item))
	}
	return deploys, nil
}

func parseDeployment(raw map[string]interface{}) DeploymentInfo {
	d := DeploymentInfo{}
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		d.Name, _ = meta["name"].(string)
		d.Namespace, _ = meta["namespace"].(string)
	}
	if spec, ok := raw["spec"].(map[string]interface{}); ok {
		if r, ok := spec["replicas"].(float64); ok {
			d.Replicas = int(r)
		}
		if tmpl, ok := spec["template"].(map[string]interface{}); ok {
			if ps, ok := tmpl["spec"].(map[string]interface{}); ok {
				if cs, ok := ps["containers"].([]interface{}); ok {
					for _, c := range cs {
						if cm, ok := c.(map[string]interface{}); ok {
							if n, ok := cm["name"].(string); ok {
								d.Containers = append(d.Containers, n)
							}
							if img, ok := cm["image"].(string); ok {
								d.Images = append(d.Images, img)
							}
						}
					}
				}
			}
		}
	}
	if status, ok := raw["status"].(map[string]interface{}); ok {
		if r, ok := status["readyReplicas"].(float64); ok {
			d.ReadyReplicas = int(r)
		}
		if a, ok := status["availableReplicas"].(float64); ok {
			d.AvailableReplicas = int(a)
		}
		if u, ok := status["updatedReplicas"].(float64); ok {
			d.UpToDate = int(u)
		}
		d.Available = d.AvailableReplicas
		d.Ready = fmt.Sprintf("%d/%d", d.ReadyReplicas, d.Replicas)
	}
	return d
}

// Scale sets the replica count of a deployment.
func Scale(deployment string, replicas int, namespace string) (Result, error) {
	args := []string{"scale", "deployment/" + deployment, fmt.Sprintf("--replicas=%d", replicas)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return Result{Error: stderr}, fmt.Errorf("kubectl scale failed: %s: %w", stderr, err)
	}
	return Result{Changed: true, Status: "scaled", Message: strings.TrimSpace(stdout)}, nil
}

// RolloutStatus checks the rollout status of a deployment.
func RolloutStatus(deployment string, namespace string) (Result, error) {
	args := []string{"rollout", "status", "deployment/" + deployment}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return Result{Error: stderr, Status: "failed"}, fmt.Errorf("kubectl rollout status failed: %s: %w", stderr, err)
	}
	return Result{Changed: false, Status: "ready", Message: strings.TrimSpace(stdout)}, nil
}

// Exec executes a command in a pod container.
func Exec(pod string, command string, namespace string, container string) (ExecResult, error) {
	args := []string{"exec", pod}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, "--", "sh", "-c", command)
	bin, err := kubectl()
	if err != nil {
		return ExecResult{}, fmt.Errorf("kubernetes.exec: %w", err)
	}
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return ExecResult{ExitCode: -1, Stderr: stderr.String(), Pod: pod, Container: container},
				fmt.Errorf("kubectl exec failed: %w", err)
		}
	}
	return ExecResult{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  exitCode,
		Pod:       pod,
		Container: container,
	}, nil
}

// Logs retrieves logs from a pod container.
func Logs(pod string, namespace string, container string, tail int) (LogsResult, error) {
	args := []string{"logs", pod}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if container != "" {
		args = append(args, "-c", container)
	}
	if tail > 0 {
		args = append(args, fmt.Sprintf("--tail=%d", tail))
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return LogsResult{}, fmt.Errorf("kubectl logs failed: %s: %w", stderr, err)
	}
	lines := 0
	if stdout != "" {
		lines = strings.Count(stdout, "\n")
		if !strings.HasSuffix(stdout, "\n") {
			lines++
		}
	}
	return LogsResult{
		Logs:      stdout,
		Pod:       pod,
		Container: container,
		Lines:     lines,
	}, nil
}

// WaitReady waits for a resource to become ready within the given timeout (seconds).
func WaitReady(resourceType string, name string, namespace string, timeout int) (Result, error) {
	args := []string{"wait", resourceType + "/" + name, "--for=condition=ready", fmt.Sprintf("--timeout=%ds", timeout)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	stdout, stderr, err := runKubectl(args...)
	if err != nil {
		return Result{Error: stderr, Status: "timeout"}, fmt.Errorf("kubectl wait failed: %s: %w", stderr, err)
	}
	return Result{Changed: false, Status: "ready", Message: strings.TrimSpace(stdout)}, nil
}
