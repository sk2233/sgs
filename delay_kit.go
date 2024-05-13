/*
@author: sk
@date: 2024/5/12
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type DelayKit struct {
	X, Y  float32
	Card  *CardWrap
	Skill ISkill
}

func (d *DelayKit) Draw(screen *ebiten.Image) {
	DrawText(screen, d.Card.Desc.Alias, d.X, d.Y, AnchorTopLeft, Font18, ClrFFFFFF)
}

func NewDelayKit(card *CardWrap) *DelayKit {
	return &DelayKit{Card: card, Skill: GetSkillForDelayKit(card.Desc)}
}

func GetSkillForDelayKit(card *Card) ISkill {
	switch card.Name {
	case "乐不思蜀":
		return NewLeBuSiShuSkill()
	case "闪电":
		return NewShanDianSkill()
	default:
		panic(fmt.Sprintf("invalid delay kit card %s", card.Name))
	}
}
