/*
@author: sk
@date: 2024/5/2
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

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
			Type:       EventPlayerStage, // 这里为了调用Step必须使用Event不太好 TODO
			Src:        player,
			StageExtra: extra,
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
	return &JudgeStage{BaseStage: NewBaseStage(NewJudgeStageCheckStep(), NewJudgeCardJudgeStep(), NewJudgeCardEndStep(), NewJudgeStageExecuteStep())}
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
//****************非bot*********************

type PlayStage struct { // 非bot专用
	*BaseStage
	Buttons []*Button // 主动技能还是画到武将身上吧，这里只有「出牌」与「取消」
}

// 刚回来一定是阶段 0
func (p *PlayStage) TopStage(player *Player, extra *StageExtra) {
	player.ResetCard()
	player.CheckCard(p.Extra)
}

// 绘制区域 240 ~ 1200-240 y底部是280*2 绘制「出牌」「取消」
func (p *PlayStage) InitStage(player *Player, extra *StageExtra) {
	p.Buttons = []*Button{NewButton(TextPlayCard), NewButton(TextCancel)}
	last := float32(WinWidth - 240 - 240)
	for _, button := range p.Buttons {
		last -= button.W
	}
	offset := last / float32(len(p.Buttons)+1) // 计算偏移 若是技能太多，offset会变成负数，按钮会叠在一起
	x := 240 + offset
	y := 280 + 280 - p.Buttons[0].H - 20
	for _, button := range p.Buttons {
		button.X, button.Y = x, y
		x += offset + button.W
	}
	player.ResetCard()
	player.CheckCard(p.Extra)
}

func (p *PlayStage) DrawStage(screen *ebiten.Image, player *Player, extra *StageExtra) {
	for _, button := range p.Buttons {
		button.Draw(screen)
	}
}

func (p *PlayStage) GetStage() StageType {
	return StagePlay
}

func NewPlayStage() *PlayStage {
	res := &PlayStage{}
	base := NewBaseStage(NewPlayStageCardStep(res), NewPlayStagePlayerStep(res))
	res.BaseStage = base
	return res
}

//******************bot专用******************

type BotPlayStage struct { // bot专用
	*BaseStage
}

func (p *BotPlayStage) GetStage() StageType {
	return StagePlay
}

func NewBotPlayStage() *BotPlayStage {
	return &BotPlayStage{BaseStage: NewBaseStage(NewBotPlayStageStep())}
}

//====================DiscardStage弃牌阶段======================
//********************非bot专用**********************

type DiscardStage struct {
	*BaseStage
	Buttons []*Button
}

func (d *DiscardStage) InitStage(player *Player, extra *StageExtra) {
	// 只有「确定」与「取消」
	d.Buttons = []*Button{NewButton(TextConfirm), NewButton(TextCancel)}
	last := float32(WinWidth - 240 - 240)
	for _, button := range d.Buttons {
		last -= button.W
	}
	offset := last / float32(len(d.Buttons)+1) // 计算偏移
	x := 240 + offset
	y := 280 + 280 - d.Buttons[0].H - 20
	for _, button := range d.Buttons {
		button.X, button.Y = x, y
		x += offset + button.W
	}
}

func (d *DiscardStage) DrawStage(screen *ebiten.Image, player *Player, extra *StageExtra) {
	for _, button := range d.Buttons {
		button.Draw(screen)
	}
}

func (d *DiscardStage) GetStage() StageType {
	return StageDiscard
}

func NewDiscardStage() *DiscardStage {
	res := &DiscardStage{}
	base := NewBaseStage(NewDiscardStageCheckStep(), NewDiscardStageMainStep(res))
	res.BaseStage = base
	return res
}

//**********************bot专用************************

type BotDiscardStage struct { // bot专用
	*BaseStage
}

func (p *BotDiscardStage) GetStage() StageType {
	return StagePlay
}

func NewBotDiscardStage() *BotDiscardStage {
	return &BotDiscardStage{BaseStage: NewBaseStage(NewBotDiscardStageStep())}
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
