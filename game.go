/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

var MainGame *Game

type Game struct {
	Players     []*Player
	Index       int
	ActionStack *Stack[IAction]
	SkillHolder *SkillHolder
	CardManager *CardManager
}

func NewGame() *Game {
	players := LoadPlayer()
	stack := NewStack[IAction]()
	stack.Push(NewGamePrepareAction(players))
	MainGame = &Game{
		Players:     players,
		ActionStack: stack,
		SkillHolder: NewSkillHolder(NewSysInitCardSkill(), NewSysGameStartSkill()),
		CardManager: NewCardManager(),
	}
	return MainGame
}

// 每帧执行的逻辑，error我没用过
func (g *Game) Update() error {
	g.ActionStack.Peek().Update() // 栈中不能为空，采用对象手动弹栈方式
	return nil
}

// 每帧绘制的画面
func (g *Game) Draw(screen *ebiten.Image) {
	for _, player := range g.Players {
		player.Draw(screen)
	}
}

// 设置画布的大小，入参窗口大小，返回画布大小
func (g *Game) Layout(w, h int) (int, int) {
	return w, h
}

func (g *Game) ComputeCondition(condition *Condition) *Condition {
	holders := g.GetAllSortSkillHolder(condition.Src)
	for _, holder := range holders {
		if holder.HandleCondition(condition) {
			break
		}
	}
	return condition
}

func (g *Game) TriggerEvent(event *Event) {
	holders := g.GetAllSortSkillHolder(event.Src)
	effects := make([]*Effect, 0)
	for _, holder := range holders {
		effects = append(effects, holder.CreateEffects(event)...)
	}
	if len(effects) > 0 { // 存在要被触发的事件进行触发
		g.PushAction(NewEffectGroup(event, effects))
	}
}

func (g *Game) GetAllSortSkillHolder(src *Player) []*SkillHolder {
	players := g.Players
	if src != nil { // 若是存在事件源玩家，就已他为起点逆时针结算
		for i := 0; i < len(players); i++ {
			if players[i] == src {
				players = append(players[i:], players[:i]...)
				break
			}
		}
	}
	res := make([]*SkillHolder, 0)
	for _, player := range players { // 玩家技能最大
		res = append(res, player.SkillHolder)
	}
	for _, player := range players { // 装备技能次之
		res = append(res, player.GetEquipSkillHolders()...)
	}
	res = append(res, g.SkillHolder) // 系统规则最小
	return res
}

func (g *Game) NextPlayer() {
	player := g.Players[g.Index]
	g.Index = (g.Index + 1) % len(g.Players)
	g.PushAction(NewPlayerStageAction(player))
}

func (g *Game) PushAction(action IAction) {
	g.ActionStack.Push(action)
}

func (g *Game) PopAction() {
	g.ActionStack.Pop()
}

func (g *Game) DrawCard(num int) []*Card {
	return g.CardManager.DrawCard(num)
}

func (g *Game) DiscardCard(cards []*Card) {
	g.CardManager.DiscardCard(cards)
}
