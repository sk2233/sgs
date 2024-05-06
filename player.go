/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	X, Y           float32              // 位置的左上角
	IsBot          bool                 // 是否为机器人
	Hp, MaxHp      int                  // 体力 体力上限
	General        *General             // 武将信息 主要包含一些元数据信息，需要变化的信息都会扩展到外面
	Role, MarkRole Role                 // 真实身份 表面标记身份
	Force          Force                // 势力
	Cards          []*CardUI            // 手牌
	JudgeCards     []*CardWrap          // 判定牌,可能是转换牌
	Equips         map[EquipType]*Equip // 装备
	SkillHolder    *SkillHolder         // 技能
	Select         bool                 // 是否被选择
	CanSelect      bool                 // 是否可以被选择
}

func NewPlayer(x, y float32, isBot bool, general *General, role Role) *Player {
	markRole := RoleUnknown
	hp := general.Hp
	maxHp := general.MaxHp
	if role == RoleZhuGong {
		markRole = role
		hp++
		maxHp++
	}
	res := &Player{
		X:        x,
		Y:        y,
		IsBot:    isBot,
		Hp:       hp,
		MaxHp:    maxHp,
		General:  general,
		Role:     role,
		MarkRole: markRole,
		Force:    general.Force,
		Cards:    make([]*CardUI, 0),
		Select:   false,
		Equips:   make(map[EquipType]*Equip),
	}
	res.SkillHolder = BuildSkillForPlayer(res)
	return res
}

//========================绘制相关=============================

func (p *Player) Draw(screen *ebiten.Image) {
	if p.IsBot {
		p.drawBot(screen)
	} else {
		p.drawPlayer(screen)
	}
}

//装备栏：宽 200 高 40
//武将头图：宽 200 高 120
//身份,血条,手牌侧边：宽 40 高 280
func (p *Player) drawBot(screen *ebiten.Image) {
	FillRect(screen, p.X, p.Y, 200, 280, Clr65401E)
	FillRect(screen, p.X+200, p.Y, 40, 280, Clr362618)
	DrawText(screen, string(p.Force), p.X+10, p.Y+10, AnchorTopLeft, Font18, ClrFFFFFF)
	name := VerticalText(p.General.Name)
	DrawText(screen, name, p.X+10, p.Y+50, AnchorTopLeft, Font18, ClrFFFFFF)
	role := VerticalText(string(p.MarkRole))
	DrawText(screen, role, p.X+200+20, p.Y, AnchorTopCenter, Font18, ClrFFFFFF)
	p.drawHp(screen, p.X+200+20, p.Y+50)
	for _, equip := range p.Equips {
		equip.Draw(screen)
	}
	cardNum := Int2Str(len(p.Cards))
	DrawText(screen, cardNum, p.X+200+20, p.Y+280-20, AnchorMidCenter, Font18, Clr000000)
	if p.Select {
		StrokeCircle(screen, p.X+120, p.Y+140, 40, 4, Clr00FF00)
	}
	if !p.CanSelect {
		FillRect(screen, p.X, p.Y, 200+40, 280, Clr00000080)
	}
}

//装备栏：宽 200 高 40
//武将头图：宽 200 高 120
//身份,血条,手牌侧边：宽 40 高 160  实际玩家的位置是相对确定的
func (p *Player) drawPlayer(screen *ebiten.Image) {
	FillRect(screen, p.X, p.Y, 200, 160, Clr553010)
	FillRect(screen, p.X+200, p.Y, WinWidth-200-40-200, 160, Clr0D0D0D)
	FillRect(screen, p.X+WinWidth-200-40, p.Y, 200, 160, Clr553010)
	FillRect(screen, p.X+WinWidth-40, p.Y, 40, 160, Clr362618)
	DrawText(screen, string(p.Force), p.X+WinWidth-200-40+10, p.Y+10, AnchorTopLeft, Font18, ClrFFFFFF)
	name := VerticalText(p.General.Name)
	DrawText(screen, name, p.X+WinWidth-200-40+10, p.Y+50, AnchorTopLeft, Font18, ClrFFFFFF)
	role := VerticalText(string(p.Role))
	DrawText(screen, role, p.X+WinWidth-20, p.Y, AnchorTopCenter, Font18, ClrFFFFFF)
	p.drawHp(screen, p.X+WinWidth-20, p.Y+50)
	for _, equip := range p.Equips {
		equip.Draw(screen)
	}
	for _, card := range p.Cards {
		card.Draw(screen)
	}
	if p.Select {
		StrokeCircle(screen, p.X+WinWidth-120, p.Y+80, 40, 4, Clr00FF00)
	}
	if !p.CanSelect {
		FillRect(screen, p.X+WinWidth-200-40, p.Y, 200+40, 160, Clr00000080)
	}
}

func (p *Player) drawHp(screen *ebiten.Image, x float32, y float32) {
	clr := GetHpClr(p.Hp, p.MaxHp)
	if p.MaxHp > 5 { // 缩略展示
		FillCircle(screen, x, y+11, 10, clr)
		hpStr := fmt.Sprintf("%d\n\\\n%d", p.Hp, p.MaxHp)
		DrawText(screen, hpStr, x, y+22, AnchorTopCenter, Font18, clr)
	} else { // 全部展示
		for i := 0; i < p.Hp; i++ {
			FillCircle(screen, x, y+11+22*float32(i), 10, clr)
		}
		for i := p.Hp; i < p.MaxHp; i++ {
			StrokeCircle(screen, x, y+11+22*float32(i), 10, 2, ClrFFFFFF)
		}
	}
}

func (p *Player) DrawCard(num int) {
	if num == 0 {
		return
	}
	cards := MainGame.DrawCard(num)
	p.AddCard(cards...)
}

func (p *Player) AddCard(cards ...*Card) {
	if len(cards) == 0 {
		return
	}
	p.Cards = append(p.Cards, Map(cards, NewCardUI)...)
	p.TidyCard()
}

func (p *Player) RemoveCard(cards ...*Card) {
	if len(cards) == 0 {
		return
	}
	set := NewSet[*Card](cards...)
	p.Cards = Filter(p.Cards, func(item *CardUI) bool {
		return !set.Contain(item.Card)
	})
	p.TidyCard()
}

func (p *Player) RemoveEquip(cards ...*Card) {
	if len(cards) == 0 {
		return
	}
	set := NewSet[*Card](cards...)
	for type0, equip := range p.Equips {
		if set.Contain(equip.Card) {
			delete(p.Equips, type0)
		}
	}
}

func (p *Player) GetEquipSkillHolders() []*SkillHolder {
	res := make([]*SkillHolder, 0)
	for _, equip := range p.Equips {
		if equip.Enable {
			res = append(res, equip.SkillHolder)
		}
	}
	return res
}

//卡牌：宽 110 高 160  范围从 200 ～ 1200-200-40 只有非bot才需要绘制
func (p *Player) TidyCard() {
	if p.IsBot {
		return
	}
	offset := 110
	if len(p.Cards)*110 > WinWidth-200-200-40 {
		offset = (WinWidth - 200 - 200 - 40 - 110) / (len(p.Cards) - 1)
	}
	for i := 0; i < len(p.Cards); i++ {
		p.Cards[i].X, p.Cards[i].Y = float32(200+i*offset), p.Y
	}
}

func (p *Player) ResetCard() {
	for i := 0; i < len(p.Cards); i++ {
		p.Cards[i].CanSelect = true
		p.Cards[i].Select0 = false
	}
}

func (p *Player) DarkLastCard() {
	for i := 0; i < len(p.Cards); i++ {
		if !p.Cards[i].Select0 {
			p.Cards[i].CanSelect = false
		}
	}
}

func (p *Player) ToggleCard(x, y float32) bool {
	for _, card := range p.Cards {
		if card.Click(x, y) {
			card.Toggle()
			return true
		}
	}
	return false
}

//装备栏：宽 200 高 40<br>
//武将头图：宽 200 高 120<br>
//身份,血条,手牌侧边：宽 40 高 280(玩家的话是160)<br>
func (p *Player) ToggleSelect(tx, ty float32) bool {
	if !p.CanSelect {
		return false
	}
	x := p.X
	y := p.Y
	w := float32(200 + 40)
	h := float32(280)
	if !p.IsBot {
		x = WinWidth - w
		h = 160
	}
	if tx > x && tx < x+w && ty > y && ty < y+h {
		p.Select = !p.Select
		return true
	}
	return false
}

func (p *Player) GetSelectCard() []*Card {
	res := make([]*Card, 0)
	for _, card := range p.Cards {
		if card.Select0 {
			res = append(res, card.Card)
		}
	}
	return res
}

func (p *Player) CheckCard(extra *StepExtra) {
	for _, card := range p.Cards {
		card.CanSelect = card.Card.Skill.CheckUse(p, card.Card, extra)
	}
}

func (p *Player) CheckCardByWrapFilter(filter CardWrapFilter) {
	for _, card := range p.Cards {
		card.CanSelect = filter(NewSimpleCardWrap(card.Card))
	}
}

func (p *Player) CheckCardByFilter(filter CardFilter) {
	for _, card := range p.Cards {
		card.CanSelect = filter(card.Card)
	}
}

func (p *Player) ChangeHp(val int) bool {
	p.Hp += val
	if p.Hp > p.MaxHp {
		p.Hp = p.MaxHp
	}
	return p.Hp <= 0
}

//装备栏：宽 200 高 40
func (p *Player) AddEquip(card *Card) *Card {
	old := p.Equips[card.EquipType]
	equip := NewEquip(card, p)
	y := p.Y + 40*EquipIndexes[card.EquipType]
	if p.IsBot {
		y += 120
	}
	equip.X, equip.Y = p.X, y
	p.Equips[card.EquipType] = equip

	if old == nil {
		return nil
	}
	return old.Card
}
