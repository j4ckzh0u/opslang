package process

import "testing"

func TestParseJavaCommandLine(t *testing.T) {
	cmd := "/usr/bin/java -Xmx512m -cp /opt/app/app-1.2.3.jar:/opt/lib/log4j-core-2.20.0.jar com.example.Main"
	libs := parseJavaLibraries(cmd)
	if len(libs) != 2 {
		t.Fatalf("libraries = %+v, want 2", libs)
	}
	if libs[0].Name != "app" || libs[0].Version != "1.2.3" {
		t.Errorf("first library = %+v", libs[0])
	}
	if libs[1].Name != "log4j-core" || libs[1].Version != "2.20.0" {
		t.Errorf("second library = %+v", libs[1])
	}
}

func TestParseContainerID(t *testing.T) {
	for _, tc := range []struct {
		name, wantRuntime, wantID string
	}{
		{"0::/docker/0123456789abcdef0123456789abcdef", "docker", "0123456789abcdef0123456789abcdef"},
		{"0::/kubepods.slice/docker-abcdef.scope", "docker", "abcdef"},
	} {
		runtime, id := parseContainerCgroup(tc.name)
		if runtime != tc.wantRuntime || id != tc.wantID {
			t.Errorf("parseContainerCgroup(%q) = %q,%q; want %q,%q", tc.name, runtime, id, tc.wantRuntime, tc.wantID)
		}
	}
}
