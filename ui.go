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
	*BaseRect // 高  42
	Show      string
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
	last := float32(WinWidth - 240*2)
	for _, button := range res {
		last -= button.W
	}
	offset := last / float32(len(res)+1) // 计算偏移 若是按钮太多，offset会变成负数，按钮会叠在一起
	x := 240 + offset
	y := float32(280*2 - 42 - 20)
	for _, button := range res {
		button.X, button.Y = x, y
		x += offset + button.W
	}
	return res
}

//=======================Text简单的文本展示没有点击事件=========================

type Text struct {
	X, Y float32
	Show string // 高 20
}

func NewText(format string, args ...any) *Text {
	return &Text{Show: fmt.Sprintf(format, args...), X: WinWidth / 2, Y: 280*2 - 42 - 20*2}
}

func (t *Text) Draw(screen *ebiten.Image) {
	DrawText(screen, t.Show, t.X, t.Y, AnchorBtmCenter, Font16, ClrFFFFFF)
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
	for i := len(c.Alls) - 1; i >= 0; i-- {
		item := c.Alls[i]
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

//===================ChooseCard======================

type ChooseCard struct {
	*BaseRect
	Cards []*CardUI
}

func (c *ChooseCard) Draw(screen *ebiten.Image) {
	FillRect(screen, c.X, c.Y, c.W, c.H, Clr4B403F)
	StrokeRect(screen, c.X, c.Y, c.W, c.H, 2, ClrFFFFFF)
	// 绘制出区域轮廓  卡牌：宽 110 高 160
	StrokeRect(screen, c.X+20-2, c.Y+20-2, 110*5+4, 160+4, 2, ClrFFFFFF)
	for _, card := range c.Cards {
		card.Draw(screen)
	}
}

// 边距 20  5张牌宽
// 卡牌：宽 110 高 160
func (c *ChooseCard) TidyCard() {
	offset := float32(110)
	if len(c.Cards)*110 > 5*110 {
		offset = (5*110 - 110) / float32(len(c.Cards)-1)
	}
	x := c.X + 20
	y := c.Y + 20
	for _, card := range c.Cards {
		card.X, card.Y = x, y
		x += offset
	}
}

func (c *ChooseCard) GetSelectCard() []*Card {
	res := make([]*Card, 0)
	for _, card := range c.Cards {
		if card.Select0 {
			res = append(res, card.Card)
		}
	}
	return res
}

func (c *ChooseCard) Reset() {
	for _, card := range c.Cards {
		card.Select0 = false
		card.CanSelect = true
	}
	c.TidyCard()
}

func (c *ChooseCard) ToggleCard(x float32, y float32) bool {
	for i := len(c.Cards) - 1; i >= 0; i-- {
		card := c.Cards[i]
		if card.Click(x, y) {
			card.Toggle()
			return true
		}
	}
	return false
}

func (c *ChooseCard) SetAllCanSelect() {
	for _, card := range c.Cards {
		card.CanSelect = true
	}
}

func (c *ChooseCard) DarkLastCard() {
	for _, card := range c.Cards {
		if !card.Select0 {
			card.CanSelect = false
		}
	}
}

func NewChooseCard(cards []*Card) *ChooseCard {
	res := &ChooseCard{Cards: Map(cards, NewCardUI), BaseRect: NewBaseRect(110*5+40, 160+40)}
	res.X = (WinWidth - res.W) / 2
	res.Y = (280*2 - 60 - res.H) / 2 // 要考虑不要遮挡按钮
	res.TidyCard()
	return res
}

//==================GameOver=================

type GameOver struct {
	Info string
}

func NewGameOver(info string) *GameOver {
	return &GameOver{Info: info}
}

func (g *GameOver) Draw(screen *ebiten.Image) {
	DrawText(screen, g.Info, WinWidth/2, WinHeight/2, AnchorMidCenter, Font64, ClrFFFFFF)
}

//=================SkillUI==================

type SkillUI struct {
	*BaseRect
	Skill ISkill
}

func (s *SkillUI) Draw(screen *ebiten.Image) {
	if s.Skill.GetTag()&TagActive > 0 { // 主动技能标识
		StrokeRect(screen, s.X+1, s.Y+1, s.W-2, s.H-2, 2, Clr00FF00)
	} else {
		StrokeRect(screen, s.X+1, s.Y+1, s.W-2, s.H-2, 2, ClrFFFFFF)
	}
	name := VerticalText(s.Skill.GetName())
	DrawText(screen, name, s.X+10, s.Y+10, AnchorTopLeft, Font18, ClrFFFFFF)
}

func (s *SkillUI) Click(x, y float32) bool {
	return (s.Skill.GetTag()&TagActive) > 0 && s.BaseRect.Click(x, y)
}

func NewSkillUI(skill ISkill) *SkillUI {
	// 技能都是两个字的，边距 10
	return &SkillUI{Skill: skill, BaseRect: NewBaseRect(42, 66)}
}

//======================================

type GuanXing struct {
	*BaseRect
	Cards      []*CardUI // 原来的顺序
	OrderCards []*CardUI // 长度为 10 前 5 个位于上面，后 5 个位于下面
	SelectCard *CardUI
}

func (g *GuanXing) Draw(screen *ebiten.Image) {
	FillRect(screen, g.X, g.Y, g.W, g.H, Clr4B403F)
	StrokeRect(screen, g.X, g.Y, g.W, g.H, 2, ClrFFFFFF)
	// 绘制出区域轮廓
	StrokeRect(screen, g.X+20-2, g.Y+20-2, 110*5+4, 160+4, 2, ClrFFFFFF)
	StrokeRect(screen, g.X+20-2, g.Y+160+20*2-2, 110*5+4, 160+4, 2, ClrFFFFFF)
	for _, card := range g.Cards {
		card.Draw(screen)
	}
}

func (g *GuanXing) TidyCard() {
	for i, card := range g.OrderCards {
		x := g.X + 20 + float32(i%5)*110
		y := g.Y + 20 + float32(i/5)*(160+20)
		card.X, card.Y = x, y
	}
}

func (g *GuanXing) ToggleCard(x float32, y float32) {
	for _, card := range g.OrderCards {
		if card.Click(x, y) {
			if g.SelectCard == nil { // 第一个选择的牌不能是占位的
				if len(card.Card.Name) > 0 {
					card.Toggle()
					g.SelectCard = card
				}
			} else if g.SelectCard == card { // 取消选择
				card.Toggle()
				g.SelectCard = nil
			} else { // 交换位置
				index1 := g.getIndex(g.SelectCard)
				index2 := g.getIndex(card)
				g.OrderCards[index1], g.OrderCards[index2] = g.OrderCards[index2], g.OrderCards[index1]
				g.SelectCard.Toggle()
				g.SelectCard = nil
				g.TidyCard()
			}
			return
		}
	}

}

func (g *GuanXing) getIndex(card *CardUI) int {
	for i := 0; i < len(g.OrderCards); i++ {
		if g.OrderCards[i] == card {
			return i
		}
	}
	return -1
}

func (g *GuanXing) GetUpCards() []*Card {
	res := make([]*Card, 0)
	for i := 0; i < 5; i++ {
		if len(g.OrderCards[i].Card.Name) > 0 {
			res = append(res, g.OrderCards[i].Card)
		}
	}
	return res
}

func (g *GuanXing) GetDownCards() []*Card {
	res := make([]*Card, 0)
	for i := 5; i < 10; i++ {
		if len(g.OrderCards[i].Card.Name) > 0 {
			res = append(res, g.OrderCards[i].Card)
		}
	}
	return res
}

// 110 * 160
func NewGuanXing(cards []*Card) *GuanXing {
	res := &GuanXing{BaseRect: NewBaseRect(110*5+20*2, 160*2+20*3), Cards: Map(cards, NewCardUI), OrderCards: make([]*CardUI, 10)}
	for i, card := range res.Cards {
		res.OrderCards[i] = card
	}
	for i := len(cards); i < 10; i++ {
		res.OrderCards[i] = NewCardUI(&Card{}) // 占位
	}
	res.X = (WinWidth - res.W) / 2
	res.Y = (280*2 - 60 - res.H) / 2 // 要考虑不要遮挡按钮
	res.TidyCard()
	return res
}
