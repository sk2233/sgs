/*
@author: sk
@date: 2024/5/2
*/
package main

type StageExtra struct {
	SkipStage StageType
}

func NewStageExtra() *StageExtra {
	return &StageExtra{SkipStage: StageNone}
}

type IStage interface {
	Update(player *Player, extra *StageExtra) bool
	GetStage() StageType
}

type BaseStage struct {
	Steps []IStep    // 一个阶段有多个步骤组成
	Extra *StepExtra // 多个步骤存储的中间变量
}

func NewBaseStage(steps ...IStep) *BaseStage {
	return &BaseStage{Steps: steps, Extra: NewStepExtra()}
}

func (b *BaseStage) Update(player *Player, extra *StageExtra) bool {
	if b.Extra.Index < len(b.Steps) {
		b.Steps[b.Extra.Index].Update(&Event{ // 大部分到Step层就不再区分事件了
			Type:  EventPlayerStage, // 这里为了调用Step必须使用Event不太好 TODO
			Src:   player,
			Extra: extra,
		}, b.Extra)
		return false
	}
	return true
}

func (b *BaseStage) GetStage() StageType {
	return StageNone
}

//===================PrepareStage准备阶段=====================

type PrepareStage struct {
	*BaseStage
}

func (p *PrepareStage) GetStage() StageType {
	return StagePrepare
}

func NewPrepareStage() *PrepareStage {
	return &PrepareStage{BaseStage: NewBaseStage(NewTriggerEventStep(EventStagePrepare))}
}

//=====================JudgeStage判定阶段=====================

type JudgeStage struct {
	*BaseStage
}

func (j *JudgeStage) GetStage() StageType {
	return StageJudge
}

func NewJudgeStage() *JudgeStage {
	return &JudgeStage{BaseStage: NewBaseStage()}
}

//===================DrawStage摸牌阶段====================

type DrawStage struct {
	*BaseStage
}

func (d *DrawStage) GetStage() StageType {
	return StageDraw
}

func NewDrawStage() *DrawStage {
	return &DrawStage{BaseStage: NewBaseStage(NewDrawStageMainStep())}
}

//====================PlayStage出牌阶段======================

type PlayStage struct {
	*BaseStage
}

func (p *PlayStage) GetStage() StageType {
	return StagePlay
}

func NewPlayStage() *PlayStage {
	return &PlayStage{BaseStage: NewBaseStage()}
}

//====================DiscardStage弃牌阶段======================

type DiscardStage struct {
	*BaseStage
}

func (d *DiscardStage) GetStage() StageType {
	return StageDiscard
}

func NewDiscardStage() *DiscardStage {
	return &DiscardStage{BaseStage: NewBaseStage()}
}

//====================EndStage回合结束阶段========================

type EndStage struct {
	*BaseStage
}

func (e *EndStage) GetStage() StageType {
	return StageEnd
}

func NewEndStage() *EndStage {
	return &EndStage{BaseStage: NewBaseStage(NewTriggerEventStep(EventStageEnd))}
}
