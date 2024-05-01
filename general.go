/*
@author: sk
@date: 2024/5/1
*/
package main

import "fmt"

type General struct {
	Name      string
	Hp, MaxHp int   // 元数据不能改的
	Force     Force // 势力
}

var (
	generalMap   = make(map[string]*General)
	generalNames = make([]string, 0)
)

func InitGeneral() {
	generals := []*General{{
		Name:  "刘备",
		Hp:    4,
		MaxHp: 4,
		Force: ForceShu,
	}, {
		Name:  "孙尚香",
		Hp:    3,
		MaxHp: 3,
		Force: ForceWu,
	}, {
		Name:  "董卓",
		Hp:    8,
		MaxHp: 8,
		Force: ForceQun,
	}, {
		Name:  "张辽",
		Hp:    4,
		MaxHp: 4,
		Force: ForceWei,
	}, {
		Name:  "孙权",
		Hp:    4,
		MaxHp: 4,
		Force: ForceWu,
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
