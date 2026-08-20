package tomcat

import (
	"os"
	"testing"
)

// TestFindCatalinaShNotFound tests behavior when catalina.sh is absent.
func TestFindCatalinaShNotFound(t *testing.T) {
	_, err := findCatalinaSh("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent catalina.sh, got nil")
	}
}

// TestFindCatalinaShCustomPath tests custom CATALINA_HOME path.
func TestFindCatalinaShCustomPath(t *testing.T) {
	// Create a temp directory with catalina.sh
	tmpDir := t.TempDir()
	binDir := tmpDir + "/bin"
	os.MkdirAll(binDir, 0755)
	os.WriteFile(binDir+"/catalina.sh", []byte("#!/bin/sh\n"), 0755)

	p, err := findCatalinaSh(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == "" {
		t.Fatal("expected path, got empty string")
	}
}

// TestDeployValidation tests input validation.
func TestDeployValidation(t *testing.T) {
	_, err := Deploy("", "", "")
	if err == nil {
		t.Fatal("expected error for empty war path, got nil")
	}
}

// TestUndeployValidation tests input validation.
func TestUndeployValidation(t *testing.T) {
	_, err := Undeploy("", "")
	if err == nil {
		t.Fatal("expected error for empty context path, got nil")
	}
}

// TestReloadValidation tests input validation.
func TestReloadValidation(t *testing.T) {
	_, err := Reload("", "")
	if err == nil {
		t.Fatal("expected error for empty context path, got nil")
	}
}

// TestStatusEmptyHome tests status with empty catalina home.
func TestStatusEmptyHome(t *testing.T) {
	info, _ := Status("")
	if info.Status == "" {
		t.Fatal("expected non-empty status")
	}
}

// TestListAppsInvalidDir tests list apps with invalid directory.
func TestListAppsInvalidDir(t *testing.T) {
	_, err := ListApps("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent webapps dir, got nil")
	}
}

// TestListAppsEmptyDir tests list apps with empty webapps directory.
func TestListAppsEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(tmpDir+"/webapps", 0755)

	apps, err := ListApps(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 0 {
		t.Fatalf("expected 0 apps, got %d", len(apps))
	}
}

// TestListAppsWithWar tests list apps with a WAR file.
func TestListAppsWithWar(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(tmpDir+"/webapps", 0755)
	os.WriteFile(tmpDir+"/webapps/myapp.war", []byte("fake"), 0644)

	apps, err := ListApps(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].Name != "myapp" {
		t.Fatalf("expected app name 'myapp', got %q", apps[0].Name)
	}
}
