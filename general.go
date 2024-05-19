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
		Name:   "曹操",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWei,
		Skills: []string{"奸雄", "护驾"},
		Gender: GenderMan,
	}, {
		Name:   "司马懿",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWei,
		Skills: []string{"反馈", "鬼才"},
		Gender: GenderMan,
	}, {
		Name:   "夏侯惇",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWei,
		Skills: []string{"刚烈"},
		Gender: GenderMan,
	}, {
		Name:   "张辽",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWei,
		Skills: []string{"突袭"},
		Gender: GenderMan,
	}, {
		Name:   "许褚",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWei,
		Skills: []string{"裸衣"},
		Gender: GenderMan,
	}, {
		Name:   "郭嘉",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWei,
		Skills: []string{"天妒", "遗计"},
		Gender: GenderMan,
	}, {
		Name:   "甄姬",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWei,
		Skills: []string{"倾国", "洛神"},
		Gender: GenderWoman,
	}, {
		Name:   "刘备",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceShu,
		Skills: []string{"仁德", "激将"},
		Gender: GenderMan,
	}, {
		Name:   "关羽",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceShu,
		Skills: []string{"武圣"},
		Gender: GenderMan,
	}, {
		Name:   "张飞",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceShu,
		Skills: []string{"咆哮"},
		Gender: GenderMan,
	}, {
		Name:   "诸葛亮",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceShu,
		Skills: []string{"观星", "空城"},
		Gender: GenderMan,
	}, {
		Name:   "赵云",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceShu,
		Skills: []string{"龙胆"},
		Gender: GenderMan,
	}, {
		Name:   "马超",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceShu,
		Skills: []string{"马术", "铁骑"},
		Gender: GenderMan,
	}, {
		Name:   "黄月英",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceShu,
		Skills: []string{"集智", "奇才"},
		Gender: GenderWoman,
	}, {
		Name:   "孙权",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWu,
		Skills: []string{"制衡", "救援"},
		Gender: GenderMan,
	}, {
		Name:   "甘宁",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWu,
		Skills: []string{"奇袭"},
		Gender: GenderMan,
	}, {
		Name:   "吕蒙",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWu,
		Skills: []string{"克己"},
		Gender: GenderMan,
	}, {
		Name:   "黄盖",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceWu,
		Skills: []string{"苦肉"},
		Gender: GenderMan,
	}, {
		Name:   "周瑜",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWu,
		Skills: []string{"英姿", "反间"},
		Gender: GenderMan,
	}, {
		Name:   "大乔",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWu,
		Skills: []string{"国色", "流离"},
		Gender: GenderWoman,
	}, {
		Name:   "陆逊",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWu,
		Skills: []string{"谦逊", "连营"},
		Gender: GenderMan,
	}, {
		Name:   "孙尚香",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceWu,
		Skills: []string{"枭姬", "结姻"},
		Gender: GenderWoman,
	}, {
		Name:   "华佗",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceQun,
		Skills: []string{"青囊", "急救"},
		Gender: GenderMan,
	}, {
		Name:   "吕布",
		Hp:     4,
		MaxHp:  4,
		Force:  ForceQun,
		Skills: []string{"无双"},
		Gender: GenderMan,
	}, {
		Name:   "貂蝉",
		Hp:     3,
		MaxHp:  3,
		Force:  ForceQun,
		Skills: []string{"离间", "闭月"},
		Gender: GenderWoman,
	}, {
		Name:   "华雄",
		Hp:     6,
		MaxHp:  6,
		Force:  ForceQun,
		Skills: []string{"耀武"},
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
	skills := make([]ISkill, 0)
	for _, skill := range player.General.Skills {
		skills = append(skills, CreateSkillForPlayer(skill, player))
	}
	if player.IsBot {
		skills = append(skills, NewBotAskCardSkill(player), NewBotChooseCardSkill(player), NewPlayerDyingSkill(player))
	} else {
		skills = append(skills, NewPlayerAskCardSkill(player), NewPlayerChooseCardSkill(player), NewPlayerDyingSkill(player))
	}
	skills = Filter(skills, func(skill ISkill) bool {
		// 非主公过滤掉主公技
		if player.Role != RoleZhuGong && (skill.GetTag()&TagZhuGong) > 0 {
			return false
		}
		return true
	})
	return NewSkillHolder(skills...)
}

func CreateSkillForPlayer(skill string, player *Player) ISkill {
	switch skill {
	case "奸雄":
		return NewJianXiongSkill(player)
	case "护驾":
		return NewHuJiaSkill(player)
	case "反馈":
		return NewFanKuiSkill(player)
	case "鬼才":
		return NewGuiCaiSkill(player)
	case "刚烈":
		return NewGangLieSkill(player)
	case "突袭":
		return NewTuXiSkill(player)
	case "裸衣":
		return NewLuoYiSkill(player)
	case "天妒":
		return NewTianDuSkill(player)
	case "遗计":
		return NewYiJiSkill(player)
	case "倾国":
		return NewQingGuoSkill(player)
	case "洛神":
		return NewLuoShenSkill(player)
	case "仁德":
		return NewRenDeSkill(player)
	case "激将":
		return NewJiJiangSkill(player)
	case "武圣":
		return NewWuShengSkill(player)
	case "咆哮":
		return NewPaoXiaoSkill(player)
	case "观星":
		return NewGuanXingSkill(player)
	case "空城":
		return NewKongChengSkill(player)
	case "龙胆":
		return NewLongDanSkill(player)
	case "马术":
		return NewMaShuSkill(player)
	case "铁骑":
		return NewTieQiSkill(player)
	case "集智":
		return NewJiZhiSkill(player)
	case "奇才":
		return NewQiCaiSkill(player)
	case "制衡":
		return NewZhiHengSkill(player)
	case "救援":
		return NewJiuYuanSkill(player)
	case "奇袭":
		return NewQiXiSkill(player)
	case "克己":
		return NewKeJiSkill(player)
	case "苦肉":
		return NewKuRouSkill(player)
	case "英姿":
		return NewYingZiSkill(player)
	case "反间":
		return NewFanJianSkill(player)
	case "国色":
		return NewGuoSeSkill(player)
	case "流离":
		return NewLiuLiSkill(player)
	case "谦逊":
		return NewQianXunSkill(player)
	case "连营":
		return NewLianYingSkill(player)
	case "枭姬":
		return NewXiaoJiSkill(player)
	case "结姻":
		return NewJieYinSkill(player)
	case "青囊":
		return NewQingNangSkill(player)
	case "急救":
		return NewJiJiuSkill(player)
	case "无双":
		return NewWuShuangSkill(player)
	case "离间":
		return NewLiJianSkill(player)
	case "闭月":
		return NewBiYueSkill(player)
	case "耀武":
		return NewYaoWuSkill(player)
	default:
		panic(fmt.Sprintf("invalid skill %v", skill))
	}
}
