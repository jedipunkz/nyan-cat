package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	draw2 "image/draw"
	"image/png"
	"math/rand/v2"
	"os"
)

func frameImage(c *canvas, scale int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.w*scale, c.h*scale))
	for y := 0; y < c.h*scale; y++ {
		for x := 0; x < c.w*scale; x++ {
			i := c.at(x/scale, y/scale)
			if i == 0 && transparent {
				continue // 0 = 背景。透明モードでは何も置かない
			}
			p := palette[i]
			img.Set(x, y, color.RGBA{uint8(p[0]), uint8(p[1]), uint8(p[2]), 255})
		}
	}
	return img
}

// 毎フレーム encode するので圧縮より速度を優先する。
var pngEnc = png.Encoder{CompressionLevel: png.BestSpeed}

func framePNG(c *canvas, scale int) []byte {
	var buf bytes.Buffer
	pngEnc.Encode(&buf, frameImage(c, scale))
	return buf.Bytes()
}

// writeITerm は iTerm2 inline image protocol（WezTerm / iTerm2 / VS Code 等）。
func writeITerm(out *bufio.Writer, c *canvas, scale int) {
	b := framePNG(c, scale)
	fmt.Fprintf(out, "\x1b]1337;File=inline=1;size=%d;doNotMoveCursor=1:%s\a",
		len(b), base64.StdEncoding.EncodeToString(b))
}

// writeKitty は kitty graphics protocol（kitty / WezTerm / ghostty 等）。
// 4096 バイトごとに分割送信し、毎フレーム前の画像を消す。
func writeKitty(out *bufio.Writer, c *canvas, scale int) {
	fmt.Fprint(out, "\x1b_Ga=d,q=2\x1b\\") // 前フレームを削除
	b64 := base64.StdEncoding.EncodeToString(framePNG(c, scale))
	first := true
	for len(b64) > 0 {
		n := min(4096, len(b64))
		chunk := b64[:n]
		b64 = b64[n:]
		more := 0
		if len(b64) > 0 {
			more = 1
		}
		if first {
			fmt.Fprintf(out, "\x1b_Gf=100,a=T,q=2,m=%d;%s\x1b\\", more, chunk)
			first = false
			continue
		}
		fmt.Fprintf(out, "\x1b_Gm=%d;%s\x1b\\", more, chunk)
	}
}

// writeSocial は GitHub の social preview 用に 1280x640 の 1 枚絵を書き出す。
// 猫と虹は透明背景で描き、星空の上に重ねる。
func writeSocial(path string) error {
	const w, h, scale = 1280, 640, 2

	c := newCanvas(w/scale, catH)
	transparent = true
	draw(c, c.w-catW-40, 0) // 猫を右に寄せ、虹を左端まで伸ばす
	cat := frameImage(c, scale)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := palette[0]
	space := color.RGBA{uint8(bg[0]), uint8(bg[1]), uint8(bg[2]), 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, space)
		}
	}
	// 星。毎回同じ絵になるよう種を固定する。
	rnd := rand.New(rand.NewPCG(4, 2))
	for i := 0; i < 120; i++ {
		x, y := rnd.IntN(w), rnd.IntN(h)
		for _, d := range [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			img.Set(x+d[0], y+d[1], color.RGBA{255, 255, 255, 255})
		}
	}
	b := cat.Bounds()
	at := image.Rect((w-b.Dx())/2, (h-b.Dy())/2, (w-b.Dx())/2+b.Dx(), (h-b.Dy())/2+b.Dy())
	draw2.Draw(img, at, cat, image.Point{}, draw2.Over)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
