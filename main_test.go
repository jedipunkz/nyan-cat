package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestArtIsRectangular(t *testing.T) {
	w := len([]rune(art[0]))
	for i, l := range art {
		if got := len([]rune(l)); got != w {
			t.Fatalf("art[%d] width = %d, want %d", i, got, w)
		}
	}
}

func TestZigAlternates(t *testing.T) {
	x := 20
	for c := x - 2; c >= 0; c -= 2 {
		if zig(x, c) == zig(x, c-2) {
			t.Fatalf("zig not alternating at c=%d", c)
		}
	}
	if zig(20, 10) == zig(21, 10) {
		t.Fatal("zig must shift as the cat advances")
	}
}

func TestGIFLoads(t *testing.T) {
	if len(frames) != 12 || catW != 272 || catH != 168 {
		t.Fatalf("frames=%d size=%dx%d", len(frames), catW, catH)
	}
	for i, f := range frames { // 全フレームが同じパレットでないと gifMap が壊れる
		for j, c := range f.Palette {
			if c != frames[0].Palette[j] {
				t.Fatalf("frame %d palette differs at %d", i, j)
			}
		}
	}
	if rainbow0 != len(palette)-len(rainbowRGB) {
		t.Fatalf("rainbow0=%d palette=%d", rainbow0, len(palette))
	}
}

func TestSixelEncodesPixel(t *testing.T) {
	c := newCanvas(1, 6)
	c.set(0, 0, 1) // 最上段に輪郭色を 1 px
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	writeSixel(out, c, 1)
	out.Flush()
	s := buf.String()
	if !strings.HasPrefix(s, "\x1bP0;0;0q\"1;1;1;6") || !strings.HasSuffix(s, "\x1b\\") {
		t.Fatalf("DCS wrapper broken: %q", s)
	}
	if !strings.Contains(s, "#1@") { // 色 1、bit0 のみ = 63+1 = '@'
		t.Fatalf("pixel not encoded: %q", s)
	}
}

func TestTransparentBackground(t *testing.T) {
	transparent = true
	defer func() { transparent = false }()
	c := newCanvas(catW, catH)
	draw(c, 0, 0)
	img := frameImage(c, 1)
	if _, _, _, a := img.At(1, 160).RGBA(); a != 0 { // 猫も虹も無い隅
		t.Fatalf("背景が透明でない: alpha=%d", a)
	}
	if _, _, _, a := img.At(120, 60).RGBA(); a == 0 { // ポップタルトの中
		t.Fatal("猫まで透明になっている")
	}
	if _, _, _, a := img.At(10, 60).RGBA(); a == 0 { // 虹の中
		t.Fatal("虹まで透明になっている")
	}
}

func TestReverseMirrorsCat(t *testing.T) {
	const y = 150 // 脚の行。虹 (bodyBot+amp まで) が届かないので猫だけを比べられる
	a := newCanvas(catW, catH)
	draw(a, 0, 0)
	reverse = true
	defer func() { reverse = false }()
	b := newCanvas(catW, catH)
	draw(b, 0, 0)
	for x := 0; x < catW; x++ {
		if a.at(x, y) != b.at(catW-1-x, y) {
			t.Fatalf("x=%d で鏡像になっていない", x)
		}
	}
	mirrored := true
	for x := 0; x < catW; x++ { // 元の行が左右対称だと上の検証が自明に通ってしまう
		if a.at(x, y) != a.at(catW-1-x, y) {
			mirrored = false
			break
		}
	}
	if mirrored {
		t.Fatal("比較行が左右対称で検証にならない")
	}
}
