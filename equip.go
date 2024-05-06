/*
@author: sk
@date: 2024/5/4
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Equip struct {
	X, Y        float32
	Card        *Card        // 一般装备不太可能是转换牌
	SkillHolder *SkillHolder // 技能
	Enable      bool         // 用于控制装备是否有效
}

func NewEquip(card *Card, player *Player) *Equip {
	return &Equip{Card: card, SkillHolder: BuildSkillForEquip(card, player), Enable: true}
}

//装备栏：宽 200 高 40
func (e *Equip) Draw(screen *ebiten.Image) {
	pointAndSuit := fmt.Sprintf("%s%s", e.Card.Suit, e.Card.Point)
	suitClr := GetSuitClr(e.Card.Suit)
	DrawText(screen, pointAndSuit, e.X+10, e.Y+20, AnchorMidLeft, Font18, suitClr)
	DrawText(screen, e.Card.EquipAlias, e.X+200-10, e.Y+20, AnchorMidRight, Font18, ClrFFFFFF)
}

func BuildSkillForEquip(card *Card, player *Player) *SkillHolder {
	// 武器直接根据名称分配技能吧，玩家技能不会涉及武器上的某个效果
	switch card.Name {
	case "诸葛连弩":
		return NewSkillHolder(NewEquipShaDistSkill(player, 1), NewEquipZhuGeLianNuSkill(player))
	case "丈八蛇矛":
		return NewSkillHolder(NewEquipShaDistSkill(player, 3), NewEquipZhangBaSheMaoRespSkill(player),
			NewEquipZhangBaSheMaoActiveSkill(player))
	case "贯石斧":
		return NewSkillHolder(NewEquipShaDistSkill(player, 3), NewEquipGuanShiFuSkill(player))
	case "方天画戟":
		return NewSkillHolder(NewEquipShaDistSkill(player, 4), NewEquipFangTianHuaJiSkill(player))
	case "青虹剑":
		return NewSkillHolder(NewEquipShaDistSkill(player, 2), NewEquipQingHongJianSkill(player))
	case "麒麟弓":
		return NewSkillHolder(NewEquipShaDistSkill(player, 5), NewEquipQiLinGongSkill(player))
	case "雌雄双股剑":
		return NewSkillHolder(NewEquipShaDistSkill(player, 2), NewEquipCiXiongShuangGuJianSkill(player))
	case "青龙偃月刀":
		return NewSkillHolder(NewEquipShaDistSkill(player, 3), NewEquipQingLongYanYueDaoSkill(player))
	case "寒冰剑":
		return NewSkillHolder(NewEquipShaDistSkill(player, 2), NewEquipHanBingJianSkill(player))
	case "八卦阵":
		return NewSkillHolder(NewEquipBaGuaZhenSkill(player))
	case "仁王盾":
		return NewSkillHolder(NewEquipRenWangDunSkill(player))
	case "赤兔", "大宛", "紫骍":
		return NewSkillHolder(NewEquipAttackHorseSkill(player))
	case "爪黄飞电", "的卢", "绝影":
		return NewSkillHolder(NewEquipDefenseHorseSkill(player))
	default:
		panic(fmt.Sprintf("invalid equip card %s", card.Name))
	}
}
