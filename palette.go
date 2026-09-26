package roboticon

import "image/color"

func rgb(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
}

// Background, shell, face, ink, accent. Ink stays dark on every light face.
var palettes = [...]Palette{
	{rgb(0xEAF3EC), rgb(0x85BBA3), rgb(0xF8F7E9), rgb(0x294D47), rgb(0xE99C83)}, // matcha
	{rgb(0xF9EEE6), rgb(0xECAA86), rgb(0xFFF8E9), rgb(0x624C48), rgb(0x88B9AA)}, // peach
	{rgb(0xEAF0F8), rgb(0x88ACD5), rgb(0xF4F7F9), rgb(0x334C6B), rgb(0xEEB875)}, // sky
	{rgb(0xF0EBF7), rgb(0xB19CCF), rgb(0xFAF5F3), rgb(0x514463), rgb(0xDEA0AA)}, // lilac
	{rgb(0xF8F1D9), rgb(0xE6C36F), rgb(0xFFFAEA), rgb(0x635438), rgb(0xC88676)}, // honey
	{rgb(0xF8EBEF), rgb(0xDCA0B4), rgb(0xFFF6EE), rgb(0x654653), rgb(0x94B6AC)}, // rose
	{rgb(0xE5F2F0), rgb(0x70B9BC), rgb(0xF2FAEF), rgb(0x2F5058), rgb(0xEBA582)}, // lagoon
	{rgb(0xF0EEE8), rgb(0xB6B9AE), rgb(0xFCFAEE), rgb(0x454E48), rgb(0xDD977B)}, // pebble
	{rgb(0xF7ECE5), rgb(0xCD8F75), rgb(0xFFF2DC), rgb(0x5D4038), rgb(0xAABD87)}, // terracotta
	{rgb(0xEBF1E1), rgb(0xA6BC7D), rgb(0xFCF9E9), rgb(0x49553A), rgb(0xD9989D)}, // pear
	{rgb(0xEAEFF4), rgb(0x8EA8B5), rgb(0xF4F7ED), rgb(0x364D59), rgb(0xE9BD6F)}, // mist
	{rgb(0xEEEAF3), rgb(0x999BC5), rgb(0xFFF4E8), rgb(0x454560), rgb(0xCBA2BE)}, // periwinkle
}
