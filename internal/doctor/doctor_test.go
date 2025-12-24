package doctor

import (
	"bytes"
	"testing"
)

func TestDoctor(t *testing.T) {
	doc := NewDoctor()
	if doc == nil {
		t.Fatal("NewDoctor returned nil")
	}
}

func TestCheckResults(t *testing.T) {
	results := &CheckResults{
		Checks: []CheckResult{
			{Name: "test1", Status: "PASS", Message: "OK"},
			{Name: "test2", Status: "PASS", Message: "OK"},
		},
	}

	if !results.AllPassed() {
		t.Error("AllPassed should return true when all checks pass")
	}

	results.Checks = append(results.Checks, CheckResult{
		Name:    "test3",
		Status:  "FAIL",
		Message: "Error",
	})

	if results.AllPassed() {
		t.Error("AllPassed should return false when any check fails")
	}
}

func TestPrintText(t *testing.T) {
	results := &CheckResults{
		Checks: []CheckResult{
			{Name: "kubectl", Status: "PASS", Message: "Found"},
			{Name: "cluster", Status: "FAIL", Message: "Not found"},
		},
	}

	var buf bytes.Buffer
	results.PrintText(&buf)

	output := buf.String()
	if output == "" {
		t.Error("PrintText should produce output")
	}

	if !bytes.Contains(buf.Bytes(), []byte("kubectl")) {
		t.Error("Output should contain check names")
	}
}

func TestPrintJSON(t *testing.T) {
	results := &CheckResults{
		Checks: []CheckResult{
			{Name: "test", Status: "PASS", Message: "OK"},
		},
	}

	var buf bytes.Buffer
	err := results.PrintJSON(&buf)
	if err != nil {
		t.Errorf("PrintJSON failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("PrintJSON should produce output")
	}
}
