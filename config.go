/*
@author: sk
@date: 2024/5/1
*/
package main

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
	for i := 0; i < 108; i++ {
		res = append(res, &Card{})
	}
	RandSlice(res)
	return res
}
