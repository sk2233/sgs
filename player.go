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
	X, Y           float32      // 位置的左上角
	IsBot          bool         // 是否为机器人
	Hp, MaxHp      int          // 体力 体力上限
	General        *General     // 武将信息 主要包含一些元数据信息，需要变化的信息都会扩展到外面
	Role, MarkRole Role         // 真实身份 表面标记身份
	Force          Force        // 势力
	Cards          []*Card      // 手牌
	JudgeCards     []*CardWrap  // 判定牌,可能是转换牌
	SkillHolder    *SkillHolder // 技能
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
	return &Player{
		X:           x,
		Y:           y,
		IsBot:       isBot,
		Hp:          hp,
		MaxHp:       maxHp,
		General:     general,
		Role:        role,
		MarkRole:    markRole,
		Force:       general.Force,
		Cards:       make([]*Card, 0),
		SkillHolder: NewSkillHolder(),
	}
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
	cardNum := Int2Str(len(p.Cards))
	DrawText(screen, cardNum, p.X+200+20, p.Y+280-20, AnchorMidCenter, Font18, Clr000000)
}

//装备栏：宽 200 高 40
//武将头图：宽 200 高 120
//身份,血条,手牌侧边：宽 40 高 160  实际玩家的位置是相对确定的
func (p *Player) drawPlayer(screen *ebiten.Image) {
	FillRect(screen, p.X, p.Y, WinWidth-200-40, 160, Clr0D0D0D)
	FillRect(screen, p.X+WinWidth-200-40, p.Y, 200, 160, Clr553010)
	FillRect(screen, p.X+WinWidth-40, p.Y, 40, 160, Clr362618)
	DrawText(screen, string(p.Force), p.X+WinWidth-200-40+10, p.Y+10, AnchorTopLeft, Font18, ClrFFFFFF)
	name := VerticalText(p.General.Name)
	DrawText(screen, name, p.X+WinWidth-200-40+10, p.Y+50, AnchorTopLeft, Font18, ClrFFFFFF)
	role := VerticalText(string(p.Role))
	DrawText(screen, role, p.X+WinWidth-20, p.Y, AnchorTopCenter, Font18, ClrFFFFFF)
	p.drawHp(screen, p.X+WinWidth-20, p.Y+50)
	p.drawCard(screen)
}

//卡牌：宽 110 高 160  范围从 200 ～ 1200-200-40 只有非bot才需要绘制
func (p *Player) drawCard(screen *ebiten.Image) {
	offset := 110
	if len(p.Cards)*110 > WinWidth-200-200-40 {
		offset = (WinWidth - 200 - 200 - 40 - 110) / (len(p.Cards) - 1)
	}
	for i := 0; i < len(p.Cards); i++ {
		DrawCard(screen, float32(200+i*offset), p.Y, p.Cards[i])
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
	cards := MainGame.DrawCard(num)
	p.Cards = append(p.Cards, cards...)
}

func (p *Player) GetEquipSkillHolders() []*SkillHolder {
	return make([]*SkillHolder, 0)
}
