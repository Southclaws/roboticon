package roboticon

import "math"

var proportionPresets = [...]struct{ headWidth, headHeight, faceWidth, faceHeight float64 }{
	{1, 1, 1, 1},
	{1.06, .96, 1.02, .96},
	{.96, 1.04, .96, 1.03},
	{1.04, 1.02, .94, .96},
}

// Scale around a shared anchor. Circles stay circular and corner radii/strokes
// use the smaller scale, so proportional variation never introduces sharp edges.
type layoutTransform struct{ cx, cy, sx, sy, dy float64 }

// Explicit float64 conversions prevent native fused multiply-add operations
// from disagreeing with WASM. Snap to a 0.0001-unit logical grid so the geometry
// also has readable, stable SVG coordinates (well below a pixel at any size).
func logicalCoordinate(v float64) float64 { return math.Round(v*10000) / 10000 }

func (s scene) transform(t layoutTransform) {
	for i := range s {
		sh := &s[i]
		sh.x = logicalCoordinate(t.cx + float64((sh.x-t.cx)*t.sx))
		sh.y = logicalCoordinate(t.cy + float64((sh.y-t.cy)*t.sy) + t.dy)
		sh.w = logicalCoordinate(sh.w * t.sx)
		sh.h = logicalCoordinate(sh.h * t.sy)
		sh.radius = logicalCoordinate(sh.radius * math.Min(t.sx, t.sy))
		sh.width = logicalCoordinate(sh.width * math.Min(t.sx, t.sy))
		for _, c := range sh.path {
			for j := 0; j < len(c.v); j += 2 {
				c.v[j] = logicalCoordinate(t.cx + float64((c.v[j]-t.cx)*t.sx))
				c.v[j+1] = logicalCoordinate(t.cy + float64((c.v[j+1]-t.cy)*t.sy) + t.dy)
			}
		}
	}
}

func (r Robot) drawFacePanel(s *scene) {
	panel := [...]struct{ x, y, w, h, r float64 }{
		{24, 34, 52, 37, 13},
		{24, 34, 52, 37, 18.5},
		{26, 32.5, 48, 40, 14},
		{22.5, 35, 55, 35, 13},
	}[r.FacePanel]
	s.rect(panel.x, panel.y+1, panel.w, panel.h+1, panel.r, shade(r.Palette.Body, .85))
	s.rect(panel.x, panel.y, panel.w, panel.h, panel.r, r.Palette.Face)
}

func (r Robot) drawCheeks(s *scene) {
	for _, x := range []float64{32, 68} {
		switch r.Cheeks {
		case CheeksDots:
			s.circle(x, 57.5, 1.3, r.Palette.Accent)
		case CheeksBlush:
			s.rect(x-2.5, 56.5, 5, 2.8, 1.4, r.Palette.Accent)
		case CheeksLines:
			for _, dx := range []float64{-1.6, 1.6} {
				s.curve(r.Palette.Accent, 1.3, move(x+dx-.5, 56.5), line(x+dx+.5, 58.5))
			}
		}
	}
}

// Markings and accessories occupy separate areas of the shell. None of these
// cover an eye or mouth, and every non-empty marking remains visible with a badge.
func (r Robot) drawMarking(s *scene) {
	p := r.Palette
	ink := shade(p.Body, .76)
	switch r.Marking {
	case MarkingBolts:
		for _, x := range []float64{33, 41} {
			s.circle(x, 77, 1.6, ink)
			s.circle(x, 77, .65, p.Face)
		}
	case MarkingSerialDots:
		for _, x := range []float64{33, 37, 41} {
			s.circle(x, 77, 1, ink)
		}
	case MarkingForeheadPanel:
		s.rect(43, 28.5, 14, 2.5, 1.25, ink)
		s.rect(45, 29.2, 5, 1.1, .55, p.Face)
	case MarkingVents:
		for _, x := range []float64{33, 37, 41} {
			s.rect(x-.7, 75, 1.4, 4, .7, ink)
		}
	case MarkingSeam:
		s.curve(ink, 1.2, move(29, 74.5), quad(31, 79, 38, 79), line(47, 79))
	case MarkingChinStripe:
		s.rect(32, 75.5, 10, 2.8, 1.4, p.Accent)
	case MarkingTempleDots:
		for _, x := range []float64{21, 79} {
			for _, y := range []float64{50, 54, 58} {
				s.circle(x, y, .7, ink)
			}
		}
	}
}
