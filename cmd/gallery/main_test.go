package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGallery(t *testing.T) {
	out := t.TempDir()
	if err := run(out, `<robot>`, 2); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(data)
	if strings.Count(page, "<article ") != 2 {
		t.Fatal("wrong card count")
	}
	if strings.Contains(page, "<robot>") || !strings.Contains(page, "&lt;robot&gt;-000") {
		t.Fatal("seed must be HTML escaped")
	}
	for _, name := range []string{"preview.png", "contact-sheet.png", "images/000-24.png", "images/000-32.png", "images/001-64.png", "images/001-128.png", "images/001.svg"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []int{0, 501} {
		if err := run(out, "robot", n); err == nil {
			t.Fatal("invalid count accepted")
		}
	}
}
