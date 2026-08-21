package nomad

import (
	"testing"
)

func TestJobList(t *testing.T) {
	// May fail if nomad not installed
	_, _ = JobList("")
}

func TestJobRunEmptyFile(t *testing.T) {
	_, err := JobRun("", "")
	if err == nil {
		t.Fatal("JobRun with empty file should return error")
	}
}

func TestJobStopEmptyID(t *testing.T) {
	_, err := JobStop("", "")
	if err == nil {
		t.Fatal("JobStop with empty ID should return error")
	}
}

func TestAllocList(t *testing.T) {
	// May fail if nomad not installed
	_, _ = AllocList("", "")
}

func TestNodeList(t *testing.T) {
	// May fail if nomad not installed
	_, _ = NodeList()
}

func TestNodeDrainEmptyID(t *testing.T) {
	_, err := NodeDrain("", true)
	if err == nil {
		t.Fatal("NodeDrain with empty ID should return error")
	}
}
