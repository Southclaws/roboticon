// Package roboticon generates friendly, deterministic robot avatars.
package roboticon

import (
	"crypto/sha256"
	"encoding/binary"
	"image/color"
)

// Robot is a size-independent description of an avatar. It is comparable and
// contains no hidden random state. Render methods do not modify it.
type Robot struct {
	Head      Head
	Eyes      Eyes
	Mouth     Mouth
	Antenna   Antenna
	Ears      Ears
	Accessory Accessory
	Palette   Palette
}

type Head uint8

const (
	HeadRounded Head = iota
	HeadPill
	HeadTall
	HeadDome
	HeadSquircle
)

type Eyes uint8

const (
	EyesDots Eyes = iota
	EyesPortholes
	EyesSleepy
	EyesHappy
	EyesVisor
	EyesButtons
	EyesWink
	EyesCurious
)

type Mouth uint8

const (
	MouthSmile Mouth = iota
	MouthFlat
	MouthOh
	MouthGrille
	MouthCat
	MouthGrin
)

type Antenna uint8

const (
	AntennaNone Antenna = iota
	AntennaBobble
	AntennaTwin
	AntennaLoop
	AntennaSprout
)

type Ears uint8

const (
	EarsNone Ears = iota
	EarsPads
	EarsDiscs
	EarsNubs
)

type Accessory uint8

const (
	AccessoryNone Accessory = iota
	AccessoryBlush
	AccessoryFreckles
	AccessoryBadge
)

// Palette uses concrete RGBA values so Robots can be compared with ==. All
// generated colors are opaque; callers can also supply their own palette.
type Palette struct {
	Background color.RGBA
	Body       color.RGBA
	Face       color.RGBA
	Detail     color.RGBA
	Accent     color.RGBA
}

// Generate selects traits from SHA-256 entropy. Named streams keep existing
// traits stable when another trait is added or generation order changes.
// Empty strings and arbitrary UTF-8 or binary strings are valid seeds.
func Generate(seed string) Robot {
	root := sha256.Sum256([]byte(seed))
	pick := func(label string, n uint64) uint8 {
		e := newEntropy(root, label)
		return uint8(e.below(n))
	}
	accessory := AccessoryNone
	if pick("accessory-presence", 3) == 0 {
		accessory = Accessory(1 + pick("accessory", 3))
	}
	return Robot{
		Head: Head(pick("head", 5)), Eyes: Eyes(pick("eyes", 8)),
		Mouth: Mouth(pick("mouth", 6)), Antenna: Antenna(pick("antenna", 5)),
		Ears: Ears(pick("ears", 4)), Accessory: accessory,
		Palette: palettes[pick("palette", uint64(len(palettes)))],
	}
}

// entropy expands a domain-separated seed into a counter-mode SHA-256 stream.
// It is deliberately independent of math/rand's implementation and global state.
type entropy struct {
	key     [32]byte
	counter uint64
}

func newEntropy(root [32]byte, label string) entropy {
	h := sha256.New()
	h.Write(root[:])
	h.Write([]byte(label))
	var e entropy
	copy(e.key[:], h.Sum(nil))
	return e
}

func (e *entropy) next() uint64 {
	var input [40]byte
	copy(input[:32], e.key[:])
	binary.BigEndian.PutUint64(input[32:], e.counter)
	e.counter++
	block := sha256.Sum256(input[:])
	return binary.BigEndian.Uint64(block[:8])
}

func (e *entropy) below(n uint64) uint64 {
	// Reject the short remainder at the bottom of the uint64 range.
	threshold := -n % n
	for {
		v := e.next()
		if v >= threshold {
			return v % n
		}
	}
}
