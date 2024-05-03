/*
@author: sk
@date: 2024/5/1
*/
package main

type Event struct {
	Type  EventType
	Abort bool
	// 泛用参数
	Src        *Player // 来源最多是一个玩家
	StageExtra *StageExtra
	StepExtra  *StepExtra
}

type Condition struct {
	Type ConditionType
	// 泛用参数
	Src, Desc *Player // 条件的目标对象与原对象只能是一个
	CardNum   int
	MaxCard   int
}
