package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
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
