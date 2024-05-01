/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"fmt"
	"image/color"
	"os"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

//=====================绘图========================

// 只支持 6 位或 8 位的颜色转换
func Hex2Clr(hex string) color.Color {
	if len(hex) != 6 && len(hex) != 8 {
		panic(fmt.Sprintf("err hex %s", hex))
	}

	data := []uint8{255, 255, 255, 255}
	for i := 0; i < len(hex); i += 2 {
		temp, err := strconv.ParseInt(hex[i:i+2], 16, 64)
		HandleErr(err)
		data[i/2] = uint8(temp)
	}
	return color.RGBA{R: data[0], G: data[1], B: data[2], A: data[3]}
}

func FillRect(screen *ebiten.Image, x, y, w, h float32, clr color.Color) {
	vector.DrawFilledRect(screen, x, y, w, h, clr, false)
}

func StoryRect(screen *ebiten.Image, x, y, w, h, sw float32, clr color.Color) {
	vector.StrokeRect(screen, x, y, w, h, sw, clr, false)
}

var defaultFont *opentype.Font

func NewFont(size float64) font.Face {
	if defaultFont == nil {
		// 必须使用支持中文的字体
		bs, err := os.ReadFile("res/ipix.ttf")
		HandleErr(err)
		defaultFont, err = opentype.Parse(bs)
		HandleErr(err)
	}

	// 设置字体样式  dpi不清楚什么意思，这里固定为96
	face, err := opentype.NewFace(defaultFont, &opentype.FaceOptions{Size: size, DPI: 96, Hinting: font.HintingFull})
	HandleErr(err)
	return face
}

func DrawText(screen *ebiten.Image, str string, x, y float32, anchor Anchor, face font.Face, clr color.Color) {
	bound := text.BoundString(face, str)
	// 根据测量的结果，先把整体左上角移动到 x,y 处，再根据锚点与测量的宽高进行偏移
	x = x - float32(bound.Min.X) + float32(bound.Dx())*real(anchor)
	y = y - float32(bound.Min.Y) + float32(bound.Dy())*imag(anchor)
	text.Draw(screen, str, face, int(x), int(y), clr)
}
