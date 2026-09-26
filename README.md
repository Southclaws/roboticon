# Roboticon

Cute, deterministic robot avatars for Go.

![Twelve robot avatars](docs/examples.png)

**[Try the live playground →](https://southclaws.github.io/roboticon/)**

**1,061,683,200 possible trait combinations.** The original 230,400 combinations
(5 heads × 8 eyes × 6 mouths × 5 antenna choices × 4 ear choices × 4 accessories ×
12 palettes) now have 4 proportion presets × 4 face panels × 3 eye spacings ×
3 eye heights × 4 cheek styles × 8 casing markings. All geometry and colors are
curated; fine details become subtler at small sizes. Optional features include
“none”; accessories appear on about one third of robots.

![Curated robot variants, changing one trait at a time](docs/traits.png)

```go
import "github.com/Southclaws/roboticon"

avatar := roboticon.Generate("some-stable-id")
err := avatar.RenderPNG(w, 256)
err = avatar.RenderSVG(w, 256)

// The generated background is included by default.
err = avatar.RenderPNG(w, 256, roboticon.TransparentBackground)
err = avatar.RenderSVG(w, 256, roboticon.TransparentBackground)
```

Requires **Go 1.27+**. Pure Go. Render sizes: 1–2048px. Square, unmasked outputs
fit comfortably inside a circular avatar crop. Traits are independent of size;
SHA-256 streams keep generation deterministic. Distinct seeds can select the
same character.

The live playground runs the **real Go library through WebAssembly**, including
PNG and SVG downloads. Seeds stay in your browser.

```sh
go run ./cmd/gallery        # Offline gallery: gallery/index.html
sh scripts/build-web.sh    # Live playground: serve web/dist over HTTP
go test -race ./...
```
