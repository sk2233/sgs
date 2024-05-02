/*
@author: sk
@date: 2024/5/1
*/
package main

type StepExtra struct {
	Index int // 步骤进行到那里了
}

func NewStepExtra() *StepExtra {
	return &StepExtra{Index: 0}
}

type IStep interface {
	Update(event *Event, extra *StepExtra) // 执行效果并返回是否执行结束
}

//==================SysGameStartStep==================

type SysNextPlayerStep struct {
}

func NewSysNextPlayerStep() *SysNextPlayerStep {
	return &SysNextPlayerStep{}
}

func (s *SysNextPlayerStep) Update(event *Event, extra *StepExtra) {
	extra.Index = MaxIndex // 结束效果
	MainGame.NextPlayer()  // 轮到下一个玩家了
}
