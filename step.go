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
	MainGame.NextPlayer()  // 轮到下一个玩家了
	extra.Index = MaxIndex // 结束效果
}

//===================TriggerEventStep简单触发一下事件=====================

type TriggerEventStep struct {
	EventType EventType
}

func NewTriggerEventStep(eventType EventType) *TriggerEventStep {
	return &TriggerEventStep{EventType: eventType}
}

func (t *TriggerEventStep) Update(event *Event, extra *StepExtra) {
	// 简单触发一下事件就继续向下走
	MainGame.TriggerEvent(&Event{Type: t.EventType, Src: event.Src}) // TODO 参数后续可能需要继续补充
	extra.Index++
}

//=====================DrawStageMainStep摸牌阶段的主要步骤========================

type DrawStageMainStep struct {
}

func NewDrawStageMainStep() *DrawStageMainStep {
	return &DrawStageMainStep{}
}

func (d *DrawStageMainStep) Update(event *Event, extra *StepExtra) {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionDrawStageCard, Src: event.Src})
	event.Src.DrawCard(condition.CardNum)
}
