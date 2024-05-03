/*
@author: sk
@date: 2024/5/1
*/
package main

type StepExtra struct {
	Index     int // 步骤进行到那里了
	JudgeCard *CardWrap
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

//======================JudgeStageCheckStep判定阶段检查步骤=======================

type JudgeStageCheckStep struct {
}

func NewJudgeStageCheckStep() *JudgeStageCheckStep {
	return &JudgeStageCheckStep{}
}

func (j *JudgeStageCheckStep) Update(event *Event, extra *StepExtra) {
	if len(event.Src.JudgeCards) > 0 { // 这时判定牌应该放到处理区了 TODO
		extra.Index++ // 还有判定牌接着判定
	} else {
		extra.Index = MaxIndex // 没有了结束
	}
}

//============================JudgeStageExecuteStep判定阶段判定牌生效完清理步骤=================================

type JudgeStageExecuteStep struct {
}

func NewJudgeStageExecuteStep() *JudgeStageExecuteStep {
	return &JudgeStageExecuteStep{}
}

func (j *JudgeStageExecuteStep) Update(event *Event, extra *StepExtra) { // 普通判定也有这也这一步骤，但是判定阶段需要构成循环
	extra.Index = 0
	src := event.Src
	card := src.JudgeCards[0]
	src.JudgeCards = src.JudgeCards[1:]
	MainGame.PushAction(NewEffectGroupBySkill(&Event{
		Type:      EventCardSkill,
		Src:       src,   // 这里相当于延时技能自己成来源了
		StepExtra: extra, // 主要为了传递判定牌
	}, card.Desc.Skill))
	MainGame.DiscardCard(append(card.Src, extra.JudgeCard.Src...)) // 丢弃延时锦囊牌与判定牌，若是他们还有实体牌的话
}

//===========================JudgeCardJudgeStep判定牌判定Step===============================

type JudgeCardJudgeStep struct {
}

func NewJudgeCardJudgeStep() *JudgeCardJudgeStep {
	return &JudgeCardJudgeStep{}
}

func (j *JudgeCardJudgeStep) Update(event *Event, extra *StepExtra) {
	extra.Index++
	extra.JudgeCard = NewSimpleCardWrap(MainGame.DrawCard(1)[0]) // 进行判定可能经历修改
	MainGame.TriggerEvent(&Event{Type: EventJudgeCard, Src: event.Src, StepExtra: extra})
}

//=======================JudgeCardEndStep判定牌生效步骤==========================

type JudgeCardEndStep struct {
}

func NewJudgeCardEndStep() *JudgeCardEndStep {
	return &JudgeCardEndStep{}
}

func (j *JudgeCardEndStep) Update(event *Event, extra *StepExtra) {
	// 判定结束后，目标可以判定牌做自己的操作并需要负责回收判定牌到弃牌区域
	extra.Index++
	MainGame.TriggerEvent(&Event{Type: EventJudgeEnd, Src: event.Src, StepExtra: extra}) // 触发判定生效事件
}
