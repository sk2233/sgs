/*
@author: sk
@date: 2024/5/1
*/
package main

type IAction interface { // 用于统一 效果组与游戏刚开始的一些简单效果
	Update() bool
}

type OnceAction struct {
	Over   bool
	Action func()
}

func (o *OnceAction) Update() bool {
	if o.Over {
		return true
	}
	o.Action() // 对于执行中可能把新事件压入栈顶行为不能立即返回true进行弹栈，可能把新压入的弹出，必须保证改流程中堆栈没有变化才能弹栈
	o.Over = true
	return false
}

func NewOnceAction() *OnceAction {
	return &OnceAction{Over: false}
}

//====================GamePrepareAction第一个Action=====================

type GamePrepareAction struct {
	*OnceAction
	Players []*Player
}

func (g *GamePrepareAction) GamePrepare() {
	// 主要是触发所有人获取初始手牌，这个阶段不会触发任何事件，但是会计算Condition
	for _, player := range g.Players {
		condition := MainGame.ComputeCondition(&Condition{Type: ConditionInitCard, Desc: player})
		player.DrawCard(condition.CardNum)
	} // 最终触发游戏开始事件
	MainGame.TriggerEvent(&Event{Type: EventGameStart})
}

func NewGamePrepareAction(players []*Player) *GamePrepareAction {
	res := &GamePrepareAction{OnceAction: NewOnceAction(), Players: players}
	res.Action = res.GamePrepare
	return res
}
