/*
@author: sk
@date: 2024/5/1
*/
package main

type EffectExtra struct {
}

type Effect struct { // 一个事件触发的一个效果
	Steps []IStep      // 一个效果有多个步骤组成
	Index int          // 步骤进行到那里了
	Extra *EffectExtra // 多个步骤存储的中间变量
}

func NewEffect(steps ...IStep) *Effect {
	return &Effect{Steps: steps, Index: 0, Extra: &EffectExtra{}}
}

func (e *Effect) Update(event *Event) bool { // 执行效果并返回是否结束
	if e.Index < len(e.Steps) {
		e.Steps[e.Index].Update(e, event) // Index 是由Step内部控制的，可能向前调整，也可能向后调整
		return false
	}
	return true
}

type EffectGroup struct {
	Event   *Event
	Effects []*Effect // 一个事件触发的所有效果
	Index   int       // 事件进行到那个了
}

func NewEffectGroup(event *Event, effects []*Effect) *EffectGroup {
	return &EffectGroup{Event: event, Effects: effects, Index: 0}
}

func (e *EffectGroup) Update() bool { // 执行效果组并返回是否执行结束
	if e.Event.Abort { // 可能因为某个效果的发动，后续效果被中断了，不能在改变后立即弹栈，可能弹错
		return true
	}
	if e.Index < len(e.Effects) {
		if e.Effects[e.Index].Update(e.Event) { // 结束了下一个效果
			e.Index++
		}
		return false
	}
	return true
}
