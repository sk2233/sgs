/*
@author: sk
@date: 2024/5/1
*/
package main

import "github.com/hajimehoshi/ebiten/v2"

type IAction interface { // 用于统一 效果组与游戏刚开始的一些简单效果
	Update()
}

//====================GamePrepareAction第一个Action=====================

type GamePrepareAction struct {
	Players []*Player
}

func (g *GamePrepareAction) Update() {
	MainGame.PopAction() // 只触发一次就结束了，先把自己弹出
	// 主要是触发所有人获取初始手牌，这个阶段不会触发任何事件，但是会计算Condition
	for _, player := range g.Players {
		condition := MainGame.ComputeCondition(&Condition{Type: ConditionInitCard, Desc: player})
		player.DrawCard(condition.CardNum)
	} // 最终触发游戏开始事件
	MainGame.TriggerEvent(&Event{Type: EventGameStart})
}

func NewGamePrepareAction(players []*Player) *GamePrepareAction {
	return &GamePrepareAction{Players: players}
}

//========================PlayerStageExtra玩家阶段行为============================
// 这里没有使用事件，因为玩家阶段是普遍稳定的，使用事件触发适用于不稳定的场景，而且若是使用玩家阶段触发另外一个玩家阶段将导致栈越压越多无法释放

type PlayerStageAction struct {
	Player *Player
	Extra  *StageExtra // 回合内共享数据
	Stages []IStage    // 暂时不考虑效果对阶段数组的修改
	Index  int
}

func (p *PlayerStageAction) Top() {
	if p.Index < len(p.Stages) { // 只能监听到从非栈顶，到栈顶的过程
		InvokeTopStage(p.Stages[p.Index], p.Player, p.Extra)
	}
}

func NewPlayerStageAction(player *Player) *PlayerStageAction {
	var playStage, discardStage IStage
	if player.IsBot {
		playStage = NewBotPlayStage()
		discardStage = NewBotDiscardStage()
	} else {
		playStage = NewPlayStage()
		discardStage = NewDiscardStage()
	}
	return &PlayerStageAction{Player: player,
		Stages: []IStage{NewPrepareStage(), NewJudgeStage(), NewDrawStage(), playStage,
			discardStage, NewEndStage()}, Extra: NewStageExtra()}
}

func (p *PlayerStageAction) Update() {
	if p.Index < len(p.Stages) {
		if p.Stages[p.Index].Update(p.Player, p.Extra) {
			p.Index++ // 寻找下一个阶段
			for p.Index < len(p.Stages) && (p.Stages[p.Index].GetStage()&p.Extra.SkipStage) > 0 {
				p.Index++
			}
			if p.Index < len(p.Stages) { // 可能照顾不到第一个阶段「开始阶段」
				InvokeInitStage(p.Stages[p.Index], p.Player, p.Extra)
			}
		}
	} else {
		MainGame.PopAction()  // 弹出玩家阶段行为
		MainGame.NextPlayer() // 寻找下一个合法玩家压栈
	}
}

func (p *PlayerStageAction) Draw(screen *ebiten.Image) {
	if p.Index < len(p.Stages) {
		InvokeDrawStage(p.Stages[p.Index], screen, p.Player, p.Extra)
	}
}
