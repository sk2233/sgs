/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type IEffect interface {
	Update(event *Event) bool
}

type Effect struct { // 一个事件触发的一个效果
	Steps []IStep    // 一个效果有多个步骤组成
	Extra *StepExtra // 多个步骤存储的中间变量
}

func NewEffect(steps ...IStep) *Effect {
	return &Effect{Steps: steps, Extra: NewStepExtra()}
}

func (e *Effect) Update(event *Event) bool { // 执行效果并返回是否结束
	if e.Extra.Index < len(e.Steps) {
		e.Steps[e.Extra.Index].Update(event, e.Extra) // Index 是由Step内部控制的，可能向前调整，也可能向后调整
		return false
	}
	return true
}

type EffectWithUI struct { // 有绘图能力的 Effect 一些效果需要用户做选择
	*Effect
	UIs []IDraw
}

func (e *EffectWithUI) DrawEffect(screen *ebiten.Image, event *Event) {
	for _, ui := range e.UIs {
		ui.Draw(screen)
	}
}

func (e *EffectWithUI) SetSteps(steps ...IStep) { // EffectWithUI比较特殊，最好等创建完实例再设置IStep
	e.Effect = NewEffect(steps...)
}

func NewEffectWithUI() *EffectWithUI {
	return &EffectWithUI{UIs: make([]IDraw, 0)}
}

type EffectGroup struct {
	Event   *Event
	Effects []IEffect // 一个事件触发的所有效果
	Index   int       // 事件进行到那个了
}

func (e *EffectGroup) Draw(screen *ebiten.Image) {
	if e.Index < len(e.Effects) {
		InvokeDrawEffect(e.Effects[e.Index], screen, e.Event)
	}
}

func NewEffectGroup(event *Event, effects []IEffect) *EffectGroup {
	return &EffectGroup{Event: event, Effects: effects, Index: 0}
}

func (e *EffectGroup) Update() { // 执行效果组并返回是否执行结束
	if e.Event.Abort { // 可能因为某个效果的发动，后续效果被中断了，不能在改变后立即弹栈，可能弹错
		MainGame.PopAction()
		return
	}
	if e.Index < len(e.Effects) {
		if e.Effects[e.Index].Update(e.Event) { // 结束了下一个效果
			e.Index++
		}
	} else {
		MainGame.PopAction()
	}
}

func NewEffectGroupBySkill(event *Event, skill ISkill) *EffectGroup {
	effect := skill.CreateEffect(event) // 指定skill时这里不能返回nil，他必须可以处理
	return NewEffectGroup(event, []IEffect{effect})
}
