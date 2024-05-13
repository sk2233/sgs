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
	cards := event.Desc.GetHandCards()
	equips := event.Desc.GetEquipCards()
	delayKits := event.Desc.GetDelayKitCards()
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
	cards := event.Descs[0].GetHandCards()
	equips := event.Descs[0].GetEquipCards()
	delayKits := event.Descs[0].GetDelayKitCards()
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
	} // 必须有牌可拆
	if len(desc.Cards) == 0 && len(desc.Equips) == 0 && len(desc.DelayKits) == 0 {
		return false
	}
	// 这里的Dist是位置距离算上马
	distCond := MainGame.ComputeCondition(&Condition{Type: ConditionGetDist, Src: src, Desc: desc})
	return distCond.Dist <= 1
}

func (g *ShunShouQianYangSkill) CreateEffect(event *Event) IEffect {
	if len(event.Descs) != 1 {
		MainGame.AddTip("[顺手牵羊]需要指定一名角色")
		return nil
	}
	res := NewEffectWithUI() // 暂时没有考虑多目标选择，若需要参考「万箭齐发」
	cards := event.Descs[0].GetHandCards()
	equips := event.Descs[0].GetEquipCards()
	delayKits := event.Descs[0].GetDelayKitCards()
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
	return &BotChooseCardSkill{Player: player}
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
	res.SetSteps(NewChooseNumCardStep(event.ChooseNum, res, NewChooseCard(event.Cards)), NewPlayerChooseCardStep())
	return res
}

func NewPlayerChooseCardSkill(player *Player) *PlayerChooseCardSkill {
	return &PlayerChooseCardSkill{Player: player}
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
	return src != desc
}

func (g *JueDouSkill) CreateEffect(event *Event) IEffect {
	event.HurtVal = 1
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
