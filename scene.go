package roboticon

import "image/color"

// A small display list is shared by both renderers. All dimensions are in the
// logical 100x100 canvas; only the renderer knows the output size. Keep robot
// geometry inside the circle centered at (50,50) with radius 47. Outputs remain
// square; the extra three units provide clearance for downstream circle crops.
type shape struct {
	kind               byte
	x, y, w, h, radius float64
	fill               color.RGBA
	stroke             color.RGBA
	width              float64
	path               []command
}
type command struct {
	op byte
	v  []float64
}
type scene []shape

func (s *scene) rect(x, y, w, h, r float64, c color.RGBA) {
	*s = append(*s, shape{kind: 'r', x: x, y: y, w: w, h: h, radius: r, fill: c})
}
func (s *scene) circle(x, y, r float64, c color.RGBA) {
	*s = append(*s, shape{kind: 'c', x: x, y: y, radius: r, fill: c})
}
func (s *scene) curve(c color.RGBA, width float64, commands ...command) {
	*s = append(*s, shape{kind: 'p', stroke: c, width: width, path: commands})
}
func (s *scene) filled(c color.RGBA, commands ...command) {
	*s = append(*s, shape{kind: 'p', fill: c, path: commands})
}
func move(x, y float64) command              { return command{'M', []float64{x, y}} }
func line(x, y float64) command              { return command{'L', []float64{x, y}} }
func quad(cx, cy, x, y float64) command      { return command{'Q', []float64{cx, cy, x, y}} }
func cubic(a, b, c, d, x, y float64) command { return command{'C', []float64{a, b, c, d, x, y}} }
func closePath() command                     { return command{op: 'Z'} }

func shade(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * factor), uint8(float64(c.G) * factor), uint8(float64(c.B) * factor), c.A}
}

func (r Robot) scene(showBackground bool) scene {
	var s scene
	p := r.Palette
	if showBackground {
		s.rect(0, 0, 100, 100, 0, p.Background)
		// A quiet grounding shadow; the head still has ample breathing room.
		s.rect(29, 86, 42, 3, 1.5, shade(p.Background, .94))
	}
	var body, face scene
	r.drawAntenna(&body, showBackground)
	r.drawEars(&body)
	r.drawHead(&body)
	r.drawFacePanel(&face)
	r.drawEyes(&face)
	r.drawMouth(&face)
	r.drawCheeks(&face)
	layout := proportionPresets[r.Proportions]
	face.transform(layoutTransform{cx: 50, cy: 53, sx: layout.faceWidth, sy: layout.faceHeight})
	body = append(body, face...)
	r.drawMarking(&body)
	r.drawAccessory(&body)
	body.transform(layoutTransform{cx: 50, cy: 54, sx: layout.headWidth, sy: layout.headHeight})
	s = append(s, body...)
	return s
}

func (r Robot) drawHead(s *scene) {
	p := r.Palette
	switch r.Head {
	case HeadRounded:
		s.rect(17, 25, 66, 58, 17, p.Body)
	case HeadPill:
		s.rect(15, 27, 70, 55, 25, p.Body)
	case HeadTall:
		s.rect(20, 22, 60, 62, 20, p.Body)
	case HeadDome:
		s.filled(p.Body, move(18, 55), cubic(18, 32, 28, 22, 50, 22), cubic(72, 22, 82, 32, 82, 55), line(82, 65), quad(82, 83, 64, 83), line(36, 83), quad(18, 83, 18, 65), closePath())
	case HeadSquircle:
		s.filled(p.Body, move(50, 24), cubic(79, 24, 83, 29, 83, 53), cubic(83, 78, 79, 83, 50, 83), cubic(21, 83, 17, 78, 17, 53), cubic(17, 29, 21, 24, 50, 24), closePath())
	}
	// Short soft highlight, not a hard outline.
	if r.Marking != MarkingForeheadPanel {
		y := [...]float64{28, 30, 25, 26, 27}[r.Head]
		s.rect(44, y, 12, 2, 1, color.RGBA{90, 90, 90, 90})
	}
}

func (r Robot) drawAntenna(s *scene, showBackground bool) {
	p := r.Palette
	stem := shade(p.Body, .77)
	switch r.Antenna {
	case AntennaBobble:
		s.curve(stem, 3, move(50, 28), line(50, 16))
		s.circle(50, 14, 4.5, p.Accent)
		s.circle(48.8, 12.8, 1.2, p.Face)
	case AntennaTwin:
		for _, x := range []float64{34, 66} {
			d := (x - 50) / 8
			s.curve(stem, 2.8, move(x, 29), quad(x, 22, x+d, 17))
			s.circle(x+d, 15.5, 3.5, p.Accent)
		}
	case AntennaLoop:
		s.curve(stem, 3, move(50, 27), line(50, 20))
		if showBackground {
			s.circle(50, 15, 5.5, p.Accent)
			s.circle(50, 15, 2.5, p.Background)
		} else {
			*s = append(*s, shape{kind: 'c', x: 50, y: 15, radius: 4, stroke: p.Accent, width: 3})
		}
	case AntennaSprout:
		s.curve(stem, 2.8, move(50, 28), quad(53, 18, 46, 14))
		s.filled(p.Accent, move(48, 17), cubic(37, 18, 37, 8, 39, 9), cubic(46, 8, 51, 12, 48, 17), closePath())
	}
}

func (r Robot) drawEars(s *scene) {
	p := r.Palette
	for _, x := range []float64{11, 79} {
		switch r.Ears {
		case EarsPads:
			s.rect(x, 44, 10, 20, 5, shade(p.Body, .85))
			s.rect(x+2, 47, 6, 12, 3, p.Accent)
		case EarsDiscs:
			s.circle(x+5, 54, 7, shade(p.Body, .86))
			s.circle(x+5, 54, 3.5, p.Accent)
		case EarsNubs:
			s.rect(x, 49, 11, 9, 4.5, p.Accent)
		}
	}
}

func (r Robot) drawEyes(s *scene) {
	p := r.Palette
	start := len(*s)
	spacing := [...]float64{12, 10.5, 13.5}[r.EyeSpacing]
	for i, x := range []float64{50 - spacing, 50 + spacing} {
		switch r.Eyes {
		case EyesDots:
			s.rect(x-3, 43, 6, 10, 3, p.Detail)
			s.circle(x-0.8, 45, .9, p.Face)
		case EyesPortholes:
			s.circle(x, 48, 6.2, p.Detail)
			s.circle(x, 48, 4.1, p.Face)
			s.circle(x+.5, 48, 2.5, p.Detail)
		case EyesSleepy:
			s.curve(p.Detail, 3.4, move(x-4, 48), quad(x, 51, x+4, 48))
		case EyesHappy:
			s.curve(p.Detail, 3.2, move(x-4, 50), quad(x, 41, x+4, 50))
		case EyesVisor:
			s.rect(x-6, 42, 12, 12, 4, p.Detail)
			s.rect(x-1.5, 45, 3, 6, 1.5, p.Face)
		case EyesButtons:
			s.circle(x, 48, 4, p.Detail)
			s.circle(x-1, 46.8, 1.2, p.Face)
		case EyesWink:
			if i == 0 {
				s.rect(x-3, 43, 6, 10, 3, p.Detail)
			} else {
				s.curve(p.Detail, 3, move(x-4, 48), quad(x, 44, x+4, 48))
			}
		case EyesCurious:
			s.circle(x, 48, 4+float64(i), p.Detail)
			s.circle(x-1, 46.5, 1.2, p.Face)
		}
	}
	(*s)[start:].transform(layoutTransform{sx: 1, sy: 1, dy: [...]float64{0, -1.5, 1.5}[r.EyeHeight]})
}

func (r Robot) drawMouth(s *scene) {
	p := r.Palette
	switch r.Mouth {
	case MouthSmile:
		s.curve(p.Detail, 2.6, move(44, 60), quad(50, 67, 56, 60))
	case MouthFlat:
		s.rect(45, 61, 10, 2.5, 1.25, p.Detail)
	case MouthOh:
		s.rect(47.5, 58.5, 5, 7, 2.5, p.Detail)
	case MouthGrille:
		s.rect(41, 58, 18, 7, 3.5, p.Detail)
		for _, x := range []float64{46, 50, 54} {
			s.rect(x-.7, 59.7, 1.4, 3.6, .7, p.Face)
		}
	case MouthCat:
		s.curve(p.Detail, 2.3, move(43, 60), cubic(43, 65, 49, 65, 50, 60), cubic(51, 65, 57, 65, 57, 60))
	case MouthGrin:
		s.filled(p.Detail, move(44, 58.5), line(56, 58.5), quad(58, 58.5, 57, 61), quad(50, 71, 43, 61), quad(42, 58.5, 44, 58.5), closePath())
		s.rect(46, 60, 8, 2, 1, p.Face)
	}
}

func (r Robot) drawAccessory(s *scene) {
	p := r.Palette
	switch r.Accessory {
	case AccessoryBadge:
		s.rect(57, 75, 8, 4, 2, p.Accent)
		s.circle(59.5, 77, .9, p.Face)
	case AccessoryButton:
		s.circle(61, 77, 2.5, p.Accent)
		s.circle(61, 77, 1, p.Face)
	case AccessoryHeart:
		s.filled(p.Accent, move(61, 75.5), cubic(57, 72, 55.5, 77, 61, 80), cubic(66.5, 77, 65, 72, 61, 75.5), closePath())
	}
}
