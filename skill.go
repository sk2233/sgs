/*
@author: sk
@date: 2024/5/1
*/
package main

type ISkill interface {
	GetTag() SkillTag
	GetName() string
	CreateEffect(event *Event) *Effect         // 产生效果
	HandleCondition(condition *Condition) bool // 返回是否终止  计算中间数据
}

type ICheckSkill interface { // 就是给手牌正常使用用的，先不考虑强制转换的效果，那个是在获取手牌时动的手脚
	ISkill

	CheckUse(src *Player, card *Card, extra *StepExtra) bool // 简单判断是否可以使用
	GetMaxDesc(src *Player, card *Card) int                  // 返回最多可以指定的目标数目
	CheckTarget(src, desc *Player, card *Card) bool          // 判断是否可以指定目标
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

//=============================SysDrawCardSkill系统 玩家摸牌阶段摸牌数===================================

type SysDrawCardSkill struct {
	*BaseSkill
}

func NewSysDrawCardSkill() *SysDrawCardSkill {
	return &SysDrawCardSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysDrawCardSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionDrawCardNum {
		return false
	}
	condition.CardNum = 2 // 默认摸 2 张
	return true
}

//=============================SysMaxCardSkill系统 玩家最多保留多少牌===================================

type SysMaxCardSkill struct {
	*BaseSkill
}

func NewSysMaxCardSkill() *SysMaxCardSkill {
	return &SysMaxCardSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysMaxCardSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionMaxCard {
		return false
	}
	condition.MaxCard = condition.Src.Hp // 默认最多保留当前体力张牌
	return true
}

//================================SysPlayerDistSkill系统默认计算两个用户之间的距离==================================

type SysPlayerDistSkill struct {
	*BaseSkill
}

func NewSysPlayerDistSkill() *SysPlayerDistSkill {
	return &SysPlayerDistSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysPlayerDistSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionGetDist {
		return false
	}
	srcIndex := MainGame.GetPlayerIndex(condition.Src)
	descIndex := MainGame.GetPlayerIndex(condition.Desc)
	offset := Abs(srcIndex - descIndex)
	condition.Dist = Min(offset, len(MainGame.Players)-offset)
	return true
}

//----卡牌效果不处理计算效果，只处理CreateEffect，若是CreateEffect返回nil表示无法处理------

type BaseCheckSkill struct {
	*BaseSkill
}

func NewBaseCheckSkill() *BaseCheckSkill {
	return &BaseCheckSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (b *BaseCheckSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	return false
}

func (b *BaseCheckSkill) GetMaxDesc(src *Player, card *Card) int {
	return 0
}

func (b *BaseCheckSkill) CheckTarget(src, desc *Player, card *Card) bool {
	return false
}

//=============================ShaSkill===========================

type ShaSkill struct {
	*BaseCheckSkill
}

func (b *ShaSkill) GetMaxDesc(src *Player, card *Card) int {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card, MaxDesc: 1})
	return condition.MaxDesc // 默认杀只能选择一个目标
}

func (b *ShaSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card, MaxUseSha: 1})
	if condition.Invalid {
		return false
	}
	return extra.ShaCount < condition.MaxUseSha // 不得大于自己最大出杀数，默认最大出杀数为1
}

func (b *ShaSkill) CheckTarget(src, desc *Player, card *Card) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseSha, Src: src, Desc: desc,
		Card: card, Dist: 1}) // 这里的Dist是攻击（杀）距离
	if condition.Invalid {
		return false
	} // 这里的Dist是位置距离算上马
	distCond := MainGame.ComputeCondition(&Condition{Type: ConditionGetDist, Src: src, Desc: desc})
	return condition.Dist >= distCond.Dist
}

func (b *ShaSkill) CreateEffect(event *Event) *Effect {
	// TODO
	return nil
}

func NewShaSkill() *ShaSkill {
	return &ShaSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//=======================ShanSkill=========================

type ShanSkill struct {
	*BaseCheckSkill
}

func NewShanSkill() *ShanSkill { // 闪是少数只能用来响应的牌，CreateEffect一直返回nil 不用技能完全打出不去
	return &ShanSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//============================PeachSkill============================

type PeachSkill struct {
	*BaseCheckSkill
}

func NewPeachSkill() *PeachSkill { // 主动只能指定自己使用
	return &PeachSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

func (b *PeachSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	if condition.Invalid {
		return false
	}
	return src.Hp < src.MaxHp // 残血才能用
}

func (b *PeachSkill) CreateEffect(event *Event) *Effect {
	// TODO
	return nil
}
