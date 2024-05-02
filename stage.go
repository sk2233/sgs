/*
@author: sk
@date: 2024/5/2
*/
package main

import "fmt"

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

//===================PrepareStage准备阶段=====================

type PrepareStage struct {
}

func (p *PrepareStage) GetStage() StageType {
	return StagePrepare
}

func NewPrepareStage() *PrepareStage {
	return &PrepareStage{}
}

func (p *PrepareStage) Update(player *Player, extra *StageExtra) bool {
	// 翻面也作为一种技能效果
	fmt.Println(player.General.Name + " PrepareStage")
	return true
}

//=====================JudgeStage判定阶段=====================

type JudgeStage struct {
	Steps []IStep    // 一个阶段有多个步骤组成
	Extra *StepExtra // 多个步骤存储的中间变量
}

func (j *JudgeStage) GetStage() StageType {
	return StageJudge
}

func NewJudgeStage() *JudgeStage {
	return &JudgeStage{}
}

func (j *JudgeStage) Update(player *Player, extra *StageExtra) bool {
	// 也需要一步步来，但不不能复用Step
	fmt.Println(player.General.Name + " JudgeStage")
	return true
}

//===================DrawStage摸牌阶段====================

type DrawStage struct {
}

func (d *DrawStage) GetStage() StageType {
	return StageDraw
}

func NewDrawStage() *DrawStage {
	return &DrawStage{}
}

func (d *DrawStage) Update(player *Player, extra *StageExtra) bool {
	fmt.Println(player.General.Name + " DrawStage")
	return true
}

//====================PlayStage出牌阶段======================

type PlayStage struct {
}

func (p *PlayStage) GetStage() StageType {
	return StagePlay
}

func NewPlayStage() *PlayStage {
	return &PlayStage{}
}

func (p *PlayStage) Update(player *Player, extra *StageExtra) bool {
	fmt.Println(player.General.Name + " PlayStage")
	return true
}

//====================DiscardStage弃牌阶段======================

type DiscardStage struct {
}

func (d *DiscardStage) GetStage() StageType {
	return StageDiscard
}

func NewDiscardStage() *DiscardStage {
	return &DiscardStage{}
}

func (d *DiscardStage) Update(player *Player, extra *StageExtra) bool {
	fmt.Println(player.General.Name + " DiscardStage")
	return true
}

//====================EndStage回合结束阶段========================

type EndStage struct {
}

func (e *EndStage) GetStage() StageType {
	return StageEnd
}

func NewEndStage() *EndStage {
	return &EndStage{}
}

func (e *EndStage) Update(player *Player, extra *StageExtra) bool {
	fmt.Println(player.General.Name + " EndStage")
	return true
}
