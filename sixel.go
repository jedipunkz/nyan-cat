package main

import (
	"bufio"
	"fmt"
)

// rle は sixel の連長圧縮出力。同じ文字が続く間まとめて "!n<char>" にする。
type rle struct {
	out *bufio.Writer
	ch  byte
	n   int
}

func (r *rle) add(ch byte, n int) {
	if r.n > 0 && ch == r.ch {
		r.n += n
		return
	}
	r.flush()
	r.ch, r.n = ch, n
}

func (r *rle) flush() {
	if r.n == 0 {
		return
	}
	if r.n > 3 {
		fmt.Fprintf(r.out, "!%d%c", r.n, r.ch)
	} else {
		for i := 0; i < r.n; i++ {
			r.out.WriteByte(r.ch)
		}
	}
	r.n = 0
}

// writeSixel は canvas を scale 倍して sixel 画像として書き出す。
// 背景色 (index 0) も塗るので、前フレームの消去は不要。
func writeSixel(out *bufio.Writer, c *canvas, scale int) {
	w, h := c.w*scale, c.h*scale
	p2 := 0 // 0 = 背景も塗る / 1 = 0 ビットの画素はそのまま残す
	if transparent {
		p2 = 1
	}
	fmt.Fprintf(out, "\x1bP0;%d;0q\"1;1;%d;%d", p2, w, h)
	for i, p := range palette {
		// sixel の RGB は 0-100
		fmt.Fprintf(out, "#%d;2;%d;%d;%d", i, p[0]*100/255, p[1]*100/255, p[2]*100/255)
	}

	bits := make([][]byte, c.w) // アート列ごと: 色 -> 6 行分のビット
	for i := range bits {
		bits[i] = make([]byte, len(palette))
	}
	r := &rle{out: out}
	for band := 0; band*6 < h; band++ {
		for i := range bits {
			for j := range bits[i] {
				bits[i][j] = 0
			}
		}
		used := make([]bool, len(palette))
		for k := 0; k < 6; k++ {
			y := band*6 + k
			if y >= h {
				break
			}
			ay := y / scale
			for ax := 0; ax < c.w; ax++ {
				ci := c.at(ax, ay)
				bits[ax][ci] |= 1 << k
				used[ci] = true
			}
		}
		first := true
		for ci := range palette {
			if transparent && ci == 0 {
				continue // 背景は描かない
			}
			if !used[ci] {
				continue
			}
			if !first {
				out.WriteByte('$') // 同じ band の先頭へ戻る
			}
			first = false
			fmt.Fprintf(out, "#%d", ci)
			for ax := 0; ax < c.w; ax++ {
				r.add(byte(63+bits[ax][ci]), scale)
			}
			r.flush()
		}
		out.WriteByte('-') // 次の band へ
	}
	fmt.Fprint(out, "\x1b\\")
}
