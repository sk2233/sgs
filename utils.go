/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

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

func RandSlice(arr any) {
	sort.Slice(arr, func(_, _ int) bool {
		return rand.Intn(2) == 0
	})
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
	// Anchor 是一个复数类型 实数当x锚点，虚数当y锚点，取值都在 0.0～1.0
	x = x - float32(bound.Min.X) - float32(bound.Dx())*real(anchor)
	y = y - float32(bound.Min.Y) - float32(bound.Dy())*imag(anchor)
	text.Draw(screen, str, face, int(x), int(y), clr)
}

func FillCircle(screen *ebiten.Image, x, y, r float32, clr color.Color) {
	vector.DrawFilledCircle(screen, x, y, r, clr, false)
}

func StrokeCircle(screen *ebiten.Image, x, y, r, sw float32, clr color.Color) {
	vector.StrokeCircle(screen, x, y, r, sw, clr, false)
}

// 卡牌：宽 110 高 160
func DrawCard(screen *ebiten.Image, x, y float32, card *Card) {
	FillRect(screen, x, y, 110, 160, ClrDECDBA)
	StoryRect(screen, x, y, 110, 160, 2, Clr000000)
	pointAndSuit := fmt.Sprintf("%s\n%s", card.Suit, card.Point)
	suitClr := GetSuitClr(card.Suit)
	DrawText(screen, pointAndSuit, x+10, y+10, AnchorTopLeft, Font18, suitClr)
	DrawText(screen, card.Name, x+55, y+80, AnchorMidCenter, Font16, Clr000000)
	if card.Type == CardKit {
		DrawText(screen, string(card.KitType), x+55, y+160-10, AnchorBtmCenter, Font16, Clr000000)
	} else if card.Type == CardEquip {
		DrawText(screen, string(card.EquipType), x+55, y+160-10, AnchorBtmCenter, Font16, Clr000000)
	}
}

func GetSuitClr(suit CardSuit) color.Color {
	if suit == SuitHeart || suit == SuitDiamond {
		return ClrFF0000
	}
	return Clr000000
}

func GetHpClr(hp, maxHp int) color.Color {
	if hp > (maxHp+1)/2 { // 大于一半 绿色
		return ClrAAD745
	} else if hp > (maxHp+1)/4 { // 大于1/4 黄色
		return ClrFEF660
	} else { // 否则红色
		return ClrD84B3C
	}
}

//===================文本=======================

func VerticalText(val string) string {
	buff := strings.Builder{}
	items := []rune(val)
	for i := 0; i < len(items); i++ {
		if i > 0 {
			buff.WriteByte('\n')
		}
		buff.WriteRune(items[i])
	}
	return buff.String()
}

func Int2Str(val int) string {
	return strconv.FormatInt(int64(val), 10)
}
