/*
@author: sk
@date: 2024/5/1
*/
package main

type ISkill interface {
	GetTag() SkillTag
	GetName() string
	CreateEffect(event *Event) *Effect
	HandleCondition(condition *Condition) bool // 返回是否终止
}

type SkillHolder struct {
	Skills []ISkill // 最好可能还是需要一下排序，暂时不排序
}

func (h *SkillHolder) HandleCondition(condition *Condition) bool { // 返回是否终止
	for _, skill := range h.Skills {
		if skill.HandleCondition(condition) {
			return true
		}
	}
	return false
}

func (h *SkillHolder) CreateEffects(event *Event) []*Effect {
	res := make([]*Effect, 0)
	for _, skill := range h.Skills {
		if effect := skill.CreateEffect(event); effect != nil {
			res = append(res, effect)
		}
	}
	return res
}

func NewSkillHolder(skills ...ISkill) *SkillHolder {
	return &SkillHolder{Skills: skills}
}

type BaseSkill struct {
	Tag  SkillTag
	Name string
}

func NewBaseSkill(tag SkillTag, name string) *BaseSkill {
	return &BaseSkill{Tag: tag, Name: name}
}

func (b *BaseSkill) GetTag() SkillTag {
	return b.Tag
}

func (b *BaseSkill) GetName() string {
	return b.Name
}

func (b *BaseSkill) CreateEffect(event *Event) *Effect {
	return nil
}

func (b *BaseSkill) HandleCondition(condition *Condition) bool {
	return false
}

//=====================SysInitCardSkill系统初始化玩家手牌技能=======================

type SysInitCardSkill struct {
	*BaseSkill
}

func NewSysInitCardSkill() *SysInitCardSkill {
	return &SysInitCardSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysInitCardSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionInitCard {
		return false
	}
	condition.CardNum = 4 // 默认初始 4 张手牌
	return true
}

//==========================SysGameStartSkill系统游戏开始技能=================================

type SysGameStartSkill struct {
	*BaseSkill
}

func NewSysGameStartSkill() *SysGameStartSkill {
	return &SysGameStartSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysGameStartSkill) CreateEffect(event *Event) *Effect {
	if event.Type == EventGameStart {
		return NewEffect(NewSysNextPlayerStep())
	}
	return nil
}
