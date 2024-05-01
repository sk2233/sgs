/*
@author: sk
@date: 2024/5/1
*/
package main

type IStep interface {
	Update(effect *Effect, event *Event) // 执行效果并返回是否执行结束
}

//==================SysGameStartStep==================

type SysNextPlayerStep struct {
}

func NewSysNextPlayerStep() *SysNextPlayerStep {
	return &SysNextPlayerStep{}
}

func (s *SysNextPlayerStep) Update(effect *Effect, event *Event) {
	//effect.Index = MaxIndex
	//MainGame.NextPlayer()
}
