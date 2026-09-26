package roboticon

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

// RenderOption controls the canvas without changing the robot's selected traits.
type RenderOption uint8

const (
	// GeneratedBackground includes the curated background (the default).
	GeneratedBackground RenderOption = iota
	// TransparentBackground omits the background and grounding shadow.
	TransparentBackground
)

func background(options []RenderOption) (bool, error) {
	show := true
	for _, option := range options {
		switch option {
		case GeneratedBackground:
			show = true
		case TransparentBackground:
			show = false
		default:
			return false, fmt.Errorf("roboticon: invalid render option %d", option)
		}
	}
	return show, nil
}

// RenderPNG writes an antialiased PNG. Size is the width and height in pixels
// and must be between 1 and 2048. Small avatars are rendered at 4x resolution.
// Output is deterministic for a fixed library, dependency, and Go toolchain version.
func (r Robot) RenderPNG(w io.Writer, size int, options ...RenderOption) error {
	showBackground, err := background(options)
	if err != nil {
		return err
	}
	if err := r.validate(size); err != nil {
		return err
	}
	factor := 1
	if size <= 256 {
		factor = 4
	} else if size <= 1024 {
		factor = 2
	}
	dc := gg.NewContext(size*factor, size*factor)
	scale := float64(size*factor) / 100
	dc.SetLineCapRound()
	dc.SetLineJoinRound()
	for _, sh := range r.scene(showBackground) {
		drawShape(dc, sh, scale)
	}
	var output image.Image = dc.Image()
	if factor > 1 {
		resized := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(resized, resized.Bounds(), output, output.Bounds(), draw.Src, nil)
		output = resized
	}
	return png.Encode(w, output)
}

func drawShape(dc *gg.Context, s shape, k float64) {
	switch s.kind {
	case 'r':
		dc.DrawRoundedRectangle(s.x*k, s.y*k, s.w*k, s.h*k, s.radius*k)
	case 'c':
		dc.DrawCircle(s.x*k, s.y*k, s.radius*k)
	case 'p':
		for _, c := range s.path {
			v := c.v
			switch c.op {
			case 'M':
				dc.MoveTo(v[0]*k, v[1]*k)
			case 'L':
				dc.LineTo(v[0]*k, v[1]*k)
			case 'Q':
				dc.QuadraticTo(v[0]*k, v[1]*k, v[2]*k, v[3]*k)
			case 'C':
				dc.CubicTo(v[0]*k, v[1]*k, v[2]*k, v[3]*k, v[4]*k, v[5]*k)
			case 'Z':
				dc.ClosePath()
			}
		}
	}
	if s.width > 0 {
		dc.SetColor(s.stroke)
		dc.SetLineWidth(s.width * k)
		dc.Stroke()
	} else {
		dc.SetColor(s.fill)
		dc.Fill()
	}
}

// RenderSVG writes standalone vector SVG using the same geometry as RenderPNG.
// Size sets the document's pixel dimensions; the viewBox is always 0 0 100 100.
func (r Robot) RenderSVG(w io.Writer, size int, options ...RenderOption) error {
	showBackground, err := background(options)
	if err != nil {
		return err
	}
	if err := r.validate(size); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 100 100" fill="none" stroke-linecap="round" stroke-linejoin="round">`, size, size)
	b.WriteString("\n")
	for _, s := range r.scene(showBackground) {
		writeShape(&b, s)
	}
	b.WriteString("</svg>\n")
	n, err := io.WriteString(w, b.String())
	if err == nil && n != b.Len() {
		return io.ErrShortWrite
	}
	return err
}

func writeShape(b *strings.Builder, s shape) {
	switch s.kind {
	case 'r':
		fmt.Fprintf(b, `<rect x="%g" y="%g" width="%g" height="%g" rx="%g"`, s.x, s.y, s.w, s.h, s.radius)
	case 'c':
		fmt.Fprintf(b, `<circle cx="%g" cy="%g" r="%g"`, s.x, s.y, s.radius)
	case 'p':
		b.WriteString(`<path d="`)
		for _, c := range s.path {
			b.WriteByte(c.op)
			for i, v := range c.v {
				if i > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			}
		}
		b.WriteByte('"')
	}
	if s.width > 0 {
		writeColor(b, "stroke", s.stroke)
		fmt.Fprintf(b, ` stroke-width="%g"`, s.width)
	} else {
		writeColor(b, "fill", s.fill)
	}
	b.WriteString("/>\n")
}

func writeColor(b *strings.Builder, attr string, c color.RGBA) {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	fmt.Fprintf(b, ` %s="#%02x%02x%02x"`, attr, n.R, n.G, n.B)
	if n.A < 255 {
		fmt.Fprintf(b, ` %s-opacity="%s"`, attr, strconv.FormatFloat(float64(n.A)/255, 'f', -1, 64))
	}
}

func (r Robot) validate(size int) error {
	if size < 1 || size > 2048 {
		return fmt.Errorf("roboticon: size must be between 1 and 2048, got %d", size)
	}
	if r.Head > HeadSquircle || r.Eyes > EyesCurious || r.Mouth > MouthGrin ||
		r.Antenna > AntennaSprout || r.Ears > EarsNubs || r.Accessory > AccessoryHeart ||
		r.Proportions > ProportionsBigHead || r.FacePanel > FacePanelWide || r.EyeSpacing > EyeSpacingWide ||
		r.EyeHeight > EyeHeightLow || r.Cheeks > CheeksLines || r.Marking > MarkingTempleDots {
		return fmt.Errorf("roboticon: invalid trait value")
	}
	return nil
}
