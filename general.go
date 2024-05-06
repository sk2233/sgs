/*
@author: sk
@date: 2024/5/1
*/
package main

import "fmt"

type General struct {
	Name      string
	Hp, MaxHp int      // 元数据不能改的
	Force     Force    // 势力
	Skills    []string // 技能组
	Gender    Gender
}

var (
	generalMap   = make(map[string]*General)
	generalNames = make([]string, 0)
)

func InitGeneral() {
	generals := []*General{{
		Name:   "刘备",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceShu,
		Gender: GenderMan,
	}, {
		Name:   "孙尚香",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWu,
		Gender: GenderWoman,
	}, {
		Name:   "华雄",
		Hp:     6,
		MaxHp:  6,
		Force:  ForceQun,
		Gender: GenderMan,
	}, {
		Name:   "张辽",
		Hp:     3,
		MaxHp:  4,
		Force:  ForceWei,
		Gender: GenderMan,
	}, {
		Name:   "孙权",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWu,
		Gender: GenderMan,
	}}
	for _, general := range generals {
		generalMap[general.Name] = general
		generalNames = append(generalNames, general.Name)
	}
}

func GetGeneral(name string) *General {
	return generalMap[name]
}

func GetGeneralN(num int) []*General {
	if num > len(generalNames) {
		panic(fmt.Sprintf("num %d > len(generalNames) %d", num, len(generalNames)))
	}
	RandSlice(generalNames)
	res := make([]*General, 0)
	for i := 0; i < num; i++ {
		res = append(res, GetGeneral(generalNames[i]))
	}
	return res
}

func BuildSkillForPlayer(player *Player) *SkillHolder {
	if player.IsBot {
		return NewSkillHolder(NewBotAskCardSkill(player))
	} else {
		return NewSkillHolder(NewPlayerAskCardSkill(player))
	}
}
