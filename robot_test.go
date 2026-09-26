package roboticon

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	// Independently calculated with Python hashlib and big-endian integer decoding.
	cases := []struct {
		seed string
		want Robot
	}{
		{"robot-a", Robot{HeadDome, EyesButtons, MouthSmile, AntennaBobble, EarsNubs, AccessoryNone,
			Palette{color.RGBA{248, 235, 239, 255}, color.RGBA{220, 160, 180, 255}, color.RGBA{255, 246, 238, 255}, color.RGBA{101, 70, 83, 255}, color.RGBA{148, 182, 172, 255}}}},
		{"robot-b", Robot{HeadSquircle, EyesSleepy, MouthSmile, AntennaSprout, EarsNubs, AccessoryNone,
			Palette{color.RGBA{229, 242, 240, 255}, color.RGBA{112, 185, 188, 255}, color.RGBA{242, 250, 239, 255}, color.RGBA{47, 80, 88, 255}, color.RGBA{235, 165, 130, 255}}}},
		{"", Robot{HeadDome, EyesPortholes, MouthGrin, AntennaBobble, EarsNone, AccessoryFreckles,
			Palette{color.RGBA{234, 239, 244, 255}, color.RGBA{142, 168, 181, 255}, color.RGBA{244, 247, 237, 255}, color.RGBA{54, 77, 89, 255}, color.RGBA{233, 189, 111, 255}}}},
	}
	for _, tt := range cases {
		t.Run(tt.seed, func(t *testing.T) {
			if got := Generate(tt.seed); got != tt.want {
				t.Fatalf("got %+v; want %+v", got, tt.want)
			}
			if Generate(tt.seed) != Generate(tt.seed) {
				t.Fatal("generation is not deterministic")
			}
		})
	}
	if Generate("robot-a") == Generate("robot-b") {
		t.Fatal("different example seeds selected the same robot")
	}
	for _, seed := range []string{"🤖 café", "\x00\xff", strings.Repeat("long", 10000)} {
		if Generate(seed) != Generate(seed) {
			t.Fatalf("unstable seed %q", seed)
		}
	}
}

func TestRenderDeterminismAndSizes(t *testing.T) {
	for _, size := range []int{1, 24, 32, 64, 128, 256, 257, 1025} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			t.Parallel()
			r := Generate("robot-a")
			before := r
			for _, render := range []func(io.Writer, int, ...RenderOption) error{r.RenderPNG, r.RenderSVG} {
				var a, b bytes.Buffer
				if err := render(&a, size); err != nil {
					t.Fatal(err)
				}
				if err := render(&b, size); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(a.Bytes(), b.Bytes()) {
					t.Fatal("output changed on repeated rendering")
				}
				if bytes.HasPrefix(a.Bytes(), []byte("<svg")) {
					var root struct {
						XMLName xml.Name
						Width   int    `xml:"width,attr"`
						Height  int    `xml:"height,attr"`
						ViewBox string `xml:"viewBox,attr"`
					}
					if err := xml.Unmarshal(a.Bytes(), &root); err != nil {
						t.Fatal(err)
					}
					if root.Width != size || root.Height != size || root.ViewBox != "0 0 100 100" {
						t.Fatalf("bad SVG dimensions: %+v", root)
					}
				} else {
					im, err := png.Decode(&a)
					if err != nil {
						t.Fatal(err)
					}
					if im.Bounds().Dx() != size || im.Bounds().Dy() != size {
						t.Fatal("wrong PNG dimensions")
					}
				}
			}
			if r != before {
				t.Fatal("rendering mutated traits")
			}
		})
	}
}

func TestSVGSizeOnlyChangesDocumentDimensions(t *testing.T) {
	r := Generate("robot-b")
	var small, large bytes.Buffer
	if err := r.RenderSVG(&small, 24); err != nil {
		t.Fatal(err)
	}
	if err := r.RenderSVG(&large, 256); err != nil {
		t.Fatal(err)
	}
	normalized := strings.Replace(large.String(), `width="256" height="256"`, `width="24" height="24"`, 1)
	if small.String() != normalized {
		t.Fatal("size changed the SVG geometry")
	}
	if strings.Contains(small.String(), "<image") {
		t.Fatal("SVG embeds raster content")
	}
}

func TestGolden(t *testing.T) {
	for _, seed := range []string{"robot-000", "robot-011"} {
		for _, format := range []string{"png", "svg"} {
			t.Run(seed+"."+format, func(t *testing.T) {
				var got bytes.Buffer
				r := Generate(seed)
				render := r.RenderPNG
				if format == "svg" {
					render = r.RenderSVG
				}
				if err := render(&got, 64); err != nil {
					t.Fatal(err)
				}
				want, err := os.ReadFile("testdata/" + seed + "." + format)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got.Bytes(), want) {
					t.Fatal("appearance changed; inspect the gallery before intentionally updating golden files")
				}
			})
		}
	}
}

func TestGalleryPopulation(t *testing.T) {
	var heads [5]bool
	var eyes [8]bool
	var mouths [6]bool
	var antennae [5]bool
	var ears [4]bool
	var accessories [4]bool
	seenPalettes := map[Palette]bool{}
	seenRobots := map[Robot]bool{}
	for i := 0; i < 200; i++ {
		r := Generate(fmt.Sprintf("robot-%03d", i))
		heads[r.Head] = true
		eyes[r.Eyes] = true
		mouths[r.Mouth] = true
		antennae[r.Antenna] = true
		ears[r.Ears] = true
		accessories[r.Accessory] = true
		seenPalettes[r.Palette] = true
		seenRobots[r] = true
		var b bytes.Buffer
		if err := r.RenderPNG(&b, 32); err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(&b)
		if err != nil {
			t.Fatal(err)
		}
		// Every generated character leaves the border entirely clear.
		for x := 0; x < 32; x++ {
			for _, pt := range [][2]int{{x, 0}, {x, 31}, {0, x}, {31, x}} {
				if im.At(pt[0], pt[1]) != r.Palette.Background {
					t.Fatalf("seed %d clips the canvas", i)
				}
			}
		}
		// A central eye pixel must differ visibly from the background.
		if im.At(12, 15) == r.Palette.Background {
			t.Fatalf("seed %d is blank", i)
		}
	}
	for _, set := range [][]bool{heads[:], eyes[:], mouths[:], antennae[:], ears[:], accessories[:]} {
		for i, seen := range set {
			if !seen {
				t.Fatalf("trait %d not represented in gallery", i)
			}
		}
	}
	if len(seenPalettes) != 12 {
		t.Fatalf("only %d palettes represented", len(seenPalettes))
	}
	if len(seenRobots) < 195 {
		t.Fatalf("too many duplicate robots: %d unique", len(seenRobots))
	}
}

func TestPaletteContrast(t *testing.T) {
	// Relative sRGB luminance, independently checking the curated face/ink pairs.
	luminance := func(c color.RGBA) float64 {
		channel := func(v uint8) float64 {
			x := float64(v) / 255
			if x <= .04045 {
				return x / 12.92
			}
			return math.Pow((x+.055)/1.055, 2.4)
		}
		return .2126*channel(c.R) + .7152*channel(c.G) + .0722*channel(c.B)
	}
	for i, p := range palettes {
		contrast := (luminance(p.Face) + .05) / (luminance(p.Detail) + .05)
		if contrast < 4.5 {
			t.Fatalf("palette %d contrast %.2f is too low", i, contrast)
		}
	}
}

func TestCircularSafeArea(t *testing.T) {
	// Exercise every silhouette combination. A radius of 47 leaves three logical
	// units of clearance inside a normal radius-50 circular avatar crop.
	for head := HeadRounded; head <= HeadSquircle; head++ {
		for antenna := AntennaNone; antenna <= AntennaSprout; antenna++ {
			for ears := EarsNone; ears <= EarsNubs; ears++ {
				r := Generate("circle-safe")
				r.Head, r.Antenna, r.Ears = head, antenna, ears
				var b bytes.Buffer
				if err := r.RenderPNG(&b, 128); err != nil {
					t.Fatal(err)
				}
				im, err := png.Decode(&b)
				if err != nil {
					t.Fatal(err)
				}
				for y := 0; y < 128; y++ {
					for x := 0; x < 128; x++ {
						dx, dy := (float64(x)+.5)*100/128-50, (float64(y)+.5)*100/128-50
						if dx*dx+dy*dy >= 47*47 && im.At(x, y) != r.Palette.Background {
							t.Fatalf("head=%d antenna=%d ears=%d exceeds circular safe area at (%d,%d)", head, antenna, ears, x, y)
						}
					}
				}
			}
		}
	}
}

var errWrite = errors.New("writer failed")

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errWrite }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

func TestRenderErrors(t *testing.T) {
	r := Generate("robot-a")
	for _, render := range []func(io.Writer, int, ...RenderOption) error{r.RenderPNG, r.RenderSVG} {
		for _, size := range []int{-1, 0, 2049, int(^uint(0) >> 1)} {
			var b bytes.Buffer
			if err := render(&b, size); err == nil || b.Len() != 0 {
				t.Fatal("invalid size must fail before writing")
			}
		}
		if err := render(brokenWriter{}, 32); !errors.Is(err, errWrite) {
			t.Fatalf("writer error lost: %v", err)
		}
	}
	if err := r.RenderSVG(shortWriter{}, 32); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write error lost: %v", err)
	}
	r.Eyes = 255
	if err := r.RenderPNG(io.Discard, 32); err == nil {
		t.Fatal("invalid eyes accepted")
	}
	if err := r.RenderSVG(io.Discard, 32); err == nil {
		t.Fatal("invalid eyes accepted")
	}
}

func ExampleGenerate() {
	avatar := Generate("some-stable-id")
	var png bytes.Buffer
	if err := avatar.RenderPNG(&png, 256); err != nil {
		panic(err)
	}
	fmt.Println(png.Len() > 0)
	// Output: true
}
