//go:build js && wasm

// Command wasm exposes the real Go generator and renderers to a browser worker.
package main

import (
	"bytes"
	"encoding/base64"
	"syscall/js"

	"github.com/Southclaws/roboticon"
)

func main() {
	render := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 4 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeNumber || args[2].Type() != js.TypeString || args[3].Type() != js.TypeBoolean {
			return map[string]any{"error": "expected seed, size, format, transparent"}
		}
		r := roboticon.Generate(args[0].String())
		option := roboticon.GeneratedBackground
		if args[3].Bool() {
			option = roboticon.TransparentBackground
		}
		var b bytes.Buffer
		size, format := args[1].Int(), args[2].String()
		var err error
		switch format {
		case "svg":
			err = r.RenderSVG(&b, size, option)
		case "png":
			err = r.RenderPNG(&b, size, option)
		default:
			return map[string]any{"error": "unsupported format"}
		}
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		if format == "png" {
			return map[string]any{"data": base64.StdEncoding.EncodeToString(b.Bytes())}
		}
		return map[string]any{"data": b.String()}
	})
	defer render.Release()
	js.Global().Set("roboticonRender", render)
	select {}
}
