// Command gallery writes an offline HTML gallery and actual PNG/SVG samples.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/Southclaws/roboticon"
)

type card struct{ Seed, File string }
type gallery struct {
	Cards  []card
	Prefix string
	Count  int
}

func main() {
	out := flag.String("out", "gallery", "output directory")
	count := flag.Int("count", 200, "number of robots (1–500)")
	prefix := flag.String("seed", "robot", "seed prefix; each avatar appends -NNN")
	flag.Parse()
	if err := run(*out, *prefix, *count); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Gallery: %s\nContact sheet: %s\n", filepath.Join(*out, "index.html"), filepath.Join(*out, "preview.png"))
}

func run(out, prefix string, count int) error {
	if count < 1 || count > 500 {
		return fmt.Errorf("count must be between 1 and 500")
	}
	if err := os.MkdirAll(filepath.Join(out, "images"), 0755); err != nil {
		return err
	}
	data := gallery{Prefix: prefix, Count: count}
	// First 24 seeds form a compact preview; all seeds appear in contact-sheet.png.
	preview := newSheet(6, min(count, 24), 144)
	contact := newSheet(10, count, 88)
	for i := 0; i < count; i++ {
		seed := fmt.Sprintf("%s-%03d", prefix, i)
		file := fmt.Sprintf("%03d", i)
		r := roboticon.Generate(seed)
		for _, size := range []int{24, 32, 64, 128} {
			var buf bytes.Buffer
			if err := r.RenderPNG(&buf, size); err != nil {
				return err
			}
			name := filepath.Join(out, "images", fmt.Sprintf("%s-%d.png", file, size))
			if err := os.WriteFile(name, buf.Bytes(), 0644); err != nil {
				return err
			}
			if size == 128 && i < 24 {
				if err := place(preview, buf.Bytes(), i, 6, 144, 8); err != nil {
					return err
				}
			}
			if size == 64 {
				if err := place(contact, buf.Bytes(), i, 10, 88, 12); err != nil {
					return err
				}
			}
		}
		var svg bytes.Buffer
		if err := r.RenderSVG(&svg, 128); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, "images", file+".svg"), svg.Bytes(), 0644); err != nil {
			return err
		}
		data.Cards = append(data.Cards, card{Seed: seed, File: file})
	}
	for name, im := range map[string]image.Image{"preview.png": preview, "contact-sheet.png": contact} {
		var buf bytes.Buffer
		if err := png.Encode(&buf, im); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, name), buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	var html bytes.Buffer
	if err := template.Must(template.New("gallery").Parse(page)).Execute(&html, data); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "index.html"), html.Bytes(), 0644)
}

func newSheet(cols, count, cell int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, cols*cell, ((count+cols-1)/cols)*cell))
	draw.Draw(im, im.Bounds(), image.NewUniform(color.RGBA{250, 248, 243, 255}), image.Point{}, draw.Src)
	return im
}

func place(dst *image.RGBA, data []byte, i, cols, cell, pad int) error {
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	p := image.Pt((i%cols)*cell+pad, (i/cols)*cell+pad)
	draw.Draw(dst, im.Bounds().Add(p), im, image.Point{}, draw.Src)
	return nil
}

const page = `<!doctype html>
<html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Roboticon · the little robot workshop</title>
<style>
:root{font-family:ui-sans-serif,system-ui,sans-serif;color:#394941;background:#faf8f3}*{box-sizing:border-box}body{margin:0;padding:48px 5vw}header{max-width:900px;margin-bottom:36px}.eyebrow{font-size:12px;text-transform:uppercase;letter-spacing:.16em;color:#758379}h1{font-size:clamp(32px,5vw,56px);font-weight:650;letter-spacing:-.05em;margin:10px 0}p{line-height:1.6;color:#748077}input{font:inherit;padding:12px 16px;border:1px solid #d9ded5;background:#fffefa;border-radius:12px;width:280px;max-width:100%;margin-top:12px}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(270px,1fr));gap:16px}article{background:#fffefa;border:1px solid #e8e8dd;border-radius:20px;padding:20px}article[hidden]{display:none}.faces{display:flex;align-items:end;justify-content:center;gap:16px;min-height:148px}figure{margin:0;text-align:center}img{display:block;border-radius:20%;max-width:none}figcaption{font:10px ui-monospace,monospace;color:#92998e;margin-top:8px}.tiny{display:flex;align-items:center;justify-content:center;gap:16px;margin:18px 0}footer{display:flex;align-items:center;justify-content:space-between;border-top:1px solid #edeee5;padding-top:14px}code{font-size:12px;color:#66736b;overflow-wrap:anywhere}a{font-size:11px;color:#607f70;text-decoration:none}.intro{max-width:580px}
.controls{display:flex;align-items:center;gap:20px;flex-wrap:wrap}.controls label{font-size:13px;display:flex;align-items:center;gap:8px}.controls input[type=checkbox]{width:16px;height:16px;margin:0;accent-color:#607f70}.circle-preview img{border-radius:50%}</style>
<header><div class="eyebrow">Roboticon / trait playground</div><h1>A few friendly faces.</h1><p class="intro">{{.Count}} little robots, grown from stable seeds. Every card shows real PNG renders at 128, 64, 32 and 24 pixels. Same personality, any size.</p><div class="controls"><input id="filter" type="search" placeholder="Find a seed…" aria-label="Filter by seed"><label><input id="circle" type="checkbox">Preview circle crop</label></div></header>
<main class="grid">{{range .Cards}}<article data-seed="{{.Seed}}"><div class="faces"><figure><img src="images/{{.File}}-128.png" width="128" height="128" alt="Robot {{.Seed}}"><figcaption>128 px</figcaption></figure><figure><img src="images/{{.File}}-64.png" width="64" height="64" alt=""><figcaption>64 px</figcaption></figure></div><div class="tiny"><figure><img src="images/{{.File}}-32.png" width="32" height="32" alt=""><figcaption>32 px</figcaption></figure><figure><img src="images/{{.File}}-24.png" width="24" height="24" alt=""><figcaption>24 px</figcaption></figure></div><footer><code>{{.Seed}}</code><a href="images/{{.File}}.svg">SVG ↗</a></footer></article>{{end}}</main>
<script>document.querySelector("#circle").addEventListener("change",e=>document.body.classList.toggle("circle-preview",e.target.checked));document.querySelector('#filter').addEventListener('input',e=>{const q=e.target.value.toLowerCase();document.querySelectorAll('article').forEach(c=>c.hidden=!c.dataset.seed.toLowerCase().includes(q));});</script></html>`
