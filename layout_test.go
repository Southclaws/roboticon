package roboticon

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"testing"

	"github.com/fogleman/gg"
)

// Compare painted pixels, not implementation coordinates: facial features must
// stay in the inset, and all inset/decoration pixels must stay on the shell.
func TestLayoutContainment(t *testing.T) {
	for proportions := ProportionsBalanced; proportions <= ProportionsBigHead; proportions++ {
		r := Generate("layout")
		r.Proportions = proportions
		for panel := FacePanelRounded; panel <= FacePanelWide; panel++ {
			r.FacePanel = panel
			var inset scene
			r.drawFacePanel(&inset)
			// The last shape is the face surface, excluding the inset's shadow.
			faceMask := paintLayout(inset[len(inset)-1:], r, true)
			for head := HeadRounded; head <= HeadSquircle; head++ {
				r.Head = head
				var shell scene
				r.drawHead(&shell)
				assertInside(t, faceMask, paintLayout(shell, r, false), fmt.Sprintf("panel=%d head=%d proportions=%d", panel, head, proportions))
			}
			checkFeatures(t, r, faceMask)
			checkMarkingsClear(t, r, faceMask)
		}
		for head := HeadRounded; head <= HeadSquircle; head++ {
			r.Head = head
			var shell scene
			r.drawHead(&shell)
			shellMask := paintLayout(shell, r, false)
			for marking := MarkingNone; marking <= MarkingTempleDots; marking++ {
				r.Marking = marking
				var decoration scene
				r.drawMarking(&decoration)
				assertInside(t, paintLayout(decoration, r, false), shellMask, fmt.Sprintf("marking=%d head=%d proportions=%d", marking, head, proportions))
			}
			for accessory := AccessoryNone; accessory <= AccessoryHeart; accessory++ {
				r.Accessory = accessory
				var decoration scene
				r.drawAccessory(&decoration)
				assertInside(t, paintLayout(decoration, r, false), shellMask, fmt.Sprintf("accessory=%d head=%d proportions=%d", accessory, head, proportions))
			}
		}
	}
}

func checkMarkingsClear(t *testing.T, r Robot, face image.Image) {
	t.Helper()
	for marking := MarkingNone; marking <= MarkingTempleDots; marking++ {
		r.Marking = marking
		var decoration scene
		r.drawMarking(&decoration)
		ink := paintLayout(decoration, r, false)
		for y := 0; y < 150; y++ {
			for x := 0; x < 150; x++ {
				_, _, _, a := ink.At(x, y).RGBA()
				_, _, _, b := face.At(x, y).RGBA()
				if a > 0x8000 && b > 0x8000 {
					t.Fatalf("marking=%d overlaps panel=%d proportions=%d at (%d,%d)", marking, r.FacePanel, r.Proportions, x, y)
				}
			}
		}
	}
}

func checkFeatures(t *testing.T, r Robot, face image.Image) {
	t.Helper()
	for spacing := EyeSpacingNormal; spacing <= EyeSpacingWide; spacing++ {
		for height := EyeHeightNormal; height <= EyeHeightLow; height++ {
			for eyes := EyesDots; eyes <= EyesCurious; eyes++ {
				r.EyeSpacing, r.EyeHeight, r.Eyes = spacing, height, eyes
				var features scene
				r.drawEyes(&features)
				assertInside(t, paintLayout(features, r, true), face, fmt.Sprintf("eyes=%d spacing=%d height=%d panel=%d proportions=%d", eyes, spacing, height, r.FacePanel, r.Proportions))
			}
		}
	}
	for mouth := MouthSmile; mouth <= MouthGrin; mouth++ {
		r.Mouth = mouth
		var features scene
		r.drawMouth(&features)
		assertInside(t, paintLayout(features, r, true), face, fmt.Sprintf("mouth=%d panel=%d proportions=%d", mouth, r.FacePanel, r.Proportions))
	}
	for cheeks := CheeksNone; cheeks <= CheeksLines; cheeks++ {
		r.Cheeks = cheeks
		var features scene
		r.drawCheeks(&features)
		assertInside(t, paintLayout(features, r, true), face, fmt.Sprintf("cheeks=%d panel=%d proportions=%d", cheeks, r.FacePanel, r.Proportions))
	}
}

func paintLayout(s scene, r Robot, face bool) image.Image {
	p := proportionPresets[r.Proportions]
	if face {
		s.transform(layoutTransform{cx: 50, cy: 53, sx: p.faceWidth, sy: p.faceHeight})
	}
	s.transform(layoutTransform{cx: 50, cy: 54, sx: p.headWidth, sy: p.headHeight})
	dc := gg.NewContext(150, 150)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()
	for _, sh := range s {
		drawShape(dc, sh, 1.5)
	}
	return dc.Image()
}

func assertInside(t *testing.T, inside, outside image.Image, label string) {
	t.Helper()
	for y := 0; y < inside.Bounds().Dy(); y++ {
		for x := 0; x < inside.Bounds().Dx(); x++ {
			_, _, _, a := inside.At(x, y).RGBA()
			_, _, _, b := outside.At(x, y).RGBA()
			// Ignore only the faintest antialias fringe; solid feature ink must
			// never escape even the fringe of its containing surface.
			if a > 0x8000 && b == 0 {
				t.Fatalf("%s escapes its surface at (%d,%d)", label, x, y)
			}
		}
	}
}

func TestEachNewTraitChangesAppearance(t *testing.T) {
	setters := []struct {
		name  string
		count int
		set   func(*Robot, int)
	}{
		{"proportions", 4, func(r *Robot, v int) { r.Proportions = Proportions(v) }},
		{"panel", 4, func(r *Robot, v int) { r.FacePanel = FacePanel(v) }},
		{"spacing", 3, func(r *Robot, v int) { r.EyeSpacing = EyeSpacing(v) }},
		{"height", 3, func(r *Robot, v int) { r.EyeHeight = EyeHeight(v) }},
		{"cheeks", 4, func(r *Robot, v int) { r.Cheeks = Cheeks(v) }},
		{"marking", 8, func(r *Robot, v int) { r.Marking = Marking(v) }},
	}
	for _, trait := range setters {
		t.Run(trait.name, func(t *testing.T) {
			r := Generate("robot-a")
			seen := map[string]bool{}
			for v := 0; v < trait.count; v++ {
				trait.set(&r, v)
				var b bytes.Buffer
				if err := r.RenderPNG(&b, 128); err != nil {
					t.Fatal(err)
				}
				if seen[b.String()] {
					t.Fatalf("variant %d makes no visual difference", v)
				}
				seen[b.String()] = true
			}
			trait.set(&r, 255)
			if err := r.RenderPNG(io.Discard, 32); err == nil {
				t.Fatal("invalid trait accepted")
			}
			if err := r.RenderSVG(io.Discard, 32); err == nil {
				t.Fatal("invalid trait accepted")
			}
		})
	}
}
