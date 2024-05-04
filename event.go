/*
@author: sk
@date: 2024/5/1
*/
package main

type Event struct {
	Type  EventType
	Abort bool
	// 泛用参数
	Src        *Player   // 来源最多是一个玩家
	Desc       []*Player // 目标可以是多个
	Card       *CardWrap // 在使用的卡牌
	StageExtra *StageExtra
	StepExtra  *StepExtra
}

type Condition struct {
	Type ConditionType
	// 泛用参数
	Src, Desc *Player // 条件的目标对象与原对象
	Card      *Card   // 在使用的卡牌
	CardNum   int
	MaxCard   int
	MaxDesc   int  // 可以指定的最大目标数
	MaxUseSha int  // 最多使用杀的次数
	Dist      int  // 计算得到的距离，可能是玩家与玩家间的距离，也可能是武器的距离
	Invalid   bool // 是否不合法
}
