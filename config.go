/*
@author: sk
@date: 2024/5/1
*/
package main

import "math/rand"

func LoadPlayer() []*Player {
	roles := []Role{RoleZhuGong, RoleZhongChen, RoleNeiJian, RoleFanZei, RoleFanZei}
	RandSlice(roles)
	res := make([]*Player, 0)
	generals := GetGeneralN(len(roles))
	//界面：宽 1200 高 720
	//装备栏：宽 200 高 40
	//武将头图：宽 200 高 120
	//身份,血条,手牌侧边：宽 40 高 280(玩家的话是160)
	poss := [][2]float32{{0, 280 * 2}, {WinWidth - 200 - 40, 280}, {(200 + 40) * 3, 0}, {200 + 40, 0}, {0, 280}}
	for i, role := range roles {
		pos := poss[i]
		res = append(res, NewPlayer(pos[0], pos[1], i != 0, generals[i], role))
	}
	return res
}

func LoadCard() []*Card {
	res := make([]*Card, 0)
	// 基本牌
	// 杀
	for i := 0; i < 30; i++ {
		res = append(res, &Card{Name: "杀", Point: CardPoint(rand.Intn(13) + 1),
			Suit: CardSuit(rand.Intn(4) + 1), Type: CardBasic, Skill: NewShaSkill()})
	}
	// 闪
	for i := 0; i < 15; i++ {
		res = append(res, &Card{Name: "闪", Point: CardPoint(rand.Intn(13) + 1),
			Suit: CardSuit(rand.Intn(4) + 1), Type: CardBasic, Skill: NewShanSkill()})
	}
	// 桃
	for i := 0; i < 8; i++ {
		res = append(res, &Card{Name: "桃", Point: CardPoint(rand.Intn(13) + 1),
			Suit: CardSuit(rand.Intn(4) + 1), Type: CardBasic, Skill: NewTaoSkill()})
	}
	// 装备牌
	// 武器
	for i := 0; i < 2; i++ {
		res = append(res, &Card{Name: "诸葛连弩", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "诸葛连弩1",
			Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	}
	res = append(res, &Card{Name: "丈八蛇矛", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "丈八蛇矛3",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "贯石斧", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "贯石斧3",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "方天画戟", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "方天画戟4",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "青虹剑", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "青虹剑2",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "麒麟弓", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "麒麟弓5",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "雌雄双股剑", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "雌雄双股剑2",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "青龙偃月刀", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "青龙偃月刀3",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "寒冰剑", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "寒冰剑2",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	// 防具
	for i := 0; i < 2; i++ {
		res = append(res, &Card{Name: "八卦阵", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "八卦阵",
			Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipArmor, Skill: NewEquipSkill()})
	}
	res = append(res, &Card{Name: "仁王盾", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "仁王盾",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipArmor, Skill: NewEquipSkill()})
	// 进攻马
	res = append(res, &Card{Name: "赤兔", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "赤兔-1",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipAttack, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "大宛", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "大宛-1",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipAttack, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "紫骍", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "紫骍-1",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipAttack, Skill: NewEquipSkill()})
	// 防御马
	res = append(res, &Card{Name: "爪黄飞电", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "爪黄飞电+1",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipDefense, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "的卢", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "的卢+1",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipDefense, Skill: NewEquipSkill()})
	res = append(res, &Card{Name: "绝影", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "绝影+1",
		Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipDefense, Skill: NewEquipSkill()})
	//// 锦囊牌
	//// 即时
	//for i := 0; i < 4; i++ {
	//	res = append(res, &Card{Name: "无中生有", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//	res = append(res, &Card{Name: "无懈可击", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//}
	//for i := 0; i < 6; i++ {
	//	res = append(res, &Card{Name: "过河拆桥", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//}
	//for i := 0; i < 5; i++ {
	//	res = append(res, &Card{Name: "顺手牵羊", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//}
	//for i := 0; i < 2; i++ {
	//	res = append(res, &Card{Name: "借刀杀人", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//	res = append(res, &Card{Name: "五谷丰登", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//}
	//for i := 0; i < 3; i++ {
	//	res = append(res, &Card{Name: "决斗", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//	res = append(res, &Card{Name: "南蛮入侵", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//}
	//res = append(res, &Card{Name: "万箭齐发", Point: CardPoint(rand.Intn(13) + 1),
	//	Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//res = append(res, &Card{Name: "桃园结义", Point: CardPoint(rand.Intn(13) + 1),
	//	Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitInstant})
	//// 延时
	//for i := 0; i < 3; i++ {
	//	res = append(res, &Card{Name: "乐不思蜀", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitDelay})
	//}
	//for i := 0; i < 2; i++ {
	//	res = append(res, &Card{Name: "闪电", Point: CardPoint(rand.Intn(13) + 1),
	//		Suit: CardSuit(rand.Intn(4) + 1), Type: CardKit, KitType: KitDelay})
	//}
	RandSlice(res)
	return res
}
