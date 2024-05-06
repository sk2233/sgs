/*
@author: sk
@date: 2024/5/1
*/
package main

type ISkill interface {
	GetTag() SkillTag
	GetName() string
	CreateEffect(event *Event) IEffect         // 产生效果
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

func (h *SkillHolder) CreateEffects(event *Event) []IEffect {
	res := make([]IEffect, 0)
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

func (b *BaseSkill) CreateEffect(event *Event) IEffect {
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

func (s *SysGameStartSkill) CreateEffect(event *Event) IEffect {
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
	condition.Dist += Min(offset, len(MainGame.Players)-offset) // 不能直接覆盖，可能还是做成类似中间件的处理逻辑好一点
	return true
}

//===========================SysRespCardSkill玩家响应卡牌技能必须排在sys保证最后触发============================

type SysRespCardSkill struct {
	*BaseSkill
}

func NewSysRespCardSkill() *SysRespCardSkill {
	return &SysRespCardSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (b *SysRespCardSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventRespCard {
		return nil
	}
	if event.Desc.IsBot {
		return NewEffect(NewBotRespCardStep())
	} else {
		res := NewEffectWithUI()
		res.SetSteps(NewPlayerRespCardStep(res))
		return res
	}
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
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionCardMaxDesc, Src: src, Card: card, MaxDesc: 1})
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
	if src == desc {
		return false
	}
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseSha, Src: src, Desc: desc,
		Card: card, Dist: 1}) // 这里的Dist是攻击（杀）距离
	if condition.Invalid {
		return false
	} // 这里的Dist是位置距离算上马
	distCond := MainGame.ComputeCondition(&Condition{Type: ConditionGetDist, Src: src, Desc: desc})
	return condition.Dist >= distCond.Dist
}

func (b *ShaSkill) CreateEffect(event *Event) IEffect {
	if len(event.Descs) == 0 {
		MainGame.AddTip("[杀]至少指定一名角色")
		return nil
	}
	return NewEffect(NewUseShaLoopStep(), NewRespShaCardStep(), NewShaHitCheckStep(), NewRespShaExecuteStep())
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

//============================TaoSkill============================

type TaoSkill struct {
	*BaseCheckSkill
}

func NewTaoSkill() *TaoSkill { // 主动只能指定自己使用
	return &TaoSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

func (b *TaoSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	if condition.Invalid {
		return false
	}
	return src.Hp < src.MaxHp // 残血才能用
}

func (b *TaoSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewTaoMainExecuteStep())
}

//======================EquipSkill=========================

type EquipSkill struct {
	*BaseCheckSkill
}

func NewEquipSkill() *EquipSkill { // 装备牌上的技能只是为了上装备，没有其他用处
	return &EquipSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

func (b *EquipSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card, MaxUseSha: 1})
	return !condition.Invalid
}

func (b *EquipSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewEquipMainExecuteStep())
}

//=======================EquipAttackHorseSkill======================

type EquipAttackHorseSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipAttackHorseSkill(player *Player) *EquipAttackHorseSkill {
	return &EquipAttackHorseSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *EquipAttackHorseSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionGetDist || condition.Src != s.Player {
		return false
	}
	condition.Dist--
	return false
}

//=======================EquipDefenseHorseSkill======================

type EquipDefenseHorseSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipDefenseHorseSkill(player *Player) *EquipDefenseHorseSkill {
	return &EquipDefenseHorseSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *EquipDefenseHorseSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionGetDist || condition.Desc != s.Player {
		return false
	}
	condition.Dist++
	return false
}

//=========================EquipZhuGeLianNuSkill=========================

type EquipZhuGeLianNuSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipZhuGeLianNuSkill(player *Player) *EquipZhuGeLianNuSkill {
	return &EquipZhuGeLianNuSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *EquipZhuGeLianNuSkill) HandleCondition(condition *Condition) bool {
	// 自己使用卡牌「杀」时生效
	if condition.Type != ConditionUseCard || condition.Src != s.Player || condition.Card.Name != "杀" {
		return false
	}
	condition.MaxUseSha = 999 // 也不是没限制，只是限制比较宽松
	return false
}

//=========================EquipShaDistSkill针对杀的距离条件==========================

type EquipShaDistSkill struct {
	*BaseSkill
	Player *Player
	Dist   int
}

func NewEquipShaDistSkill(player *Player, dist int) *EquipShaDistSkill {
	return &EquipShaDistSkill{Player: player, Dist: dist, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *EquipShaDistSkill) HandleCondition(condition *Condition) bool {
	// 自己使用卡牌「杀」时生效
	if condition.Type != ConditionUseSha || condition.Src != s.Player {
		return false
	}
	condition.Dist += s.Dist - 1 // 原来就有 1 的距离
	return false
}

//=========================EquipZhangBaSheMaoRespSkill==============================

type EquipZhangBaSheMaoRespSkill struct {
	*BaseSkill
	Player *Player
	Card   *CardWrap // 做为一个简单的示例
}

func NewEquipZhangBaSheMaoRespSkill(player *Player) *EquipZhangBaSheMaoRespSkill {
	return &EquipZhangBaSheMaoRespSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, ""),
		Card: NewTransCardWrap(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic}, []*Card{{}, {}})}
}

// 需要响应出「杀」时触发
func (e *EquipZhangBaSheMaoRespSkill) CreateEffect(event *Event) IEffect {
	// 要求自己响应的牌，可以通过检查，这里会以最差的作为例子，暂时bot不要使用这个技能，收益太不稳定了
	if event.Type != EventRespCard || event.Desc != e.Player || !event.WrapFilter(e.Card) || e.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(func(_ *Card) bool {
		return true // 随便来两张都行，装备牌也行，可能会有「飞刀」问题
	}, 2, true, e.Player, res), NewZhangBaSheMaoRespStep())
	return res
}

//========================EquipZhangBaSheMaoActiveSkill===============================

type EquipZhangBaSheMaoActiveSkill struct {
	*BaseSkill
	Player   *Player
	ShaSkill *ShaSkill
	Card     *Card // 做为一个简单的示例
}

func NewEquipZhangBaSheMaoActiveSkill(player *Player) *EquipZhangBaSheMaoActiveSkill {
	return &EquipZhangBaSheMaoActiveSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "丈八"),
		ShaSkill: NewShaSkill(), Card: &Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic}}
}

func (e *EquipZhangBaSheMaoActiveSkill) CreateEffect(event *Event) IEffect {
	if !e.ShaSkill.CheckUse(event.Src, e.Card, event.StepExtra) { // bot不用主动技能
		return nil // 不能使用杀
	}
	// TODO 跟主动技能做一起
	return nil
}

//=====================EquipGuanShiFuSkill========================

type EquipGuanShiFuSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipGuanShiFuSkill(player *Player) *EquipGuanShiFuSkill {
	return &EquipGuanShiFuSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipGuanShiFuSkill) CreateEffect(event *Event) IEffect {
	// 自己的杀被响应后
	if event.Type != EventRespCardAfter || event.Desc != e.Player || event.Card.Desc.Name != "杀" {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(func(_ *Card) bool {
		return true // 随便来两张都行，装备牌也行，有可能飞刀
	}, 2, true, e.Player, res), NewGuanShiFuCheckStep())
	return res
}

//======================EquipFangTianHuaJiSkill===================

type EquipFangTianHuaJiSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipFangTianHuaJiSkill(player *Player) *EquipFangTianHuaJiSkill {
	return &EquipFangTianHuaJiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipFangTianHuaJiSkill) HandleCondition(condition *Condition) bool {
	// 必须是玩家打出的最后一张牌且为杀
	if condition.Type != ConditionCardMaxDesc || condition.Src != e.Player || condition.Card.Name != "杀" ||
		len(e.Player.Cards) != 1 {
		return false
	}
	condition.MaxDesc = 3
	return true
}

//=======================EquipQingHongJianSkill==========================

type EquipQingHongJianSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipQingHongJianSkill(player *Player) *EquipQingHongJianSkill {
	return &EquipQingHongJianSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipQingHongJianSkill) CreateEffect(event *Event) IEffect {
	// 当用户的「杀」指定目标，或结算后
	if event.Type != EventCardPoint && event.Type != EventCardAfter {
		return nil
	}
	if event.Src != e.Player || event.Card.Desc.Name != "杀" {
		return nil
	}
	return NewEffect(NewQingHongJianStep())
}

//=======================EquipQiLinGongSkill========================

type EquipQiLinGongSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipQiLinGongSkill(player *Player) *EquipQiLinGongSkill {
	return &EquipQiLinGongSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipQiLinGongSkill) CreateEffect(event *Event) IEffect {
	// 用户使用「杀」对其他角色造成伤害后
	if event.Type != EventShaHit || event.Src != e.Player || event.Card.Desc.Name != "杀" {
		return nil
	}
	equips := event.Desc.Equips
	cards := make([]*Card, 0)
	if temp, ok := equips[EquipAttack]; ok {
		cards = append(cards, temp.Card)
	}
	if temp, ok := equips[EquipDefense]; ok {
		cards = append(cards, temp.Card)
	}
	if len(cards) == 0 { // 没有马跳过
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerCardStep(1, res, NewAllCard(nil, cards, nil)), NewQiLinGongExecuteStep())
	return res
}

//================EquipCiXiongShuangGuJianSkill==============

type EquipCiXiongShuangGuJianSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipCiXiongShuangGuJianSkill(player *Player) *EquipCiXiongShuangGuJianSkill {
	return &EquipCiXiongShuangGuJianSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipCiXiongShuangGuJianSkill) CreateEffect(event *Event) IEffect {
	// 必须是武器持有玩家，使用杀指定异性角色为目标时
	if event.Type != EventCardPoint || event.Src != e.Player || event.Card.Desc.Name != "杀" ||
		event.Src.General.Gender == event.Desc.General.Gender {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewCiXiongShuangGuJianReqStep(),
		NewCiXiongShuangGuJianCheckSkill())
	return res
}

//======================BotAskCardSkill bot被要牌时=======================

type BotAskCardSkill struct {
	*BaseSkill
	Player *Player
}

func NewBotAskCardSkill(player *Player) *BotAskCardSkill {
	return &BotAskCardSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *BotAskCardSkill) CreateEffect(event *Event) IEffect {
	// 冲自己要牌
	if event.Type != EventAskCard || event.Desc != e.Player {
		return nil
	}
	return NewEffect(NewBotAskCardStep())
}

//=======================PlayerAskCardSkill===========================

type PlayerAskCardSkill struct {
	*BaseSkill
	Player *Player
}

func NewPlayerAskCardSkill(player *Player) *PlayerAskCardSkill {
	return &PlayerAskCardSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *PlayerAskCardSkill) CreateEffect(event *Event) IEffect {
	// 冲自己要牌
	if event.Type != EventAskCard || event.Desc != e.Player {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(event.Filter, event.AskNum, event.WithEquip, e.Player, res), NewPlayerAskCardStep())
	return res
}

//=====================EquipQingLongYanYueDaoSkill======================

type EquipQingLongYanYueDaoSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipQingLongYanYueDaoSkill(player *Player) *EquipQingLongYanYueDaoSkill {
	return &EquipQingLongYanYueDaoSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipQingLongYanYueDaoSkill) CreateEffect(event *Event) IEffect {
	// 自己的杀被响应后
	if event.Type != EventRespCardAfter || event.Desc != e.Player || event.Card.Desc.Name != "杀" {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(func(card *Card) bool {
		return card.Name == "杀"
	}, 1, false, e.Player, res), NewQingLongYanYueDaoCheckStep())
	return res
}

//=======================EquipHanBingJianSkill=====================

type EquipHanBingJianSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipHanBingJianSkill(player *Player) *EquipHanBingJianSkill {
	return &EquipHanBingJianSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipHanBingJianSkill) CreateEffect(event *Event) IEffect {
	// 必须是用户打的杀命中了
	if event.Type != EventShaHit || event.Src != e.Player || event.Card.Desc.Name != "杀" {
		return nil
	}
	cards := Map(event.Desc.Cards, func(item *CardUI) *Card {
		return item.Card
	})
	equips := make([]*Card, 0)
	for _, equip := range event.Desc.Equips {
		equips = append(equips, equip.Card)
	}
	// TODO 延时锦囊牌
	num := Min(2, len(cards)+len(equips)) // 最多扔两张
	if num == 0 {                         // 啥都没有，没法发动
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerCardStep(num, res, NewAllCard(cards, equips, nil)), NewHanBingJianCheckStep())
	return res
}

//======================EquipBaGuaZhenSkill========================

type EquipBaGuaZhenSkill struct {
	*BaseSkill
	Player *Player
	Card   *CardWrap
}

func NewEquipBaGuaZhenSkill(player *Player) *EquipBaGuaZhenSkill {
	return &EquipBaGuaZhenSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, ""),
		Card: NewVirtualCardWrap(&Card{Name: "闪", Point: PointNone, Suit: SuitNone, Type: CardBasic})}
}

func (e *EquipBaGuaZhenSkill) CreateEffect(event *Event) IEffect {
	// 必须使用需要用户打出「闪」的情况，必须能接受虚拟闪
	if event.Type != EventRespCard || event.Desc != e.Player || !event.WrapFilter(e.Card) {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewJudgeCardJudgeStep(), NewJudgeCardEndStep(), NewBaGuaZhenCheckStep())
	return res
}

//====================EquipRenWangDunSkill======================

type EquipRenWangDunSkill struct {
	*BaseSkill
	Player *Player
}

func NewEquipRenWangDunSkill(player *Player) *EquipRenWangDunSkill {
	return &EquipRenWangDunSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (e *EquipRenWangDunSkill) CreateEffect(event *Event) IEffect {
	// 必须是针对自己的杀
	if event.Type != EventRespCard || event.Desc != e.Player || event.Card.Desc.Name != "杀" {
		return nil
	}
	return NewEffect(NewRenWangDunCheckStep()) // 锁定技 不需要询问用户
}
