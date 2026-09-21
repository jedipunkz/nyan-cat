package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// ポップタルト胴体 + 猫の顔。全行同じ幅にそろえてある。
var art = []string{
	" ,--------.        ",
	"|~~~~~~~~~| /\\_/\\  ",
	"|~~~~~~~~~|( o.o ) ",
	"|~~~~~~~~~| > ^ <  ",
	" `--------'        ",
	"  ^^    ^^         ",
}

var rainbow = []int{196, 208, 226, 46, 21, 93} // 赤橙黄緑青紫 (256 色)

var termW, termH = 80, 24

func catColor(r rune) int {
	switch r {
	case '~':
		return 218 // ポップタルトのピンク
	case '|', ',', '.', '`', '-', '\'':
		return 223 // 生地の縁
	case '^':
		return 217 // 足
	}
	return 255 // 顔
}

// zig は虹の 1 セグメント (2 桁幅) の縦ずらし量 0/1 を返す。x が進むと縞が流れる。
func zig(x, c int) int { return ((x-c)/2 + x) % 2 }

func put(out *bufio.Writer, row, col int, s string, color func(rune) int) {
	for _, r := range s {
		if r != ' ' && col >= 1 && col <= termW && row >= 1 && row <= termH {
			fmt.Fprintf(out, "\x1b[%d;%dH\x1b[38;5;%dm%c", row, col, color(r), r)
		}
		col++
	}
	fmt.Fprint(out, "\x1b[0m")
}

// runASCII は 256 色 ANSI で猫を歩かせる（sixel 非対応端末向け）。
func runASCII() {
	catW := len([]rune(art[0]))
	top := termH/2 - len(art)/2
	if top < 1 {
		top = 1
	}
	out := bufio.NewWriter(os.Stdout)
	fmt.Fprint(out, "\x1b[?25l")
	for x := -catW; x <= termW; x++ {
		fmt.Fprint(out, "\x1b[2J")
		for c := x - 2; c >= 0; c -= 2 {
			off := zig(x, c)
			for i, col := range rainbow {
				put(out, top+i+off, c+1, "██", func(rune) int { return col })
			}
		}
		for i, line := range art {
			col := x + 1
			if i == len(art)-1 {
				col += x % 2
			}
			put(out, top+i, col, line, catColor)
		}
		out.Flush()
		time.Sleep(50 * time.Millisecond)
	}
}
