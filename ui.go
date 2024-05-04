/*
@author: sk
@date: 2024/5/3
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type IRect interface {
	GetTop() float32
	GetBottom() float32
	GetLeft() float32
	GetRight() float32
}

type BaseRect struct {
	X, Y float32
	W, H float32
}

func (b *BaseRect) GetTop() float32 {
	return b.Y
}

func (b *BaseRect) GetBottom() float32 {
	return b.Y + b.H
}

func (b *BaseRect) GetLeft() float32 {
	return b.X
}

func (b *BaseRect) GetRight() float32 {
	return b.X + b.W
}

func (b *BaseRect) Click(x, y float32) bool {
	return x > b.GetLeft() && x < b.GetRight() && y > b.GetTop() && y < b.GetBottom()
}

func NewBaseRect(w float32, h float32) *BaseRect {
	return &BaseRect{X: 0, Y: 0, W: w, H: h}
}

//==================Button文本基础按钮===================

type Button struct {
	*BaseRect
	Text string
}

func (b *Button) Draw(screen *ebiten.Image) {
	FillRect(screen, b.X, b.Y, b.W, b.H, Clr348EBB)
	StoryRect(screen, b.X, b.Y, b.W, b.H, 2, ClrA66F3F)
	DrawText(screen, b.Text, b.X+b.W/2, b.Y+b.H/2, AnchorMidCenter, Font18, ClrFFFFFF)
}

func NewButton(show string) *Button {
	bound := text.BoundString(Font18, show)
	w := float32(bound.Dx() + 20) // 10边距
	h := float32(bound.Dy() + 20)
	return &Button{Text: show, BaseRect: NewBaseRect(w, h)}
}

//======================CardUI卡牌用于绘制的形式========================

type CardUI struct {
	*BaseRect
	Card      *Card
	Select0   bool
	CanSelect bool
}

func (c *CardUI) Draw(screen *ebiten.Image) {
	FillRect(screen, c.X, c.Y, c.W, c.H, ClrDECDBA)
	StoryRect(screen, c.X, c.Y, c.W, c.H, 2, Clr000000)
	pointAndSuit := fmt.Sprintf("%s\n%s", c.Card.Suit, c.Card.Point)
	suitClr := GetSuitClr(c.Card.Suit)
	DrawText(screen, pointAndSuit, c.X+10, c.Y+10, AnchorTopLeft, Font18, suitClr)
	DrawText(screen, c.Card.Name, c.X+c.W/2, c.Y+c.H/2, AnchorMidCenter, Font16, Clr000000)
	if c.Card.Type == CardKit {
		DrawText(screen, string(c.Card.KitType), c.X+c.W/2, c.Y+c.H-10, AnchorBtmCenter, Font16, Clr000000)
	} else if c.Card.Type == CardEquip {
		DrawText(screen, string(c.Card.EquipType), c.X+c.W/2, c.Y+c.H-10, AnchorBtmCenter, Font16, Clr000000)
	}
	if !c.CanSelect {
		FillRect(screen, c.X, c.Y, c.W, c.H, Clr00000080)
	}
}

func (c *CardUI) Toggle() {
	c.Select0 = !c.Select0
	if c.Select0 {
		c.Y -= 20
	} else {
		c.Y += 20
	}
}

func (c *CardUI) Click(x, y float32) bool {
	return c.CanSelect && c.BaseRect.Click(x, y)
}

func (c *CardUI) UnSelect() {
	c.Select0 = false
	c.Y += 20
}

func (c *CardUI) Select() {
	c.Select0 = true
	c.Y -= 20
}

// 卡牌：宽 110 高 160
func NewCardUI(card *Card) *CardUI {
	return &CardUI{Card: card, BaseRect: NewBaseRect(110, 160), Select0: false}
}
