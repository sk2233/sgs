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
	condition.CardNum += 2 // 默认摸 2 张
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
	players := MainGame.GetPlayers(func(player *Player) bool {
		return !player.IsDie
	})
	srcIndex := s.getIndex(condition.Src, players)
	descIndex := s.getIndex(condition.Desc, players)
	offset := Abs(srcIndex - descIndex)
	condition.Dist += Min(offset, len(MainGame.Players)-offset) // 不能直接覆盖，可能还是做成类似中间件的处理逻辑好一点
	return true
}

func (s *SysPlayerDistSkill) getIndex(player *Player, players []*Player) int {
	for i := 0; i < len(players); i++ {
		if players[i] == player {
			return i
		}
	}
	return -1
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
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionCardPoint, Src: src, Desc: desc, Card: card})
	if condition.Invalid {
		return false
	}
	condition = MainGame.ComputeCondition(&Condition{Type: ConditionUseSha, Src: src, Desc: desc,
		Card: card, Dist: 1}) // 这里的Dist是攻击（杀）距离
	// 这里的Dist是位置距离算上马
	distCond := MainGame.ComputeCondition(&Condition{Type: ConditionGetDist, Card: card, Src: src, Desc: desc})
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
	res.SetSteps(NewSelectNumCardStep(e.AnyFilter, 2, 2, true, e.Player, res), NewZhangBaSheMaoRespStep())
	return res
}

func (e *EquipZhangBaSheMaoRespSkill) AnyFilter(card *Card) bool {
	return true
}

//========================EquipZhangBaSheMaoActiveSkill===============================

type EquipZhangBaSheMaoActiveSkill struct {
	*BaseSkill
	Player *Player
	Card   *Card // 做为一个简单的示例
}

func NewEquipZhangBaSheMaoActiveSkill(player *Player) *EquipZhangBaSheMaoActiveSkill {
	return &EquipZhangBaSheMaoActiveSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "丈八"),
		Card: &Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic, Skill: NewShaSkill()}}
}

func (e *EquipZhangBaSheMaoActiveSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventUseSkill || event.Src != e.Player || e.Player.IsBot {
		return nil
	}
	if !e.Card.Skill.CheckUse(event.Src, e.Card, event.StepExtra) { // 没有使用机会了
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerStep(0, 1, e.ShaFilter, res), NewSelectPlayerVerifyStep(),
		NewSelectNumCardStep(e.AnyFilter, 2, 2, false, e.Player, res),
		NewTransCardStep(e.Card))
	return res
}

func (e *EquipZhangBaSheMaoActiveSkill) ShaFilter(player *Player) bool {
	return e.Card.Skill.CheckTarget(e.Player, player, e.Card)
}

func (e *EquipZhangBaSheMaoActiveSkill) AnyFilter(card *Card) bool {
	return true
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
	res.SetSteps(NewSelectNumCardStep(e.AnyFilter, 2, 2, true, e.Player, res), NewGuanShiFuCheckStep())
	return res
}

func (e *EquipGuanShiFuSkill) AnyFilter(card *Card) bool {
	return true
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
	res.SetSteps(NewSelectNumCardStep(event.Filter, event.AskNum, event.AskNum, event.WithEquip, e.Player, res), NewPlayerAskCardStep())
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
	res.SetSteps(NewSelectNumCardStep(e.ShaFilter, 1, 1, false, e.Player, res), NewQingLongYanYueDaoCheckStep())
	return res
}

func (e *EquipQingLongYanYueDaoSkill) ShaFilter(card *Card) bool {
	return card.Name == "杀"
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
	cards := event.Desc.GetCards()
	equips := event.Desc.GetEquips()
	delayKits := event.Desc.GetDelayKits()
	num := Min(2, len(cards)+len(equips)) // 最多扔两张
	if num == 0 {                         // 啥都没有，没法发动
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerCardStep(num, res, NewAllCard(cards, equips, delayKits)), NewHanBingJianCheckStep())
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
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(), NewJudgeCardJudgeStep(), NewJudgeCardEndStep(), NewBaGuaZhenCheckStep())
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

//====================WuZhongShengYouSkill====================

type WuZhongShengYouSkill struct {
	*BaseCheckSkill
}

func (w *WuZhongShengYouSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewWuZhongShengYouStep())
}

func (w *WuZhongShengYouSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func NewWuZhongShengYouSkill() *WuZhongShengYouSkill {
	return &WuZhongShengYouSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//===================WuXieKeJiSkill====================

type WuXieKeJiSkill struct {
	*BaseCheckSkill
}

func NewWuXieKeJiSkill() *WuXieKeJiSkill {
	return &WuXieKeJiSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

// 只能用于响应，但是还是会有效果存储在这里，放在其他地方也行，放在这里比较方便
func (b *WuXieKeJiSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewWuXieKeJiStep())
}

//===================GuoHeChaiQiaoSkill==================

type GuoHeChaiQiaoSkill struct {
	*BaseCheckSkill
}

func (g *GuoHeChaiQiaoSkill) GetMaxDesc(src *Player, card *Card) int {
	return 1
}

func (g *GuoHeChaiQiaoSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *GuoHeChaiQiaoSkill) CheckTarget(src, desc *Player, card *Card) bool {
	if src == desc { // 不能拆自己的
		return false
	} // 必须有牌可拆
	return len(desc.Cards) > 0 || len(desc.Equips) > 0 || len(desc.DelayKits) > 0
}

func (g *GuoHeChaiQiaoSkill) CreateEffect(event *Event) IEffect {
	if len(event.Descs) != 1 {
		MainGame.AddTip("[过河拆桥]需要指定一名角色")
		return nil
	}
	res := NewEffectWithUI() // 暂时没有考虑多目标选择，若需要参考「万箭齐发」
	cards := event.Descs[0].GetCards()
	equips := event.Descs[0].GetEquips()
	delayKits := event.Descs[0].GetDelayKits()
	res.SetSteps(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewCardCheckInvalidStep(),
		NewSelectPlayerCardStep(1, res, NewAllCard(cards, equips, delayKits)), NewGuoHeChaiQiaoStep())
	return res
}

func NewGuoHeChaiQiaoSkill() *GuoHeChaiQiaoSkill {
	return &GuoHeChaiQiaoSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//====================ShunShouQianYangSkill======================

type ShunShouQianYangSkill struct {
	*BaseCheckSkill
}

func (g *ShunShouQianYangSkill) GetMaxDesc(src *Player, card *Card) int {
	return 1
}

func (g *ShunShouQianYangSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *ShunShouQianYangSkill) CheckTarget(src, desc *Player, card *Card) bool {
	if src == desc { // 不能拿自己的
		return false
	} // 必须有牌可牵羊
	if len(desc.Cards) == 0 && len(desc.Equips) == 0 && len(desc.DelayKits) == 0 {
		return false
	}
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionCardPoint, Src: src, Desc: desc, Card: card})
	if condition.Invalid {
		return false
	}
	// 这里的Dist是位置距离算上马
	distCond := MainGame.ComputeCondition(&Condition{Type: ConditionGetDist, Card: card, Src: src, Desc: desc})
	return distCond.Dist <= 1
}

func (g *ShunShouQianYangSkill) CreateEffect(event *Event) IEffect {
	if len(event.Descs) != 1 {
		MainGame.AddTip("[顺手牵羊]需要指定一名角色")
		return nil
	}
	res := NewEffectWithUI() // 暂时没有考虑多目标选择，若需要参考「万箭齐发」
	cards := event.Descs[0].GetCards()
	equips := event.Descs[0].GetEquips()
	delayKits := event.Descs[0].GetDelayKits()
	res.SetSteps(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewCardCheckInvalidStep(),
		NewSelectPlayerCardStep(1, res, NewAllCard(cards, equips, delayKits)), NewShunShouQianYangStep())
	return res
}

func NewShunShouQianYangSkill() *ShunShouQianYangSkill {
	return &ShunShouQianYangSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//======================JieDaoShaRenSkill=====================

type JieDaoShaRenSkill struct {
	*BaseCheckSkill
	ShaSkill *ShaSkill
	ShaCard  *Card
}

func (g *JieDaoShaRenSkill) GetMaxDesc(src *Player, card *Card) int {
	return 2
}

func (g *JieDaoShaRenSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *JieDaoShaRenSkill) CheckTarget(src, desc *Player, card *Card) bool {
	if len(MainGame.GetSelectPlayer()) == 1 { // 已经选择了一个人，选择的另一个人必须在其攻击范围内
		target := MainGame.GetSelectPlayer()[0]
		return g.ShaSkill.CheckTarget(target, desc, g.ShaCard)
	}
	// 第一个选择选择的人必须有武器且不能是自己
	if src == desc {
		return false
	}
	return desc.Equips[EquipWeapon] != nil
}

func (g *JieDaoShaRenSkill) CreateEffect(event *Event) IEffect {
	if len(event.Descs) != 2 {
		MainGame.AddTip("[借刀杀人]必需指定两名角色")
		return nil
	}
	return NewEffect(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewJieDaoShaRenAskStep(),
		NewJieDaoShaRenCheckStep(), NewCardEndStep())
}

func NewJieDaoShaRenSkill() *JieDaoShaRenSkill {
	return &JieDaoShaRenSkill{BaseCheckSkill: NewBaseCheckSkill(), ShaSkill: NewShaSkill(), ShaCard: &Card{
		Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic,
	}}
}

//=================WuGuFengDengSkill===================

type WuGuFengDengSkill struct {
	*BaseCheckSkill
}

func NewWuGuFengDengSkill() *WuGuFengDengSkill {
	return &WuGuFengDengSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

func (g *WuGuFengDengSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *WuGuFengDengSkill) CreateEffect(event *Event) IEffect {
	event.Cards = MainGame.DrawCard(len(MainGame.Players))
	return NewEffect(NewWuGuFengDengPrepareStep(event.Src), NewLoopTriggerUseKitStep(event.Src),
		NewCheckRespKitStep(), NewWuGuFengDengChooseStep(), NewWuGuFengDengExecuteStep())
}

//================BotChooseCardSkill=====================

type BotChooseCardSkill struct {
	*BaseSkill
	Player *Player
}

func NewBotChooseCardSkill(player *Player) *BotChooseCardSkill {
	return &BotChooseCardSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

func (b *BotChooseCardSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventChooseCard || event.Desc != b.Player {
		return nil
	}
	return NewEffect(NewBotChooseCardStep())
}

//===================PlayerChooseCardSkill======================

type PlayerChooseCardSkill struct {
	*BaseSkill
	Player *Player
}

func (b *PlayerChooseCardSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventChooseCard || event.Desc != b.Player {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewChooseNumCardStep(event.ChooseMin, event.ChooseMax, res, NewChooseCard(event.Cards)), NewPlayerChooseCardStep())
	return res
}

func NewPlayerChooseCardSkill(player *Player) *PlayerChooseCardSkill {
	return &PlayerChooseCardSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

//===================JueDouSkill================

type JueDouSkill struct {
	*BaseCheckSkill
}

func NewJueDouSkill() *JueDouSkill {
	return &JueDouSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

func (g *JueDouSkill) GetMaxDesc(src *Player, card *Card) int {
	return 1
}

func (g *JueDouSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *JueDouSkill) CheckTarget(src, desc *Player, card *Card) bool {
	if src == desc {
		return false
	}
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionCardPoint, Src: src, Desc: desc, Card: card})
	return !condition.Invalid
}

func (g *JueDouSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewJueDouCheckStep(), NewJueDouExecuteStep())
}

//=============NanManRuQinSkill===============

type NanManRuQinSkill struct {
	*BaseCheckSkill
}

func (g *NanManRuQinSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *NanManRuQinSkill) CreateEffect(event *Event) IEffect {
	event.HurtVal = 1
	return NewEffect(NewAoePrepareStep(event.Src), NewLoopTriggerUseKitStep(event.Src),
		NewCheckRespKitStep(), NewAoeRespStep(g.ShaFilter), NewAoeExecuteStep())
}

func (g *NanManRuQinSkill) ShaFilter(card *CardWrap) bool {
	return card.Desc.Name == "杀"
}

func NewNanManRuQinSkill() *NanManRuQinSkill {
	return &NanManRuQinSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//================WanJianQiFaSkill================

type WanJianQiFaSkill struct {
	*BaseCheckSkill
}

func (g *WanJianQiFaSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *WanJianQiFaSkill) CreateEffect(event *Event) IEffect {
	event.HurtVal = 1
	return NewEffect(NewAoePrepareStep(event.Src), NewLoopTriggerUseKitStep(event.Src),
		NewCheckRespKitStep(), NewAoeRespStep(g.ShanFilter), NewAoeExecuteStep())
}

func (g *WanJianQiFaSkill) ShanFilter(card *CardWrap) bool {
	return card.Desc.Name == "闪"
}

func NewWanJianQiFaSkill() *WanJianQiFaSkill {
	return &WanJianQiFaSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

//================TaoYuanJieYiSkill================

type TaoYuanJieYiSkill struct {
	*BaseCheckSkill
}

func NewTaoYuanJieYiSkill() *TaoYuanJieYiSkill {
	return &TaoYuanJieYiSkill{BaseCheckSkill: NewBaseCheckSkill()}
}

func (g *TaoYuanJieYiSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *TaoYuanJieYiSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewTaoYuanJieYiPrepareStep(event.Src), NewLoopTriggerUseKitStep(event.Src),
		NewCheckRespKitStep(), NewTaoYuanJieYiExecuteStep())
}

//================DelayKitSkill===================

type DelayKitSkill struct {
	*BaseCheckSkill
	IsSelf bool // 是否是针对自己发动的延时锦囊，暂时只有闪电
}

func (g *DelayKitSkill) GetMaxDesc(src *Player, card *Card) int {
	if g.IsSelf {
		return 0
	}
	return 1
}

func (g *DelayKitSkill) CheckUse(src *Player, card *Card, extra *StepExtra) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionUseCard, Src: src, Card: card})
	return !condition.Invalid
}

func (g *DelayKitSkill) CheckTarget(src, desc *Player, card *Card) bool {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionCardPoint, Src: src, Desc: desc, Card: card})
	if condition.Invalid {
		return false
	}
	return !g.IsSelf && src != desc
}

func (g *DelayKitSkill) CreateEffect(event *Event) IEffect {
	if !g.IsSelf && len(event.Descs) != 1 {
		MainGame.AddTip("[%s]必须选择一名角色", event.Card.Desc.Name)
		return nil
	}
	return NewEffect(NewDelayKitExecuteStep())
}

func NewDelayKitSkill(isSelf bool) *DelayKitSkill {
	return &DelayKitSkill{IsSelf: isSelf, BaseCheckSkill: NewBaseCheckSkill()}
}

//===================LeBuSiShuSkill======================

type LeBuSiShuSkill struct {
	*BaseSkill
}

func (g *LeBuSiShuSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewCardCheckInvalidStep(),
		NewJudgeCardJudgeStep(), NewJudgeCardEndStep(), NewLeBuSiShuStep())
}

func NewLeBuSiShuSkill() *LeBuSiShuSkill {
	return &LeBuSiShuSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

//=========================ShanDianSkill=======================

type ShanDianSkill struct {
	*BaseSkill
}

func (g *ShanDianSkill) CreateEffect(event *Event) IEffect {
	return NewEffect(NewLoopTriggerUseKitStep(event.Src), NewCheckRespKitStep(), NewShanDianInvalidStep(),
		NewJudgeCardJudgeStep(), NewJudgeCardEndStep(), NewShanDianStep())
}

func NewShanDianSkill() *ShanDianSkill {
	return &ShanDianSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

//======================PlayerDyingSkill===========================

type PlayerDyingSkill struct {
	*BaseSkill
	Player *Player
}

func (g *PlayerDyingSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventPlayerDying || event.Desc != g.Player {
		return nil
	} // 自己要死了
	return NewEffect(NewPlayerDyingCheckStep(), NewPlayerDyingLoopStep(), NewPlayerDyingResStep())
}

func NewPlayerDyingSkill(player *Player) *PlayerDyingSkill {
	return &PlayerDyingSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

//=====================SysPlayerDieSkill=======================

type SysPlayerDieSkill struct {
	*BaseSkill
}

func NewSysPlayerDieSkill() *SysPlayerDieSkill {
	return &SysPlayerDieSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysPlayerDieSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventPlayerDie {
		return nil
	}
	return NewEffect(NewSysPlayerDieStep())
}

//======================SysGameOverSkill==========================

type SysGameOverSkill struct {
	*BaseSkill
}

func NewSysGameOverSkill() *SysGameOverSkill {
	return &SysGameOverSkill{BaseSkill: NewBaseSkill(TagNone, "")}
}

func (s *SysGameOverSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventGameOverCheck {
		return nil
	}
	// 判断游戏是否结束
	zhongChens := MainGame.GetPlayers(func(player *Player) bool {
		return !player.IsDie && player.Role == RoleZhongChen
	})
	fanZeis := MainGame.GetPlayers(func(player *Player) bool {
		return !player.IsDie && player.Role == RoleFanZei
	})
	neiJians := MainGame.GetPlayers(func(player *Player) bool {
		return !player.IsDie && player.Role == RoleNeiJian
	})
	info := ""
	// 主公死亡
	if event.Src.Role == RoleZhuGong {
		// 只有只剩余内奸时内奸胜利，其他都是反贼胜利
		if len(zhongChens) == 0 && len(fanZeis) == 0 && len(neiJians) > 0 {
			info = "内奸胜利"
		} else {
			info = "反贼胜利"
		}
	} else if len(fanZeis) == 0 && len(neiJians) == 0 { // 内奸反贼全部死亡
		info = "主公&忠臣胜利"
	}
	if len(info) == 0 {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSysGameOverStep(res, info))
	return res
}

//===================JianXiongSkill====================

type JianXiongSkill struct {
	*BaseSkill
	Player *Player
}

func NewJianXiongSkill(player *Player) *JianXiongSkill {
	return &JianXiongSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "奸雄")}
}

func (j *JianXiongSkill) CreateEffect(event *Event) IEffect {
	// 使用卡牌造成伤害
	if event.Type != EventPlayerHurt || event.Desc != j.Player || event.Card == nil || j.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(), NewJianXiongStep())
	return res
}

//====================HuJiaSkill======================

type HuJiaSkill struct {
	*BaseSkill
	Player *Player
	Card   *CardWrap
}

func NewHuJiaSkill(player *Player) *HuJiaSkill {
	return &HuJiaSkill{Player: player, BaseSkill: NewBaseSkill(TagZhuGong, "护驾"),
		Card: NewSimpleCardWrap(&Card{Name: "闪", Point: PointNone, Suit: SuitNone, Type: CardBasic})}
}

func (h *HuJiaSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventRespCard || event.Desc != h.Player || !event.WrapFilter(h.Card) || h.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(), NewHuJiaLoopStep(h.Player), NewHuJiaCheckStep())
	return res
}

//===============FanKuiSkill=================

type FanKuiSkill struct {
	*BaseSkill
	Player *Player
}

func NewFanKuiSkill(player *Player) *FanKuiSkill {
	return &FanKuiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "反馈")}
}

func (f *FanKuiSkill) CreateEffect(event *Event) IEffect {
	// 有伤害来源的伤害
	if event.Type != EventPlayerHurt || event.Desc != f.Player || event.Src == nil || f.Player.IsBot {
		return nil
	}
	src := event.Src // 无牌可拿
	cards := src.GetCards()
	equips := src.GetEquips()
	delayKits := src.GetDelayKits()
	if len(cards) == 0 && len(equips) == 0 && len(delayKits) == 0 {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
		NewSelectPlayerCardStep(1, res, NewAllCard(cards, equips, delayKits)), NewFanKuiStep())
	return res
}

//=================GuiCaiSkill====================

type GuiCaiSkill struct {
	*BaseSkill
	Player *Player
}

func (f *GuiCaiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventJudgeCard || f.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
		NewSelectNumCardStep(f.AnyFilter, 1, 1, false, f.Player, res), NewGuiCaiStep(f.Player))
	return res
}

func (f *GuiCaiSkill) AnyFilter(card *Card) bool {
	return true
}

func NewGuiCaiSkill(player *Player) *GuiCaiSkill {
	return &GuiCaiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "鬼才")}
}

//================GangLieSkill=================

type GangLieSkill struct {
	*BaseSkill
	Player *Player
}

func NewGangLieSkill(player *Player) *GangLieSkill {
	return &GangLieSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "刚烈")}
}

func (g *GangLieSkill) CreateEffect(event *Event) IEffect {
	// 有伤害来源的伤害
	if event.Type != EventPlayerHurt || event.Desc != g.Player || event.Src == nil || g.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(), NewJudgeCardJudgeStep(),
		NewJudgeCardEndStep(), NewGangLieCheckStep(), NewGangLieExecuteStep())
	return res
}

//================TuXiSkill===================

type TuXiSkill struct {
	*BaseSkill
	Player *Player
}

func NewTuXiSkill(player *Player) *TuXiSkill {
	return &TuXiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "突袭")}
}

func (t *TuXiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventStageDraw || event.Src != t.Player || t.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
		NewSelectPlayerStep(1, 2, t.HasCardFilter, res), NewTuXiLoopStep())
	return res
}

func (t *TuXiSkill) HasCardFilter(player *Player) bool {
	return player != t.Player && len(player.Cards) > 0
}

//==================LuoYiSkill==================

type LuoYiSkill struct {
	*BaseSkill
	Player *Player
}

func NewLuoYiSkill(player *Player) *LuoYiSkill {
	return &LuoYiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "裸衣")}
}

func (l *LuoYiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventStageDraw || event.Src != l.Player || l.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewLuoYiStep())
	return res
}

//=================LuoYiBuffSkill=====================

type LuoYiBuffSkill struct {
	*BaseSkill
	Player *Player
}

func (l *LuoYiBuffSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionDrawCardNum || condition.Src != l.Player {
		return false
	} // 摸牌-1
	condition.CardNum--
	return false
}

func (l *LuoYiBuffSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventRespCard && event.Src == l.Player {
		return NewEffect(NewLuoYiBuffStep())
	}
	if event.Type == EventStageEnd && event.Src == l.Player {
		return NewEffect(NewRemoveSkillStep(l, l.Player))
	}
	return nil
}

func NewLuoYiBuffSkill(player *Player) *LuoYiBuffSkill {
	return &LuoYiBuffSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "")}
}

//===================TianDuSkill=====================

type TianDuSkill struct {
	*BaseSkill
	Player *Player
}

func (t *TianDuSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventJudgeEnd || event.Src != t.Player || t.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewTianDuStep())
	return res
}

func NewTianDuSkill(player *Player) *TianDuSkill {
	return &TianDuSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "天妒")}
}

//=================YiJiSkill====================

type YiJiSkill struct {
	*BaseSkill
	Player *Player
}

func (y *YiJiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventPlayerHurt || event.Desc != y.Player || y.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewYiJiPrepareStep(), NewYiJiCheckStep(),
		NewSelectPlayerStep(1, 1, y.NotSelfFilter, res), NewYiJiExecuteStep())
	return res
}

func (y *YiJiSkill) NotSelfFilter(player *Player) bool {
	return player != y.Player
}

func NewYiJiSkill(player *Player) *YiJiSkill {
	return &YiJiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "遗计")}
}

//===================QingGuoSkill====================

type QingGuoSkill struct {
	*BaseSkill
	Player *Player
	Card   *CardWrap
}

func (q *QingGuoSkill) CreateEffect(event *Event) IEffect {
	// 必须接受转换「闪」
	if event.Type != EventRespCard || event.Desc != q.Player || !event.WrapFilter(q.Card) || q.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
		NewSelectNumCardStep(q.BlackFilter, 1, 1, false, q.Player, res), NewQingGuoStep())
	return res
}

func (q *QingGuoSkill) BlackFilter(card *Card) bool {
	return IsBlackSuit(card.Suit)
}

func NewQingGuoSkill(player *Player) *QingGuoSkill {
	return &QingGuoSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "倾国"),
		Card: NewTransCardWrap(&Card{Name: "闪", Type: CardBasic}, make([]*Card, 0))}
}

//=======================LuoShenSkill=======================

type LuoShenSkill struct {
	*BaseSkill
	Player *Player
}

func (s *LuoShenSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventStagePrepare || event.Src != s.Player || s.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(), NewJudgeCardJudgeStep(),
		NewJudgeCardEndStep(), NewLuoShenStep())
	return res
}

func NewLuoShenSkill(player *Player) *LuoShenSkill {
	return &LuoShenSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "洛神")}
}

//==================RenDeSkill=======================

type RenDeSkill struct {
	*BaseSkill
	Player  *Player
	CardNum int // 给的牌数
}

func (r *RenDeSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == r.Player {
		r.CardNum = 0 // 恢复一下
	}
	if event.Type != EventUseSkill || event.Src != r.Player || r.Player.IsBot { // bot 只会使用简单的技能
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(r.AnyFilter, 1, len(r.Player.Cards), false, r.Player, res),
		NewSelectPlayerStep(1, 1, r.NotSelfFilter, res), NewRenDeStep(r))
	return res
}

func (r *RenDeSkill) AnyFilter(card *Card) bool {
	return true
}

func (r *RenDeSkill) NotSelfFilter(player *Player) bool {
	return player != r.Player
}

func NewRenDeSkill(player *Player) *RenDeSkill {
	return &RenDeSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "仁德")}
}

//======================JiJiangSkill=======================

type JiJiangSkill struct {
	*BaseSkill
	Player   *Player
	Card     *CardWrap
	ShaSkill *ShaSkill
}

func (r *JiJiangSkill) CreateEffect(event *Event) IEffect {
	// 响应「杀」时
	if event.Type == EventRespCard && event.Desc == r.Player && event.WrapFilter(r.Card) && !r.Player.IsBot {
		res := NewEffectWithUI()
		res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
			NewJiJiangLoopStep(r.Player), NewJiJiangCheck1Step())
		return res
	} // 主动出杀
	if event.Type == EventUseSkill && event.Src == r.Player && !r.Player.IsBot {
		if !r.ShaSkill.CheckUse(event.Src, r.Card.Desc, event.StepExtra) { // 没有使用机会了
			return nil
		}
		res := NewEffectWithUI()
		res.SetSteps(NewSelectPlayerStep(0, 1, r.ShaFilter, res), NewSelectPlayerVerifyStep(),
			NewJiJiangLoopStep(r.Player), NewJiJiangCheck2Step())
		return res
	}
	return nil
}

func (r *JiJiangSkill) ShaFilter(player *Player) bool {
	return r.ShaSkill.CheckTarget(r.Player, player, r.Card.Desc)
}

func NewJiJiangSkill(player *Player) *JiJiangSkill {
	return &JiJiangSkill{Player: player, BaseSkill: NewBaseSkill(TagZhuGong|TagActive, "激将"), ShaSkill: NewShaSkill(),
		Card: NewSimpleCardWrap(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic})}
}

//====================WuShengSkill=====================

type WuShengSkill struct {
	*BaseSkill
	Player   *Player
	Card     *CardWrap
	ShaSkill *ShaSkill
}

func (r *WuShengSkill) CreateEffect(event *Event) IEffect {
	// 响应「杀」时
	if event.Type == EventRespCard && event.Desc == r.Player && event.WrapFilter(r.Card) && !r.Player.IsBot {
		res := NewEffectWithUI()
		res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
			NewSelectNumCardStep(r.RedFilter, 1, 1, true, r.Player, res), NewWuShengCheckStep())
		return res
	} // 主动出杀
	if event.Type == EventUseSkill && event.Src == r.Player && !r.Player.IsBot {
		if !r.ShaSkill.CheckUse(event.Src, r.Card.Desc, event.StepExtra) { // 没有使用机会了
			return nil
		}
		res := NewEffectWithUI()
		res.SetSteps(NewSelectPlayerStep(0, 1, r.ShaFilter, res), NewSelectPlayerVerifyStep(),
			NewSelectNumCardStep(r.RedFilter, 1, 1, true, r.Player, res),
			NewTransCardStep(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic, Skill: NewShaSkill()}))
		return res
	}
	return nil
}

func (r *WuShengSkill) RedFilter(card *Card) bool {
	return IsRedSuit(card.Suit)
}

func (r *WuShengSkill) ShaFilter(player *Player) bool {
	return r.ShaSkill.CheckTarget(r.Player, player, r.Card.Desc)
}

func NewWuShengSkill(player *Player) *WuShengSkill {
	return &WuShengSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "武圣"), ShaSkill: NewShaSkill(),
		Card: NewSimpleCardWrap(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic})}
}

//======================PaoXiaoSkill=======================

type PaoXiaoSkill struct {
	*BaseSkill
	Player *Player
}

func (s *PaoXiaoSkill) HandleCondition(condition *Condition) bool {
	// 自己使用卡牌「杀」时生效
	if condition.Type != ConditionUseCard || condition.Src != s.Player || condition.Card.Name != "杀" {
		return false
	}
	condition.MaxUseSha = 999 // 也不是没限制，只是限制比较宽松
	return false
}

func NewPaoXiaoSkill(player *Player) *PaoXiaoSkill {
	return &PaoXiaoSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "咆哮")}
}

//======================GuanXingSkill========================

type GuanXingSkill struct {
	*BaseSkill
	Player *Player
}

func (g *GuanXingSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventStagePrepare || event.Src != g.Player || g.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	num := Min(5, len(MainGame.GetPlayers(g.LiveFilter)))
	cards := MainGame.DrawCard(num)
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
		NewGuanXingStep(res, NewGuanXing(cards)))
	return res
}

func (g *GuanXingSkill) LiveFilter(player *Player) bool {
	return !player.IsDie
}

func NewGuanXingSkill(player *Player) *GuanXingSkill {
	return &GuanXingSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "观星")}
}

//==================KongChengSkill=======================

type KongChengSkill struct {
	*BaseSkill
	Player *Player
}

func (s *KongChengSkill) HandleCondition(condition *Condition) bool {
	hasCard := len(s.Player.Cards) > 0
	if condition.Type != ConditionCardPoint || condition.Desc != s.Player || hasCard {
		return false
	}
	card := condition.Card
	if card.Name != "杀" && card.Name != "决斗" {
		return false
	}
	condition.Invalid = true
	return true
}

func NewKongChengSkill(player *Player) *KongChengSkill {
	return &KongChengSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "空城")}
}

//===================LongDanSkill====================

type LongDanSkill struct {
	*BaseSkill
	Player   *Player
	ShaCard  *CardWrap
	ShanCard *CardWrap
	ShaSkill *ShaSkill
}

func (l *LongDanSkill) CreateEffect(event *Event) IEffect {
	// 响应「杀」/「闪」时
	if event.Type == EventRespCard && event.Desc == l.Player && !l.Player.IsBot {
		cardName, needCard := "", ""
		if event.WrapFilter(l.ShaCard) {
			cardName, needCard = "杀", "闪"
		} else if event.WrapFilter(l.ShanCard) {
			cardName, needCard = "闪", "杀"
		}
		if len(cardName) == 0 || len(needCard) == 0 {
			return nil
		}
		res := NewEffectWithUI()
		res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
			NewSelectNumCardStep(l.NeedFilter(needCard), 1, 1, false, l.Player, res),
			NewLongDanCheckStep(cardName))
		return res
	} // 主动出杀
	if event.Type == EventUseSkill && event.Src == l.Player && !l.Player.IsBot {
		if !l.ShaSkill.CheckUse(event.Src, l.ShaCard.Desc, event.StepExtra) { // 没有使用机会了
			return nil
		}
		res := NewEffectWithUI()
		res.SetSteps(NewSelectPlayerStep(0, 1, l.ShaFilter, res), NewSelectPlayerVerifyStep(),
			NewSelectNumCardStep(l.NeedFilter("闪"), 1, 1, false, l.Player, res),
			NewTransCardStep(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic, Skill: NewShaSkill()}))
		return res
	}
	return nil
}

func (l *LongDanSkill) NeedFilter(needCard string) CardFilter {
	return func(card *Card) bool {
		return card.Name == needCard
	}
}

func (l *LongDanSkill) ShaFilter(player *Player) bool {
	return l.ShaSkill.CheckTarget(l.Player, player, l.ShaCard.Desc)
}

func NewLongDanSkill(player *Player) *LongDanSkill {
	return &LongDanSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "龙胆"), ShaSkill: NewShaSkill(),
		ShaCard:  NewTransCardWrap(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic}, make([]*Card, 0)),
		ShanCard: NewTransCardWrap(&Card{Name: "闪", Point: PointNone, Suit: SuitNone, Type: CardBasic}, make([]*Card, 0))}
}

//=======================MaShuSkill=======================

type MaShuSkill struct {
	*BaseSkill
	Player *Player
}

func (s *MaShuSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionGetDist || condition.Src != s.Player {
		return false
	}
	condition.Dist--
	return false
}

func NewMaShuSkill(player *Player) *MaShuSkill {
	return &MaShuSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "马术")}
}

//=====================TieQiSkill=====================

type TieQiSkill struct {
	*BaseSkill
	Player *Player
}

func (t *TieQiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventRespCard || event.Src != t.Player || event.Card.Desc.Name != "杀" || t.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(), NewJudgeCardJudgeStep(),
		NewJudgeCardEndStep(), NewTieQiStep())
	return res
}

func NewTieQiSkill(player *Player) *TieQiSkill {
	return &TieQiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "铁骑")}
}

//=====================JiZhiSkill======================

type JiZhiSkill struct {
	*BaseSkill
	Player *Player
}

func (j *JiZhiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventUseCard || event.Src != j.Player || j.Player.IsBot {
		return nil
	}
	card := event.Card.Desc
	if card.Type != CardKit || card.KitType != KitInstant {
		return nil
	}
	//res := NewEffectWithUI()
	//res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewJiZhiStep())
	return NewEffect(NewDrawCardStep(j.Player, 1))
}

func NewJiZhiSkill(player *Player) *JiZhiSkill { // 暂时当成锁定技
	return &JiZhiSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "集智")}
}

//======================QiCaiSkill========================

type QiCaiSkill struct {
	*BaseSkill
	Player *Player
}

func (s *QiCaiSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionGetDist || condition.Src != s.Player || condition.Card.Type != CardKit {
		return false
	}
	condition.Dist -= 999
	return false
}

func NewQiCaiSkill(player *Player) *QiCaiSkill {
	return &QiCaiSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "奇才")}
}

//======================ZhiHengSkill=========================

type ZhiHengSkill struct {
	*BaseSkill
	Player *Player
	Used   bool
}

func (z *ZhiHengSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == z.Player {
		z.Used = false
	}
	if z.Used || event.Type != EventUseSkill || event.Src != z.Player || z.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(z.AnyFilter, 1, 999, true, z.Player, res), NewZhiHengStep(z))
	return res
}

func (z *ZhiHengSkill) AnyFilter(card *Card) bool {
	return true
}

func NewZhiHengSkill(player *Player) *ZhiHengSkill {
	return &ZhiHengSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "制衡")}
}

//=================JiuYuanSkill=====================

type JiuYuanSkill struct {
	*BaseSkill
	Player *Player
	Card   *Card
}

func (j *JiuYuanSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventAskCard || event.Src != j.Player || !event.Filter(j.Card) || event.Desc.Force != ForceWu {
		return nil
	}
	return NewEffect(NewJiuYuanStep())
}

func NewJiuYuanSkill(player *Player) *JiuYuanSkill {
	return &JiuYuanSkill{Player: player, BaseSkill: NewBaseSkill(TagZhuGong|TagLock, "救援"),
		Card: &Card{Name: "桃", Point: PointNone, Suit: SuitNone, Type: CardBasic}}
}

//===================QiXiSkill=====================

type QiXiSkill struct {
	*BaseSkill
	Player *Player
	Card   *Card
}

func (x *QiXiSkill) CreateEffect(event *Event) IEffect {
	// 若有多个主动技能这里也不会弄混，实际外面是直接通过技能调用的，这里主要防止处理非主动调用的场景
	if event.Type != EventUseSkill || event.Src != x.Player || x.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(x.BlackFilter, 1, 1, true, x.Player, res), // 确定牌
		NewSelectPlayerStep(1, 1, x.GuoHeChaiQiaoFilter, res), // 确定目标
		NewTransCardStep(x.Card))                              // 确定要转换的目标牌
	return res
}

func (x *QiXiSkill) BlackFilter(card *Card) bool {
	return IsBlackSuit(card.Suit)
}

func (x *QiXiSkill) GuoHeChaiQiaoFilter(player *Player) bool {
	return x.Card.Skill.CheckTarget(x.Player, player, x.Card)
}

func NewQiXiSkill(player *Player) *QiXiSkill {
	return &QiXiSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "奇袭"),
		Card: &Card{Name: "过河拆桥", Point: PointNone, Suit: SuitNone, Type: CardKit, KitType: KitInstant, Skill: NewGuoHeChaiQiaoSkill()}}
}

//==================KeJiSkill====================

type KeJiSkill struct {
	*BaseSkill
	Player   *Player
	ShaCount int
}

func (k *KeJiSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == k.Player {
		k.ShaCount = 0
	}
	if event.Type == EventUseCard && event.Src == k.Player && event.Card.Desc.Name == "杀" {
		k.ShaCount++
	}
	if k.ShaCount > 0 || event.Type != EventStageDiscard || event.Src != k.Player || k.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewKeJiStep())
	return res
}

func NewKeJiSkill(player *Player) *KeJiSkill {
	return &KeJiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "克己")}
}

//==================KuRouSkill====================

type KuRouSkill struct {
	*BaseSkill
	Player *Player
}

func (x *KuRouSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventUseSkill || event.Src != x.Player || x.Player.IsBot {
		return nil
	}
	return NewEffect(NewKuRouStep())
}

func NewKuRouSkill(player *Player) *KuRouSkill {
	return &KuRouSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "苦肉")}
}

//===================YingZiSkill===================

type YingZiSkill struct {
	*BaseSkill
	Player *Player
}

func (l *YingZiSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionDrawCardNum || condition.Src != l.Player {
		return false
	} // 摸牌+1
	condition.CardNum++
	return false
}

func NewYingZiSkill(player *Player) *YingZiSkill { // 这里直接当锁定技处理了
	return &YingZiSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "英姿")}
}

//=====================FanJianSkill=====================

type FanJianSkill struct {
	*BaseSkill
	Player *Player
	Used   bool
}

func (z *FanJianSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == z.Player {
		z.Used = false
	}
	if z.Used || event.Type != EventUseSkill || event.Src != z.Player || z.Player.IsBot {
		return nil
	}
	z.Used = true
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerStep(1, 1, z.NotSelfFilter, res), NewFanJianReqStep(), NewFanJianCheckStep())
	return res
}

func (z *FanJianSkill) NotSelfFilter(player *Player) bool {
	return player != z.Player
}

func NewFanJianSkill(player *Player) *FanJianSkill {
	return &FanJianSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "反间")}
}

//===================GuoSeSkill===================

type GuoSeSkill struct {
	*BaseSkill
	Player *Player
	Card   *Card
}

func (x *GuoSeSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventUseSkill || event.Src != x.Player || x.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectNumCardStep(x.DiamondFilter, 1, 1, true, x.Player, res),
		NewSelectPlayerStep(1, 1, x.LeBuSiShuFilter, res), NewTransCardStep(x.Card))
	return res
}

func (x *GuoSeSkill) DiamondFilter(card *Card) bool {
	return card.Suit == SuitDiamond
}

func (x *GuoSeSkill) LeBuSiShuFilter(player *Player) bool {
	return x.Card.Skill.CheckTarget(x.Player, player, x.Card)
}

func NewGuoSeSkill(player *Player) *GuoSeSkill {
	return &GuoSeSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "国色"),
		Card: &Card{Name: "乐不思蜀", Point: PointNone, Suit: SuitNone, Type: CardKit, Skill: NewDelayKitSkill(false),
			KitType: KitDelay, Alias: "乐"}}
}

//==================LiuLiSkill==================

type LiuLiSkill struct {
	*BaseSkill
	Player *Player
	Skill  *ShaSkill
	Card   *Card
}

func (l *LiuLiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventCardPoint || event.Desc != l.Player || event.Card.Desc.Name != "杀" || len(l.Player.Cards) == 0 || l.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel),
		NewSelectNumCardStep(l.AnyFilter, 1, 1, true, l.Player, res),
		NewSelectPlayerStep(0, 1, l.Filter(event.Src), res), NewLiuLiStep())
	return res
}

func (l *LiuLiSkill) AnyFilter(card *Card) bool {
	return true
}

func (l *LiuLiSkill) Filter(src *Player) PlayerFilter {
	return func(player *Player) bool {
		return player != src && l.Card.Skill.CheckTarget(l.Player, player, l.Card)
	}
}

func NewLiuLiSkill(player *Player) *LiuLiSkill {
	return &LiuLiSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "流离"),
		Card: &Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic, Skill: NewShaSkill()}}
}

//================QianXunSkill===============

type QianXunSkill struct {
	*BaseSkill
	Player *Player
}

func (q *QianXunSkill) HandleCondition(condition *Condition) bool {
	if condition.Type != ConditionCardPoint || condition.Desc != q.Player {
		return false
	}
	cardName := condition.Card.Name
	if cardName != "乐不思蜀" && cardName != "顺手牵羊" {
		return false
	}
	condition.Invalid = true
	return true
}

func NewQianXunSkill(player *Player) *QianXunSkill {
	return &QianXunSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "谦逊")}
}

//==================LianYingSkill====================

type LianYingSkill struct {
	*BaseSkill
	Player *Player
}

func (l *LianYingSkill) CreateEffect(event *Event) IEffect {
	if (event.Type != EventUseCard && event.Type != EventRespCardAfter) || event.Src != l.Player {
		return nil
	}
	if len(l.Player.Cards) > 0 {
		return nil
	}
	return NewEffect(NewDrawCardStep(l.Player, 1))
}

func NewLianYingSkill(player *Player) *LianYingSkill { // 一般是当锁定技处理了
	return &LianYingSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "连营")}
}

//====================XiaoJiSkill=======================

type XiaoJiSkill struct {
	*BaseSkill
	Player *Player
}

func (x *XiaoJiSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventEquipLost || event.Src != x.Player {
		return nil
	}
	return NewEffect(NewDrawCardStep(x.Player, 2))
}

func NewXiaoJiSkill(player *Player) *XiaoJiSkill { // 直接认为是锁定技
	return &XiaoJiSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "枭姬")}
}

//======================JieYinSkill===========================

type JieYinSkill struct {
	*BaseSkill
	Player *Player
	Used   bool
}

func (y *JieYinSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == y.Player {
		y.Used = false
	}
	if y.Used || event.Type != EventUseSkill || event.Src != y.Player || len(y.Player.Cards) < 2 || y.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerStep(0, 1, y.Filter, res), NewSelectPlayerVerifyStep(),
		NewSelectNumCardStep(y.AnyFilter, 2, 2, false, y.Player, res), NewJieYinStep(y))
	return res
}

func (y *JieYinSkill) AnyFilter(card *Card) bool {
	return true
}

func (y *JieYinSkill) Filter(player *Player) bool {
	return player.General.Gender == GenderMan && player.MaxHp > player.Hp
}

func NewJieYinSkill(player *Player) *JieYinSkill {
	return &JieYinSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "结姻")}
}

//======================QingNangSkill=======================

type QingNangSkill struct {
	*BaseSkill
	Player *Player
	Used   bool
}

func (y *QingNangSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == y.Player {
		y.Used = false
	}
	if y.Used || event.Type != EventUseSkill || event.Src != y.Player || len(y.Player.Cards) < 1 || y.Player.IsBot {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerStep(0, 1, y.HurtFilter, res), NewSelectPlayerVerifyStep(),
		NewSelectNumCardStep(y.AnyFilter, 1, 1, false, y.Player, res), NewQingNangStep(y))
	return res
}

func (y *QingNangSkill) HurtFilter(player *Player) bool {
	return player.Hp < player.MaxHp
}

func (y *QingNangSkill) AnyFilter(card *Card) bool {
	return true
}

func NewQingNangSkill(player *Player) *QingNangSkill {
	return &QingNangSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "青囊")}
}

//=======================JiJiuSkill========================

type JiJiuSkill struct {
	*BaseSkill
	Player *Player
	Card   *Card
}

func (j *JiJiuSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventAskCard || event.Desc != j.Player || event.AskNum != 1 || !event.Filter(j.Card) {
		return nil
	}
	res := NewEffectWithUI()
	res.SetSteps(NewButtonSelectStep(res, TextConfirm, TextCancel), NewSelectCancelStep(),
		NewSelectNumCardStep(j.RedFilter, 1, 1, true, j.Player, res), NewJiJiuStep())
	return res
}

func (j *JiJiuSkill) RedFilter(card *Card) bool {
	return IsRedSuit(card.Suit)
}

func NewJiJiuSkill(player *Player) *JiJiuSkill {
	return &JiJiuSkill{Player: player, BaseSkill: NewBaseSkill(TagNone, "急救"),
		Card: &Card{Name: "桃", Point: PointNone, Suit: SuitNone, Type: CardBasic}}
}

//====================LiJianSkill========================

type LiJianSkill struct {
	*BaseSkill
	Player *Player
	Used   bool
}

func (y *LiJianSkill) CreateEffect(event *Event) IEffect {
	if event.Type == EventStagePrepare && event.Src == y.Player {
		y.Used = false
	}
	if y.Used || event.Type != EventUseSkill || event.Src != y.Player || len(y.Player.Cards) < 1 || y.Player.IsBot {
		return nil
	}
	y.Used = true
	res := NewEffectWithUI()
	res.SetSteps(NewSelectPlayerStep(2, 2, y.ManFilter, res),
		NewSelectNumCardStep(y.AnyFilter, 1, 1, false, y.Player, res), NewLiJianStep())
	return res
}

func (y *LiJianSkill) ManFilter(player *Player) bool {
	return player.General.Gender == GenderMan
}

func (y *LiJianSkill) AnyFilter(card *Card) bool {
	return true
}

func NewLiJianSkill(player *Player) *LiJianSkill {
	return &LiJianSkill{Player: player, BaseSkill: NewBaseSkill(TagActive, "离间")}
}

//==================BiYueSkill=================

type BiYueSkill struct {
	*BaseSkill
	Player *Player
}

func (y *BiYueSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventStageEnd || event.Src != y.Player {
		return nil
	}
	return NewEffect(NewDrawCardStep(y.Player, 1))
}

func NewBiYueSkill(player *Player) *BiYueSkill {
	return &BiYueSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "闭月")}
}

//====================YaoWuSkill========================

type YaoWuSkill struct {
	*BaseSkill
	Player *Player
}

func (y *YaoWuSkill) CreateEffect(event *Event) IEffect {
	// 有伤害来源的伤害
	if event.Type != EventPlayerHurt || event.Desc != y.Player || event.Src == nil || event.Card == nil {
		return nil
	}
	card := event.Card.Desc
	if card.Name != "杀" || !IsRedSuit(card.Suit) {
		return nil
	}
	return NewEffect(NewYaoWuStep())
}

func NewYaoWuSkill(player *Player) *YaoWuSkill {
	return &YaoWuSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "耀武")}
}

//==================WuShuangSkill=====================

type WuShuangSkill struct {
	*BaseSkill
	Player *Player
}

func (r *WuShuangSkill) CreateEffect(event *Event) IEffect {
	if event.Type != EventRespCardAfter || event.Desc != r.Player || event.Card == nil {
		return nil
	}
	card := event.Card.Desc
	if card.Name != "杀" && card.Name != "决斗" {
		return nil
	}
	return NewEffect(NewWuShuangReqStep(), NewWuShuangCheckStep())
}

func NewWuShuangSkill(player *Player) *WuShuangSkill {
	return &WuShuangSkill{Player: player, BaseSkill: NewBaseSkill(TagLock, "无双")}
}
