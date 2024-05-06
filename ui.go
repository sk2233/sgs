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
	Show string
}

func (b *Button) Draw(screen *ebiten.Image) {
	FillRect(screen, b.X, b.Y, b.W, b.H, Clr348EBB)
	StrokeRect(screen, b.X, b.Y, b.W, b.H, 2, ClrA66F3F)
	DrawText(screen, b.Show, b.X+b.W/2, b.Y+b.H/2, AnchorMidCenter, Font18, ClrFFFFFF)
}

func NewButton(show string) *Button {
	bound := text.BoundString(Font18, show)
	w := float32(bound.Dx() + 20) // 10边距
	h := float32(bound.Dy() + 20)
	return &Button{Show: show, BaseRect: NewBaseRect(w, h)}
}

func NewButtons(shows ...string) []*Button {
	res := make([]*Button, 0)
	if len(shows) == 0 {
		return res
	}

	for _, show := range shows {
		res = append(res, NewButton(show))
	}
	last := float32(WinWidth - 240 - 240)
	for _, button := range res {
		last -= button.W
	}
	offset := last / float32(len(res)+1) // 计算偏移 若是按钮太多，offset会变成负数，按钮会叠在一起
	x := 240 + offset
	y := 280 + 280 - res[0].H - 20
	for _, button := range res {
		button.X, button.Y = x, y
		x += offset + button.W
	}
	return res
}

//=======================Text简单的文本展示没有点击事件=========================

type Text struct {
	X, Y float32
	Show string
}

func NewText(format string, args ...any) *Text {
	return &Text{Show: fmt.Sprintf(format, args...)}
}

func (t *Text) Draw(screen *ebiten.Image) {
	DrawText(screen, t.Show, t.X, t.Y, AnchorMidCenter, Font16, ClrFFFFFF)
}

//======================CardUI卡牌用于绘制的形式========================

type CardUI struct {
	*BaseRect
	Card      *Card
	Select0   bool
	CanSelect bool
	Flip      bool
}

func (c *CardUI) Draw(screen *ebiten.Image) {
	FillRect(screen, c.X, c.Y, c.W, c.H, ClrDECDBA)
	StrokeRect(screen, c.X, c.Y, c.W, c.H, 2, Clr000000)
	if c.Flip {
		DrawText(screen, "牌\n背", c.X+c.W/2, c.Y+c.H/2, AnchorMidCenter, Font18, Clr000000)
	} else {
		pointAndSuit := fmt.Sprintf("%s\n%s", c.Card.Suit, c.Card.Point)
		suitClr := GetSuitClr(c.Card.Suit)
		DrawText(screen, pointAndSuit, c.X+10, c.Y+10, AnchorTopLeft, Font18, suitClr)
		DrawText(screen, c.Card.Name, c.X+c.W/2, c.Y+c.H/2, AnchorMidCenter, Font16, Clr000000)
		if c.Card.Type == CardKit {
			DrawText(screen, string(c.Card.KitType), c.X+c.W/2, c.Y+c.H-10, AnchorBtmCenter, Font16, Clr000000)
		} else if c.Card.Type == CardEquip {
			DrawText(screen, string(c.Card.EquipType), c.X+c.W/2, c.Y+c.H-10, AnchorBtmCenter, Font16, Clr000000)
		}
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
	return &CardUI{Card: card, BaseRect: NewBaseRect(110, 160), Select0: false, CanSelect: true, Flip: false}
}

//====================AllCard======================

type AllCard struct {
	*BaseRect
	Cards  []*CardUI // 手牌  都是反面展示的
	Equips []*CardUI // 装备牌
	Kits   []*CardUI // 延时锦囊牌
	Alls   []*CardUI
}

func (c *AllCard) Draw(screen *ebiten.Image) {
	FillRect(screen, c.X, c.Y, c.W, c.H, Clr4B403F)
	StrokeRect(screen, c.X, c.Y, c.W, c.H, 2, ClrFFFFFF)
	// 绘制出区域轮廓
	StrokeRect(screen, c.X+20-2, c.Y+20-2, 110*6+20+4, 160+4, 2, ClrFFFFFF)
	StrokeRect(screen, c.X+20-2, c.Y+160+20*2-2, 110*4+4, 160+4, 2, ClrFFFFFF)
	StrokeRect(screen, c.X+110*4+20*2-2, c.Y+160+20*2-2, 110*2+4, 160+4, 2, ClrFFFFFF)
	for _, item := range c.Alls {
		item.Draw(screen)
	}
}

// 下面 6 张(4张装备，2个延时锦囊)，上面任意张，两排，居中显示  内部边距都是 20
// 卡牌：宽 110 高 160
func (c *AllCard) TidyCard() {
	x := c.X + 20
	y := c.Y + 20
	w := c.W - 20*2
	offset := float32(110)
	if float32(110*len(c.Cards)) > w {
		offset = (w - 110) / float32(len(c.Cards)-1)
	}
	for _, card := range c.Cards {
		card.X, card.Y = x, y
		x += offset
	}
	x = c.X + 20
	y = c.Y + 160 + 20*2
	offset = 110
	for _, equip := range c.Equips {
		equip.X, equip.Y = x+EquipIndexes[equip.Card.EquipType]*offset, y
	}
	x = c.X + 4*110 + 2*20
	for _, kit := range c.Kits {
		kit.X, kit.Y = x, y
		x += offset
	}
}

func (c *AllCard) GetSelectCard() []*Card {
	res := make([]*Card, 0)
	for _, item := range c.Alls {
		if item.Select0 {
			res = append(res, item.Card)
		}
	}
	return res
}

func (c *AllCard) ToggleCard(x, y float32) bool {
	for _, item := range c.Alls {
		if item.Click(x, y) {
			item.Toggle()
			return true
		}
	}
	return false
}

func (c *AllCard) SetAllCanSelect() {
	for _, item := range c.Alls {
		item.CanSelect = true
	}
}

func (c *AllCard) DarkLastCard() {
	for _, item := range c.Alls {
		if !item.Select0 {
			item.CanSelect = false
		}
	}
}

func NewAllCard(cards []*Card, equips []*Card, kits []*Card) *AllCard {
	res := &AllCard{Cards: Map(cards, NewCardUI), Equips: Map(equips, NewCardUI), Kits: Map(kits, NewCardUI),
		BaseRect: NewBaseRect(110*6+20*3, 160*2+20*3)}
	res.X = (WinWidth - res.W) / 2
	res.Y = (280*2 - 60 - res.H) / 2 // 要考虑不要遮挡按钮
	for _, card := range res.Cards {
		card.Flip = true
	}
	res.Alls = append(res.Cards, res.Equips...)
	res.Alls = append(res.Alls, res.Kits...)
	res.TidyCard()
	return res
}
