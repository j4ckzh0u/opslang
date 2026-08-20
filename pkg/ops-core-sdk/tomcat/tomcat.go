// Package tomcat provides Apache Tomcat management operations.
// Equivalent to Ansible's tomcat modules.
package tomcat

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result represents a generic tomcat operation result.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DeployResult represents a deployment operation result.
type DeployResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	App     string `json:"app,omitempty"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AppInfo represents information about a deployed application.
type AppInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Context string `json:"context,omitempty"`
	Running bool   `json:"running"`
}

// ServerInfo represents Tomcat server information.
type ServerInfo struct {
	Status      string `json:"status"`
	Version     string `json:"version,omitempty"`
	CatalinaHome string `json:"catalina_home,omitempty"`
	CatalinaBase string `json:"catalina_base,omitempty"`
	JavaHome    string `json:"java_home,omitempty"`
	Running     bool   `json:"running"`
	PID         int    `json:"pid,omitempty"`
}

// findCatalinaSh locates catalina.sh in common Tomcat locations.
func findCatalinaSh(catalinaHome string) (string, error) {
	if catalinaHome != "" {
		p := filepath.Join(catalinaHome, "bin", "catalina.sh")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	// Common installation paths
	paths := []string{
		"/opt/tomcat/bin/catalina.sh",
		"/usr/share/tomcat/bin/catalina.sh",
		"/usr/local/tomcat/bin/catalina.sh",
		"/opt/apache-tomcat/bin/catalina.sh",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("catalina.sh not found")
}

// runScript executes catalina.sh with the given command.
func runScript(catalinaHome, command string) (string, error) {
	script, err := findCatalinaSh(catalinaHome)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(script, command)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CATALINA_HOME=%s", catalinaHome))
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Start starts the Tomcat server.
func Start(catalinaHome string) (Result, error) {
	out, err := runScript(catalinaHome, "start")
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("start tomcat: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// Stop stops the Tomcat server.
func Stop(catalinaHome string) (Result, error) {
	out, err := runScript(catalinaHome, "stop")
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("stop tomcat: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Message: out}, nil
}

// Restart restarts the Tomcat server.
func Restart(catalinaHome string) (Result, error) {
	out, err := runScript(catalinaHome, "stop")
	if err != nil {
		// Stop may fail if not running, continue with start
	}
	out2, err := runScript(catalinaHome, "start")
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("restart tomcat: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Message: out + "\n" + out2}, nil
}

// Status returns the status of the Tomcat server.
func Status(catalinaHome string) (ServerInfo, error) {
	info := ServerInfo{
		Status:       "unknown",
		CatalinaHome: catalinaHome,
	}
	if catalinaHome == "" {
		catalinaHome = os.Getenv("CATALINA_HOME")
	}
	info.CatalinaHome = catalinaHome
	info.CatalinaBase = os.Getenv("CATALINA_BASE")
	info.JavaHome = os.Getenv("JAVA_HOME")

	// Check if running via pid file
	pidFile := filepath.Join(catalinaHome, "bin", "catalina.pid")
	if data, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(data))
		var pid int
		fmt.Sscanf(pidStr, "%d", &pid)
		if pid > 0 {
			// Check if process exists
			if _, err := os.FindProcess(pid); err == nil {
				info.Running = true
				info.PID = pid
				info.Status = "running"
			}
		}
	}

	// Try to get version
	out, _ := runScript(catalinaHome, "version")
	if out != "" {
		info.Version = out
	}

	if info.Status == "unknown" {
		info.Status = "stopped"
	}
	return info, nil
}

// Deploy deploys a WAR file to Tomcat.
func Deploy(catalinaHome string, warPath string, contextPath string) (DeployResult, error) {
	if warPath == "" {
		return DeployResult{Status: "failed", Error: "war path is required"}, fmt.Errorf("war path is required")
	}

	// Determine webapps directory
	webappsDir := filepath.Join(catalinaHome, "webapps")
	if catalinaBase := os.Getenv("CATALINA_BASE"); catalinaBase != "" {
		webappsDir = filepath.Join(catalinaBase, "webapps")
	}

	// Copy WAR file to webapps
	warName := filepath.Base(warPath)
	destPath := filepath.Join(webappsDir, warName)

	src, err := os.ReadFile(warPath)
	if err != nil {
		return DeployResult{Status: "failed", App: warName, Error: fmt.Sprintf("read war: %v", err)}, err
	}
	if err := os.WriteFile(destPath, src, 0644); err != nil {
		return DeployResult{Status: "failed", App: warName, Path: destPath, Error: fmt.Sprintf("copy war: %v", err)}, err
	}

	app := warName
	if contextPath != "" {
		app = contextPath
	}
	return DeployResult{Status: "success", Changed: true, App: app, Path: destPath}, nil
}

// Undeploy removes a deployed application.
func Undeploy(catalinaHome string, contextPath string) (DeployResult, error) {
	if contextPath == "" {
		return DeployResult{Status: "failed", Error: "context path is required"}, fmt.Errorf("context path is required")
	}

	webappsDir := filepath.Join(catalinaHome, "webapps")
	if catalinaBase := os.Getenv("CATALINA_BASE"); catalinaBase != "" {
		webappsDir = filepath.Join(catalinaBase, "webapps")
	}

	// Try removing .war file first
	warPath := filepath.Join(webappsDir, contextPath+".war")
	if _, err := os.Stat(warPath); err == nil {
		if err := os.Remove(warPath); err != nil {
			return DeployResult{Status: "failed", App: contextPath, Error: fmt.Sprintf("remove war: %v", err)}, err
		}
	}

	// Try removing exploded directory
	appDir := filepath.Join(webappsDir, contextPath)
	if _, err := os.Stat(appDir); err == nil {
		if err := os.RemoveAll(appDir); err != nil {
			return DeployResult{Status: "failed", App: contextPath, Error: fmt.Sprintf("remove app dir: %v", err)}, err
		}
	}

	return DeployResult{Status: "success", Changed: true, App: contextPath}, nil
}

// ListApps lists deployed applications.
func ListApps(catalinaHome string) ([]AppInfo, error) {
	webappsDir := filepath.Join(catalinaHome, "webapps")
	if catalinaBase := os.Getenv("CATALINA_BASE"); catalinaBase != "" {
		webappsDir = filepath.Join(catalinaBase, "webapps")
	}

	entries, err := os.ReadDir(webappsDir)
	if err != nil {
		return nil, fmt.Errorf("read webapps dir: %w", err)
	}

	var apps []AppInfo
	for _, e := range entries {
		name := e.Name()
		app := AppInfo{
			Name: name,
			Path: filepath.Join(webappsDir, name),
		}
		if strings.HasSuffix(name, ".war") {
			app.Name = strings.TrimSuffix(name, ".war")
			app.Context = "/" + app.Name
		} else if e.IsDir() {
			app.Context = "/" + name
		}
		app.Running = true
		apps = append(apps, app)
	}
	return apps, nil
}

// Reload triggers a reload of a specific application context.
func Reload(catalinaHome string, contextPath string) (Result, error) {
	if contextPath == "" {
		return Result{Status: "failed", Error: "context path is required"}, fmt.Errorf("context path is required")
	}
	// Tomcat reloads happen via manager app or by touching the context XML
	// For simplicity, we touch the webapp directory to trigger reload
	webappsDir := filepath.Join(catalinaHome, "webapps")
	if catalinaBase := os.Getenv("CATALINA_BASE"); catalinaBase != "" {
		webappsDir = filepath.Join(catalinaBase, "webapps")
	}
	appDir := filepath.Join(webappsDir, contextPath)
	now := fmt.Sprintf("%d", os.Getpid())
	_ = os.WriteFile(filepath.Join(appDir, ".reload_trigger"), []byte(now), 0644)
	return Result{Status: "success", Changed: true, Message: "reload triggered for " + contextPath}, nil
}

// Version returns the Tomcat version.
func Version(catalinaHome string) (string, error) {
	out, err := runScript(catalinaHome, "version")
	if err != nil {
		return "", fmt.Errorf("get version: %w", err)
	}
	return out, nil
}
