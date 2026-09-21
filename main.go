// nyancat: 猫が左から右へ歩いて画面外へ消えるだけのターミナルアニメーション。
// 既定は sixel 画像。非対応端末では -ascii を使う。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

func termSize() (cols, rows, xpix, ypix int) {
	var ws struct{ Row, Col, X, Y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(),
		syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 80, 24, 0, 0 // 非 tty (パイプ等) の既定値
	}
	return int(ws.Col), int(ws.Row), int(ws.X), int(ws.Y)
}

func restore() {
	fmt.Print("\x1b[0m\x1b[?25h\x1b[2J\x1b[H")
}

func main() {
	ascii := flag.Bool("ascii", false, "画像ではなく ASCII で描く")
	proto := flag.String("proto", "kitty", "画像プロトコル: sixel / iterm / kitty")
	still := flag.Bool("still", false, "1 フレームだけ出して終了する（表示確認用）")
	bg := flag.Bool("bg", false, "背景を宇宙色で塗る（既定は透明）")
	scale := flag.Int("scale", 1, "アートピクセルの拡大率")
	pngPath := flag.String("png", "", "1 フレームを PNG に書き出して終了する（確認用）")
	flag.Parse()
	transparent = !*bg

	cols, rows, xpix, ypix := termSize()
	termW, termH = cols, rows

	if *pngPath != "" {
		dumpPNG(*pngPath, *scale)
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() { <-sig; restore(); os.Exit(0) }()
	defer restore()

	if *ascii {
		runASCII()
		return
	}
	run(*proto, cols, rows, xpix, ypix, *scale, *still)
}

// writers は画像 1 枚を端末に流し込む方式。端末が対応しているものを選ぶ。
var writers = map[string]func(*bufio.Writer, *canvas, int){
	"sixel": writeSixel,
	"iterm": writeITerm,
	"kitty": writeKitty,
}

func run(proto string, cols, rows, xpix, ypix, scale int, still bool) {
	write, ok := writers[proto]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown -proto %q (sixel / iterm / kitty)\n", proto)
		os.Exit(2)
	}
	cellW, cellH := 8, 16
	if xpix > 0 && ypix > 0 {
		cellW, cellH = xpix/cols, ypix/rows
	}
	canvasW := cols * cellW / scale
	c := newCanvas(canvasW, catH)

	imgRows := (c.h*scale + cellH - 1) / cellH
	top := (rows-imgRows)/2 + 1
	if top < 1 {
		top = 1
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	if still {
		draw(c, canvasW/3, 0)
		write(out, c, scale)
		fmt.Fprintln(out)
		return
	}
	fmt.Fprint(out, "\x1b[?25l\x1b[2J")
	for x, tick := -catW, 0; x <= canvasW; x, tick = x+12, tick+1 {
		draw(c, x, tick/2) // GIF は 70ms ごと = 2 tick に 1 コマ
		if transparent {
			fmt.Fprint(out, "\x1b[2J") // 背景で塗り潰さないので前フレームを消す
		}
		fmt.Fprintf(out, "\x1b[%d;1H", top)
		write(out, c, scale)
		out.Flush()
		time.Sleep(35 * time.Millisecond)
	}
}

// dumpPNG は端末に出せない環境で絵を確認するための書き出し。
func dumpPNG(path string, scale int) {
	c := newCanvas(catW*2, catH)
	draw(c, c.w/2, 0)
	if err := os.WriteFile(path, framePNG(c, scale), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
