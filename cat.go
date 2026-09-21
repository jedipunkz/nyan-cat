package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/gif"
)

// nyan.cat が配布している本家アニメ GIF（272x168、12 フレーム）。
// 絵の著作権は Christopher Torres。手元で動かす用途のみ。
//
//go:embed nyancat.gif
var gifData []byte

// 本家 GIF は虹を含まないので、虹はこちらで描く。
// 位置はポップタルト胴体に合わせた実測値（frame 0 の生地の上端/下端/左端）。
const (
	bodyTop  = 8
	bodyBot  = 136
	bodyLeft = 74
	segW     = 16 // 虹のギザギザ 1 区画の幅
	amp      = 8  // ギザギザの振幅
)

var rainbowRGB = [][3]int{
	{255, 0, 0}, {255, 153, 0}, {255, 255, 0},
	{51, 255, 0}, {0, 153, 255}, {102, 51, 255},
}

var (
	frames      []*image.Paletted
	gifMap      [256]uint8 // GIF のパレット番号 -> canvas のパレット番号 (0 = 透明)
	palette     [][3]int
	rainbow0    int
	transparent bool // true なら背景（パレット 0）を描かず端末の背景を透かす
	catW, catH  int
)

func init() {
	g, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		panic(err)
	}
	frames = g.Image
	catW, catH = g.Config.Width, g.Config.Height

	palette = [][3]int{{12, 12, 30}} // 0 = 背景（宇宙）
	for i, c := range frames[0].Palette {
		r, gg, b, a := c.RGBA()
		if a == 0 {
			continue // 透明色は canvas の 0 のまま
		}
		gifMap[i] = uint8(len(palette))
		palette = append(palette, [3]int{int(r >> 8), int(gg >> 8), int(b >> 8)})
	}
	rainbow0 = len(palette)
	palette = append(palette, rainbowRGB...)
}

type canvas struct {
	w, h int
	px   []uint8
}

func newCanvas(w, h int) *canvas { return &canvas{w, h, make([]uint8, w*h)} }

func (c *canvas) at(x, y int) uint8 { return c.px[y*c.w+x] }

func (c *canvas) set(x, y int, v uint8) {
	if x >= 0 && x < c.w && y >= 0 && y < c.h {
		c.px[y*c.w+x] = v
	}
}

func (c *canvas) clear() {
	for i := range c.px {
		c.px[i] = 0
	}
}

// draw は猫の左端を catX に置いた 1 フレームを canvas に描く。
func draw(c *canvas, catX, frame int) {
	c.clear()

	// 虹: 画面左端から胴体の下まで。6 本を 1 区画ごとに上下させる。
	stripe := (bodyBot - bodyTop) / len(rainbowRGB)
	for x := 0; x < catX+bodyLeft && x < c.w; x++ {
		off := ((x/segW + frame) % 2) * amp
		for band := range rainbowRGB {
			for y := bodyTop + band*stripe; y < bodyTop+(band+1)*stripe; y++ {
				c.set(x, y+off, uint8(rainbow0+band))
			}
		}
	}

	// 猫（GIF の該当フレームをそのまま貼る）
	im := frames[frame%len(frames)]
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if v := gifMap[im.ColorIndexAt(x, y)]; v != 0 {
				c.set(catX+x, y, v)
			}
		}
	}
}
