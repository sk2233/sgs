/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

var MainGame *Game

type Game struct {
	Players     []*Player
	Index       int
	ActionStack *Stack[IAction]
	SkillHolder *SkillHolder
	CardManager *CardManager
	TipManager  *TipManager
	Desktop     *Desktop
}

func NewGame() *Game {
	players := LoadPlayer()
	stack := NewStack[IAction]()
	stack.Push(NewGamePrepareAction(players))
	MainGame = &Game{
		Players:     players,
		ActionStack: stack,
		SkillHolder: NewSkillHolder(NewSysInitCardSkill(), NewSysGameStartSkill(),
			NewSysDrawCardSkill(), NewSysMaxCardSkill(), NewSysPlayerDistSkill(), NewSysRespCardSkill()),
		CardManager: NewCardManager(),
		TipManager:  NewTipManager(),
		Desktop:     NewDesktop(),
	}
	return MainGame
}

// 每帧执行的逻辑，error我没用过
func (g *Game) Update() error {
	g.ActionStack.Peek().Update() // 栈中不能为空，采用对象手动弹栈方式
	g.TipManager.Update()
	return nil
}

// 每帧绘制的画面
func (g *Game) Draw(screen *ebiten.Image) {
	var player *Player
	for _, item := range g.Players { // Players是出牌顺序
		if item.IsBot {
			item.Draw(screen)
		} else {
			player = item
		}
	}
	if player != nil { // 玩家留到最后绘制，防止被其他 bot 遮挡
		player.Draw(screen)
	}
	g.Desktop.Draw(screen)
	g.CardManager.Draw(screen)
	InvokeDraw(g.ActionStack.Peek(), screen)
	g.TipManager.Draw(screen) // 提示优先级最高
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
	effects := make([]IEffect, 0)
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
	if g.ActionStack.Len() > 0 {
		InvokeTop(g.ActionStack.Peek())
	}
}

func (g *Game) DrawCard(num int) []*Card {
	return g.CardManager.DrawCard(num)
}

func (g *Game) DiscardCard(cards []*Card) {
	g.CardManager.DiscardCard(cards)
}

func (g *Game) TogglePlayer(x, y float32) bool {
	for _, player := range g.Players {
		if player.ToggleSelect(x, y) {
			return true
		}
	}
	return false
}

func (g *Game) ResetPlayer() {
	for _, player := range g.Players {
		player.CanSelect = true
		player.Select = false
	}
}

func (g *Game) AddToDesktop(cards ...*Card) {
	g.Desktop.AddCard(Map(cards, NewSimpleCardWrap))
}

func (g *Game) AddToDesktopRaw(cards ...*CardWrap) {
	g.Desktop.AddCard(cards)
}

func (g *Game) DiscardFromDesktop(cards ...*Card) {
	g.Desktop.DiscardCard(Map(cards, NewSimpleCardWrap))
}

func (g *Game) DiscardFromDesktopRaw(cards ...*CardWrap) {
	g.Desktop.DiscardCard(cards)
}

func (g *Game) RemoveFromDesktop(cards ...*Card) {
	g.Desktop.RemoveCard(Map(cards, NewSimpleCardWrap))
}

func (g *Game) RemoveFromDesktopRaw(cards ...*CardWrap) {
	g.Desktop.RemoveCard(cards)
}

func (g *Game) AddTip(format string, args ...any) {
	g.TipManager.AddTip(fmt.Sprintf(format, args...))
}

func (g *Game) GetSelectPlayer() []*Player {
	return Filter(g.Players, func(player *Player) bool {
		return player.Select
	})
}

func (g *Game) CheckPlayer(src *Player, card *Card) {
	for _, desc := range g.Players {
		desc.CanSelect = card.Skill.CheckTarget(src, desc, card)
	}
}

func (g *Game) GetPlayerIndex(player *Player) int {
	for i := 0; i < len(g.Players); i++ {
		if g.Players[i] == player {
			return i
		}
	}
	return -1
}
