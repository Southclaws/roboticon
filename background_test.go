package roboticon

import (
	"bytes"
	"image/png"
	"io"
	"strings"
	"testing"
)

func TestTransparentBackground(t *testing.T) {
	r := Generate("robot-a")
	r.Antenna = AntennaLoop
	r.Proportions = ProportionsBalanced
	before := r
	var b bytes.Buffer
	if err := r.RenderPNG(&b, 128, TransparentBackground); err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	// Background, the loop's hole, and the omitted grounding shadow are clear.
	for _, p := range [][2]int{{0, 0}, {64, 19}, {64, 112}} {
		_, _, _, a := im.At(p[0], p[1]).RGBA()
		if a != 0 {
			t.Fatalf("pixel %v has alpha %d, want transparent", p, a)
		}
	}
	_, _, _, alpha := im.At(64, 64).RGBA()
	if alpha != 65535 {
		t.Fatal("face became transparent")
	}
	partial := false
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			_, _, _, a := im.At(x, y).RGBA()
			if a > 0 && a < 65535 {
				partial = true
			}
		}
	}
	if !partial {
		t.Fatal("transparent output lost antialiased edges")
	}
	var repeated bytes.Buffer
	if err := r.RenderPNG(&repeated, 128, TransparentBackground); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b.Bytes(), repeated.Bytes()) {
		t.Fatal("transparent output is not deterministic")
	}
	b.Reset()
	if err := r.RenderSVG(&b, 128, TransparentBackground); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), `<rect x="0" y="0" width="100"`) || strings.Contains(b.String(), `y="86"`) {
		t.Fatal("transparent SVG retained background or shadow")
	}
	if !strings.Contains(b.String(), `<circle cx="50" cy="15" r="4" stroke=`) {
		t.Fatal("antenna loop is not hollow")
	}
	if r != before {
		t.Fatal("background option changed the robot")
	}
}

func TestBackgroundOptions(t *testing.T) {
	r := Generate("robot-a")
	for _, render := range []func(io.Writer, int, ...RenderOption) error{r.RenderPNG, r.RenderSVG} {
		var a, b bytes.Buffer
		if err := render(&a, 32); err != nil {
			t.Fatal(err)
		}
		if err := render(&b, 32, TransparentBackground, GeneratedBackground); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Bytes(), b.Bytes()) {
			t.Fatal("last option should win; generated background is the default")
		}
		b.Reset()
		if err := render(&b, 32, RenderOption(255)); err == nil || b.Len() != 0 {
			t.Fatal("invalid option must fail before writing")
		}
	}
}
