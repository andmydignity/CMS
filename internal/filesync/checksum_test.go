package filesync

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestChecksumCalculate(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{"missing file should return err", "", "", true},
		{"valid file should pass", "test.md", "hello everybody my name is markiplier", false},
		{"file without content should pass", "alsotest.md", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tt.filename)
			if tt.filename != "" {
				os.WriteFile(path, []byte(tt.content), 0o666)
			}
			var gotError bool
			var errText string
			_, err := checksumCalculate(path)
			if err == nil {
				gotError = false
			} else {
				errText = err.Error()
				gotError = true
			}
			if tt.wantErr != gotError {
				t.Error(fmt.Sprintf("%v : wantErr gotError mismatch.", tt.name), "Error:"+errText)
			}
		})
	}
}

func TestAppendChecksum(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		checksum string
		wantErr  bool
	}{
		{"empty path should fail", "", "", true},
		{"valid one should pass", "yippe.md", "autism", false},
		{"checksum updating should pass", "yippe.md", "", false},
	}
	db := mockDB(t)
	var gotError bool
	var errText string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appendChecksum(db, tt.path, tt.checksum)
			if err == nil {
				gotError = false
			} else {
				gotError = true
				errText = err.Error()
			}
			if tt.wantErr != gotError {
				t.Error(fmt.Sprintf("%v : wantErr gotError mismatch.", tt.name), "Error:"+errText)
			}
		})
	}
}

func TestCompareChecksum(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		firstChecksum  string
		secondChecksum string
		wantErr        bool
		shouldEqual    bool
	}{
		{"non-matching checksums should return false and pass", "yippe.md", "salsa", "noSalsa", false, false},
		{"matching checksums should return true pass", "turkishChars.md", "şılömüç", "şılömüç", false, true},
	}
	db := mockDB(t)
	var gotErr bool
	var errText string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appendChecksum(db, tt.path, tt.firstChecksum)
			if err != nil {
				t.Error("appendChecksum failed.", err.Error())
			}
			same, err := compareChecksum(db, tt.path, tt.secondChecksum)
			if err == nil {
				gotErr = false
			} else {
				gotErr = true
				errText = err.Error()
			}
			if same != tt.shouldEqual {
				t.Error(fmt.Sprintf("%v : shouldEqual-same mismatch", tt.name), "Error:"+errText)
			}

			if tt.wantErr != gotErr {
				t.Error(fmt.Sprintf("%v :, wantErr gotErr mismatch.", tt.name), "Error:"+errText)
			}
		})
	}
}

func TestDeleteChecksum(t *testing.T) {
	tests := []struct {
		name       string
		createPath string
		deletePath string
		wantErr    bool
	}{
		{"deleting existing path should pass", "yippe.md", "yippe.md", false},
		{"deleting non existing path should fail", "created.md", "deleted.md", true},
	}
	db := mockDB(t)
	var gotErr bool
	var errText string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appendChecksum(db, tt.createPath, "checker of the sums")
			if err != nil {
				t.Error("appendChecksum failed.", err.Error())
			}
			err = deleteChecksum(db, tt.deletePath)
			if err == nil {
				gotErr = false
			} else {
				gotErr = true
				errText = err.Error()
			}

			if tt.wantErr != gotErr {
				t.Error(fmt.Sprintf("%v :, wantErr gotErr mismatch.", tt.name), "Error:"+errText)
			}
		})
	}
}
