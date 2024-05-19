/*
@author: sk
@date: 2024/5/1
*/
package main

type StepExtra struct {
	Index            int // 步骤进行到那里了
	JudgeCard        *CardWrap
	ShaCount         int // 出了几次杀了
	Card             *CardUI
	Cards            []*Card
	MaxDesc          int
	Result1, Result2 *Event // 事件即是参数，也存储结果
	Desc             *Player
	Select           string
	Players          []*Player
}

func NewStepExtra() *StepExtra {
	return &StepExtra{Index: 0}
}

type IStep interface {
	Update(event *Event, extra *StepExtra) // 执行效果并返回是否执行结束
}

//==================SysGameStartStep==================

type SysNextPlayerStep struct {
}

func NewSysNextPlayerStep() *SysNextPlayerStep {
	return &SysNextPlayerStep{}
}

func (s *SysNextPlayerStep) Update(event *Event, extra *StepExtra) {
	MainGame.NextPlayer()  // 轮到下一个玩家了
	extra.Index = MaxIndex // 结束效果
}

//===================TriggerEventStep简单触发一下事件=====================

type TriggerEventStep struct {
	EventType EventType
}

func NewTriggerEventStep(eventType EventType) *TriggerEventStep {
	return &TriggerEventStep{EventType: eventType}
}

func (t *TriggerEventStep) Update(event *Event, extra *StepExtra) {
	// 简单触发一下事件就继续向下走
	MainGame.TriggerEvent(&Event{Type: t.EventType, Src: event.Src, StepExtra: extra}) // TODO 参数后续可能需要继续补充
	extra.Index++
}

//=====================DrawStageMainStep摸牌阶段的主要步骤========================

type DrawStageMainStep struct {
}

func NewDrawStageMainStep() *DrawStageMainStep {
	return &DrawStageMainStep{}
}

func (d *DrawStageMainStep) Update(event *Event, extra *StepExtra) {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionDrawCardNum, Src: event.Src, CardNum: 2})
	event.Src.DrawCard(condition.CardNum)
	// TODO TEST
	//event.Src.AddCard(&Card{Name: "丈八蛇矛", Point: CardPoint(rand.Intn(13) + 1), Alias: "丈八蛇矛3",
	//	Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	extra.Index = MaxIndex
}

//============================JudgeStageExecuteStep判定阶段判定牌生效完清理步骤=================================

type JudgeStageExecuteStep struct {
}

func NewJudgeStageExecuteStep() *JudgeStageExecuteStep {
	return &JudgeStageExecuteStep{}
}

func (j *JudgeStageExecuteStep) Update(event *Event, extra *StepExtra) { // 普通判定也有这也这一步骤，但是判定阶段需要构成循环
	if len(event.Src.DelayKits) > 0 {
		delayKit := event.Src.DelayKits[0]
		event.Src.RemoveDelayKit(delayKit.Card.Desc)
		// 以当前目标为角色发动效果即可
		temp := &Event{Type: EventUseCard, Src: event.Src, Desc: event.Src, Card: delayKit.Card,
			StepExtra: extra, StageExtra: event.StageExtra}
		effect := delayKit.Skill.CreateEffect(temp)
		MainGame.PushAction(NewEffectGroup(temp, []IEffect{effect}))
		MainGame.AddToDesktopRaw(delayKit.Card)
	} else { // 处理完了
		extra.Index = MaxIndex
	}
}

//===========================JudgeCardJudgeStep判定牌判定Step===============================

type JudgeCardJudgeStep struct {
}

func NewJudgeCardJudgeStep() *JudgeCardJudgeStep {
	return &JudgeCardJudgeStep{}
}

func (j *JudgeCardJudgeStep) Update(event *Event, extra *StepExtra) {
	extra.Index++
	card := MainGame.DrawCard(1)[0]
	MainGame.AddToDesktop(card)               // 先添加进处理区，不是任何人的牌
	extra.JudgeCard = NewSimpleCardWrap(card) // 进行判定可能经历修改
	MainGame.TriggerEvent(&Event{Type: EventJudgeCard, Src: event.Src, StepExtra: extra})
}

//=======================JudgeCardEndStep判定牌生效步骤==========================

type JudgeCardEndStep struct {
}

func NewJudgeCardEndStep() *JudgeCardEndStep {
	return &JudgeCardEndStep{}
}

func (j *JudgeCardEndStep) Update(event *Event, extra *StepExtra) {
	// 判定结束后，目标可以判定牌做自己的操作并需要负责回收判定牌到弃牌区域
	extra.Index++
	MainGame.TriggerEvent(&Event{Type: EventJudgeEnd, Src: event.Src, StepExtra: extra}) // 触发判定生效事件
}

//=========================PlayStageMainStep出牌主要阶段============================

type PlayStageCardStep struct {
	PlayStage *PlayStage // 需要获取到按钮状态
}

func NewPlayStageCardStep(playStage *PlayStage) *PlayStageCardStep {
	return &PlayStageCardStep{PlayStage: playStage}
}

// 选择卡牌 -> 下个Step选择目标
// 出牌无效
// 取消->下个阶段
// 技能->按技能处理
func (p *PlayStageCardStep) Update(event *Event, extra *StepExtra) {
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	player := event.Src
	// 判断点击手牌
	for i := len(player.Cards) - 1; i >= 0; i-- {
		card := player.Cards[i]
		if card.Click(x, y) {
			card.Select()
			extra.Card = card
			extra.MaxDesc = card.Card.Skill.GetMaxDesc(player, card.Card)
			MainGame.ResetPlayer()
			if extra.MaxDesc > 0 { // 若是不需要目标，提前全部设置为灰色
				MainGame.CheckPlayer(player, card.Card)
			} else {
				for _, item := range MainGame.Players {
					item.CanSelect = false
				}
			}
			extra.Index++ // 进入下个阶段
			return
		}
	}
	// 判断点击按钮
	if p.HandleBtnClick(x, y, extra) {
		return
	}
	p.HandleSkillClick(x, y, player, extra)
}

func (p *PlayStageCardStep) HandleBtnClick(x, y float32, extra *StepExtra) bool {
	for _, button := range p.PlayStage.Buttons {
		if button.Click(x, y) && button.Show == TextCancel {
			extra.Index = MaxIndex // 结束本阶段
			return true
		}
	}
	return false
}

func (p *PlayStageCardStep) HandleSkillClick(x float32, y float32, player *Player, extra *StepExtra) {
	for _, skill := range player.Skills {
		if skill.Click(x, y) {
			event := &Event{Type: EventUseSkill, Src: player, StepExtra: extra}
			effect := skill.Skill.CreateEffect(event)
			if effect != nil {
				MainGame.PushAction(NewEffectGroup(event, []IEffect{effect}))
			}
		}
	}
}

//===============================PlayStagePlayerStep已经选择卡牌了需要选择目标了=================================

type PlayStagePlayerStep struct {
	PlayStage *PlayStage // 需要获取到按钮状态
}

func NewPlayStagePlayerStep(playStage *PlayStage) *PlayStagePlayerStep {
	return &PlayStagePlayerStep{PlayStage: playStage}
}

// 出牌->技能简单判断选择数目是否合适,合适出牌并重置为上一个阶段，否则无事发生
// 取消-> 取消当前选择，回到上一个 Step
// 选择目标->查看是否达到最大，最大设置其他角色不再可选
func (p *PlayStagePlayerStep) Update(event *Event, extra *StepExtra) {
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	player := event.Src
	// 判断点击按钮
	if p.HandleBtnClick(player, x, y, extra) {
		return
	}
	// 判断选择目标
	if MainGame.TogglePlayer(x, y) {
		if len(MainGame.GetSelectPlayer()) >= extra.MaxDesc {
			for _, item := range MainGame.Players { // 选满了，不能再选择了
				if !item.Select {
					item.CanSelect = false
				}
			}
		} else { // 没有选满，可以再选择，可能取消选择了
			MainGame.CheckPlayer(player, extra.Card.Card)
		}
		return
	}
}

func (p *PlayStagePlayerStep) HandleBtnClick(player *Player, x, y float32, extra *StepExtra) bool {
	for _, button := range p.PlayStage.Buttons {
		if button.Click(x, y) {
			if button.Show == TextPlayCard {
				card := extra.Card.Card
				desc := MainGame.GetSelectPlayer()
				event := &Event{Type: EventUseCard, Src: player, Descs: desc, Card: NewSimpleCardWrap(card), StepExtra: extra}
				effect := card.Skill.CreateEffect(event)
				if effect != nil { // 只需要简单校验即可，例如目标数是否有意义，TODO 后面里面可能进行具体校验
					MainGame.PushAction(NewEffectGroup(event, []IEffect{effect}))
					player.RemoveCard(card)
					MainGame.AddToDesktop(card)
					extra.Index = 0
					MainGame.ResetPlayer()
					MainGame.TriggerEvent(event) // 发动效果前，先声明使用了牌
				}
			} else if button.Show == TextCancel {
				extra.Card.UnSelect()
				extra.Index = 0
				MainGame.ResetPlayer()
			}
			return true
		}
	}
	return false
}

//=================================BotPlayStageStep bot专用的出牌阶段======================================

type BotPlayStageStep struct {
	Timer int
}

func (b *BotPlayStageStep) Update(event *Event, extra *StepExtra) {
	if b.Timer > 0 { // bot暂时不出牌
		b.Timer--
	} else {
		extra.Index = MaxIndex
	}
}

func NewBotPlayStageStep() *BotPlayStageStep {
	return &BotPlayStageStep{Timer: BotTimer}
}

//===========================DiscardStageCheckStep弃牌阶段检验是否需要弃牌============================

type DiscardStageCheckStep struct {
}

func NewDiscardStageCheckStep() *DiscardStageCheckStep {
	return &DiscardStageCheckStep{}
}

func (d *DiscardStageCheckStep) Update(event *Event, extra *StepExtra) {
	player := event.Src
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionMaxCard, Src: player})
	if len(player.Cards) > condition.MaxCard {
		extra.Index++
		player.ResetCard()
	} else {
		extra.Index = MaxIndex
	}
}

//=========================DiscardStageMainStep弃牌阶段一点一点弃牌============================

type DiscardStageMainStep struct {
	DiscardStage *DiscardStage
}

func NewDiscardStageMainStep(discardStage *DiscardStage) *DiscardStageMainStep {
	return &DiscardStageMainStep{DiscardStage: discardStage}
}

// 点击卡牌切换卡牌切换状态
// 点击「确定」弃牌
// 点击「取消」重置选择
func (d *DiscardStageMainStep) Update(event *Event, extra *StepExtra) {
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	player := event.Src
	// 判断点击按钮
	if d.HandleBtnClick(player, x, y, extra) {
		return
	}
	// 判断点击手牌
	if player.ToggleCard(x, y) {
		condition := MainGame.ComputeCondition(&Condition{Type: ConditionMaxCard, Src: player})
		if len(player.Cards)-len(player.GetSelectCard()) <= condition.MaxCard {
			player.DarkLastCard() // 已经选够了，不能再选了
		} else { // 可能取消了部分选择，又可以再选了
			for _, card := range player.Cards {
				card.CanSelect = true
			}
		}
	}
}

func (d *DiscardStageMainStep) HandleBtnClick(player *Player, x, y float32, extra *StepExtra) bool {
	for _, button := range d.DiscardStage.Buttons {
		if button.Click(x, y) {
			cards := player.GetSelectCard()
			if len(cards) == 0 { // 选择 0 张牌选什么都没有意义
				return true
			}
			if button.Show == TextConfirm {
				player.RemoveCard(cards...)
				// 弃之的牌，生命周期也立即结束
				MainGame.AddToDesktop(cards...)
				MainGame.DiscardFromDesktop(cards...)
				extra.Index = 0 // 再去 Check
			} else if button.Show == TextCancel {
				player.ResetCard()
			}
			return true
		}
	}
	return false
}

//=========================BotDiscardStageStep=============================

type BotDiscardStageStep struct {
	Timer int
}

func NewBotDiscardStageStep() *BotDiscardStageStep {
	return &BotDiscardStageStep{Timer: BotTimer} // 1s后弃牌，一步到位
}

func (b *BotDiscardStageStep) Update(event *Event, extra *StepExtra) {
	if b.Timer > 0 {
		b.Timer--
	} else {
		player := event.Src
		condition := MainGame.ComputeCondition(&Condition{Type: ConditionMaxCard, Src: player})
		if len(player.Cards) > condition.MaxCard {
			l := len(player.Cards) - condition.MaxCard
			cards := Map(player.Cards[:l], func(item *CardUI) *Card {
				return item.Card
			})
			player.RemoveCard(cards...)
			// 弃之的牌，生命周期也立即结束
			MainGame.AddToDesktop(cards...)
			MainGame.DiscardFromDesktop(cards...)
		}
		extra.Index = MaxIndex
	}
}

//====================UseShaLoopStep指定任意名角色为杀的目标循环====================

type UseShaLoopStep struct {
	Index int
}

func NewUseShaLoopStep() *UseShaLoopStep {
	return &UseShaLoopStep{}
}

func (r *UseShaLoopStep) Update(event *Event, extra *StepExtra) {
	if r.Index < len(event.Descs) { // 还有目标就继续执行
		extra.Desc = event.Descs[r.Index]
		MainGame.TriggerEvent(&Event{Type: EventCardPoint, Src: event.Src, Card: event.Card, Desc: extra.Desc, StepExtra: extra})
		r.Index++
		extra.Index++
	} else {
		MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
		event.StepExtra.ShaCount++
		extra.Index = MaxIndex
	}
}

//========================RespShaCardStep指定目标要响应卡牌=============================

type RespShaCardStep struct {
}

func (r *RespShaCardStep) Update(event *Event, extra *StepExtra) {
	extra.Result1 = &Event{Type: EventRespCard, Src: event.Src, Card: event.Card, Desc: extra.Desc,
		WrapFilter: r.ShanFilter, HurtVal: 1}
	MainGame.TriggerEvent(extra.Result1)
	extra.Index++
}

func (r *RespShaCardStep) ShanFilter(card *CardWrap) bool {
	return card != nil && card.Desc.Name == "闪"
}

func NewRespShaCardStep() *RespShaCardStep {
	return &RespShaCardStep{}
}

//=========================ShaHitCheckStep检查杀是否命中======================

type ShaHitCheckStep struct {
}

func NewShaHitCheckStep() *ShaHitCheckStep {
	return &ShaHitCheckStep{}
}

func (r *ShaHitCheckStep) Update(event *Event, extra *StepExtra) {
	temp := extra.Result1
	extra.Result1 = &Event{Type: EventShaHit, Src: temp.Src, Desc: temp.Desc, Card: temp.Card, HurtVal: temp.HurtVal, ShaHit: false}
	if !temp.Invalid && (temp.Resp == nil || temp.Force) { // 没有闪或强制崩血 且改事件有效
		extra.Result1.ShaHit = true
		MainGame.TriggerEvent(extra.Result1)
	}
	extra.Index++
}

//=======================RespShaExecuteStep=============================

type RespShaExecuteStep struct {
}

func (r *RespShaExecuteStep) Update(event *Event, extra *StepExtra) {
	// 结算后事件
	MainGame.TriggerEvent(&Event{Type: EventCardAfter, Src: event.Src, Card: event.Card, Desc: extra.Desc})
	temp := extra.Result1
	if temp.ShaHit {
		if extra.Desc.ChangeHp(-temp.HurtVal) {
			MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Src: temp.Src, Desc: temp.Desc, Card: temp.Card, HurtVal: temp.HurtVal})
		} else {
			MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Src: temp.Src, Desc: temp.Desc, Card: temp.Card, HurtVal: temp.HurtVal})
		}
	}
	extra.Index = 0 // 结算完毕回去
}

func NewRespShaExecuteStep() *RespShaExecuteStep {
	return &RespShaExecuteStep{}
}

//======================PlayerRespCardStep响应牌的目标是玩家========================

type PlayerRespCardStep struct {
	UIs     *EffectWithUI
	Init0   bool
	Buttons []*Button
}

func NewPlayerRespCardStep(uis *EffectWithUI) *PlayerRespCardStep {
	return &PlayerRespCardStep{UIs: uis, Init0: false}
}

// 「确定」-> 必须有选择的牌才有效    还可以再加一个步骤，若是发现无牌满足直接结束
// 「取消」-> 直接结束
// 点击卡牌-> 切换卡牌选择
func (p *PlayerRespCardStep) Update(event *Event, extra *StepExtra) {
	p.Init(event)
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	// 按钮点击
	if p.HandleBtnClick(event, x, y, extra) {
		return
	}
	// 卡牌点击
	player := event.Desc
	if player.ToggleCard(x, y) {
		if len(player.GetSelectCard()) == 1 {
			player.DarkLastCard()
		} else {
			player.CheckCardByWrapFilter(event.WrapFilter)
		}
	}
}

func (p *PlayerRespCardStep) HandleBtnClick(event *Event, x, y float32, extra *StepExtra) bool {
	for _, button := range p.Buttons {
		if button.Click(x, y) {
			player := event.Desc
			if button.Show == TextConfirm {
				cards := player.GetSelectCard()
				if len(cards) == 1 {
					player.RemoveCard(cards[0])
					event.Resp = NewSimpleCardWrap(cards[0])
					// 响应的牌打出后就立即结束流程了
					MainGame.AddToDesktopRaw(event.Resp)
					MainGame.DiscardFromDesktopRaw(event.Resp)
					MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: player, Desc: event.Src, Card: event.Card, Event: event})
					player.ResetCard()
					extra.Index = MaxIndex
				}
			} else if button.Show == TextCancel { // 结束
				player.ResetCard()
				extra.Index = MaxIndex
			}
			return true
		}
	}
	return false
}

func (p *PlayerRespCardStep) Init(event *Event) {
	if p.Init0 {
		return
	}
	p.Init0 = true
	p.Buttons = NewButtons(TextConfirm, TextCancel)
	for _, button := range p.Buttons {
		p.UIs.UIs = append(p.UIs.UIs, button)
	}
	text := NewText("请打出牌响应[%s]", event.Card.Desc.Name)
	text.X, text.Y = WinWidth/2, 280*2-p.Buttons[0].H-45
	p.UIs.UIs = append(p.UIs.UIs, text)
	event.Desc.ResetCard()
	event.Desc.CheckCardByWrapFilter(event.WrapFilter)
}

//======================BotRespCardStep响应牌的目标是bot=======================

type BotRespCardStep struct {
	Timer int
}

func NewBotRespCardStep() *BotRespCardStep {
	return &BotRespCardStep{Timer: BotTimer}
}

func (b *BotRespCardStep) Update(event *Event, extra *StepExtra) {
	// 有就出，没有就不出
	if b.Timer > 0 {
		b.Timer--
	} else {
		player := event.Desc
		for _, card := range player.Cards {
			temp := NewSimpleCardWrap(card.Card)
			if event.WrapFilter(temp) {
				player.RemoveCard(card.Card)
				// 响应的牌打出后就立即结束流程了
				MainGame.AddToDesktopRaw(temp)
				MainGame.DiscardFromDesktopRaw(temp)
				event.Resp = temp
				MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: player, Desc: event.Src, Card: event.Card, Event: event})
				break
			}
		}
		extra.Index = MaxIndex
	}
}

//====================TaoMainExecuteStep==========================

type TaoMainExecuteStep struct {
}

func NewTaoMainExecuteStep() *TaoMainExecuteStep {
	return &TaoMainExecuteStep{}
}

func (p *TaoMainExecuteStep) Update(event *Event, extra *StepExtra) { // 孙权的救援
	event.Src.ChangeHp(1)
	MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
	extra.Index = MaxIndex
}

//========================EquipMainExecuteStep=======================

type EquipMainExecuteStep struct {
}

func (e *EquipMainExecuteStep) Update(event *Event, extra *StepExtra) {
	player := event.Src
	card := event.Card.Desc
	// 装备牌使用完后会设置到装备区，不是移除到弃牌堆
	MainGame.RemoveFromDesktop(card)
	old := player.AddEquip(card)
	if old != nil {
		MainGame.TriggerEvent(&Event{Type: EventEquipLost, Src: player, Card: NewSimpleCardWrap(old)})
		MainGame.AddToDesktop(old)
		MainGame.DiscardFromDesktop(old)
	}
	extra.Index = MaxIndex
}

func NewEquipMainExecuteStep() *EquipMainExecuteStep {
	return &EquipMainExecuteStep{}
}

//=========================SelectNumCardStep============================

type SelectNumCardStep struct { // 选择固定数量卡牌的步骤
	Filter    CardFilter // 不允许二次转换，直接过滤原始牌就行
	Min, Max  int        // 选择固定数量的牌，没有其他约束了
	WithEquip bool       // 是否包含装备牌
	Player    *Player    // 为了更通用需要确认当前执行人
	UIs       *EffectWithUI
	Init0     bool
	Buttons   []*Button
}

func NewSelectNumCardStep(filter CardFilter, min, max int, withEquip bool, player *Player, UIs *EffectWithUI) *SelectNumCardStep {
	return &SelectNumCardStep{Filter: filter, Min: min, Max: max, WithEquip: withEquip, Player: player, UIs: UIs, Buttons: make([]*Button, 0)}
}

// 「确定」->把选择的卡牌带到下一步(必须数量足够)，并移除手牌中的装备牌，复原手牌
// 「取消」->直接结束，并移除手牌中的装备牌，复原手牌
// 点击卡牌->校验可选性
func (s *SelectNumCardStep) Update(event *Event, extra *StepExtra) {
	s.Init()
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	// 检验按钮
	if s.HandleBtnClick(x, y, extra) {
		return
	}
	// 检验选择牌
	if s.Player.ToggleCard(x, y) {
		if len(s.Player.GetSelectCard()) < s.Max {
			s.Player.CheckCardByFilter(s.Filter)
		} else {
			s.Player.DarkLastCard()
		}
	}
}

func (s *SelectNumCardStep) HandleBtnClick(x, y float32, extra *StepExtra) bool {
	for _, button := range s.Buttons {
		if button.Click(x, y) {
			if button.Show == TextConfirm {
				cards := s.Player.GetSelectCard()
				if len(cards) >= s.Min && len(cards) <= s.Max {
					s.EndSelect()
					extra.Cards = cards
					extra.Index++
				}
			} else if button.Show == TextCancel {
				s.EndSelect()
				extra.Index = MaxIndex // 后面有需求再改
			}
			return true
		}
	}
	return false
}

func (s *SelectNumCardStep) Init() {
	if s.Init0 {
		return
	}
	s.Init0 = true
	s.Buttons = NewButtons(TextConfirm, TextCancel)
	for _, button := range s.Buttons {
		s.UIs.UIs = append(s.UIs.UIs, button)
	}
	if s.WithEquip {
		cards := make([]*Card, 0)
		for _, equip := range s.Player.Equips {
			cards = append(cards, equip.Card)
		}
		s.Player.AddCard(cards...) // 暂时塞进手卡中
	}
	s.Player.ResetCard()
	s.Player.CheckCardByFilter(s.Filter)
}

func (s *SelectNumCardStep) EndSelect() {
	if s.WithEquip {
		cards := make([]*Card, 0)
		for _, equip := range s.Player.Equips {
			cards = append(cards, equip.Card)
		}
		s.Player.RemoveCard(cards...) // 从手牌中移除
	}
	s.Player.ResetCard()
}

//===================ZhangBaSheMaoRespStep====================

type ZhangBaSheMaoRespStep struct {
}

func (z *ZhangBaSheMaoRespStep) Update(event *Event, extra *StepExtra) { // 来到这里就是选好的
	event.Resp = NewTransCardWrap(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic}, extra.Cards)
	// 从玩家手中移除
	event.Desc.RemoveCard(extra.Cards...)
	event.Desc.RemoveEquip(extra.Cards...)
	// 添加进处理区就可以立即弃了
	MainGame.AddToDesktopRaw(event.Resp)
	MainGame.DiscardFromDesktopRaw(event.Resp)
	extra.Index = MaxIndex
}

func NewZhangBaSheMaoRespStep() *ZhangBaSheMaoRespStep {
	return &ZhangBaSheMaoRespStep{}
}

//========================GuanShiFuCheckStep=======================

type GuanShiFuCheckStep struct {
}

func (z *GuanShiFuCheckStep) Update(event *Event, extra *StepExtra) { // 来到这里就是选好的
	event.Event.Force = true
	// 从玩家手中移除
	event.Desc.RemoveCard(extra.Cards...)
	event.Desc.RemoveEquip(extra.Cards...)
	// 添加进处理区就可以立即弃了
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	extra.Index = MaxIndex
}

func NewGuanShiFuCheckStep() *GuanShiFuCheckStep {
	return &GuanShiFuCheckStep{}
}

//=======================QingHongJianStep========================

type QingHongJianStep struct {
}

func (q *QingHongJianStep) Update(event *Event, extra *StepExtra) {
	armor := event.Desc.Equips[EquipArmor]
	if armor != nil { // 若有防具就进行失效有效处理
		if event.Type == EventCardPoint {
			armor.Enable = false
		} else if event.Type == EventCardAfter {
			armor.Enable = true
		}
	}
	extra.Index = MaxIndex
}

func NewQingHongJianStep() *QingHongJianStep {
	return &QingHongJianStep{}
}

//=========================SelectPlayerCardStep==========================

type SelectPlayerCardStep struct { // 选择Player身上的牌，具体能选什么取决于给了什么牌
	Num     int // 选择固定数量的牌，没有其他约束了
	UIs     *EffectWithUI
	Init0   bool
	Buttons []*Button
	AllCard *AllCard
}

func NewSelectPlayerCardStep(num int, UIs *EffectWithUI, allCard *AllCard) *SelectPlayerCardStep {
	return &SelectPlayerCardStep{Num: num, UIs: UIs, AllCard: allCard, Init0: false}
}

// 点击「确定」流转到下一步，需要确认数量是否足够
// 点击「取消」直接结束
// 选择卡牌
func (s *SelectPlayerCardStep) Update(event *Event, extra *StepExtra) {
	s.Init()
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	// 检验按钮
	if s.HandleBtnClick(x, y, extra) {
		return
	}
	// 检验选择牌
	if s.AllCard.ToggleCard(x, y) {
		if len(s.AllCard.GetSelectCard()) < s.Num {
			s.AllCard.SetAllCanSelect()
		} else {
			s.AllCard.DarkLastCard()
		}
	}
}

func (s *SelectPlayerCardStep) HandleBtnClick(x, y float32, extra *StepExtra) bool {
	for _, button := range s.Buttons {
		if button.Click(x, y) {
			if button.Show == TextConfirm {
				cards := s.AllCard.GetSelectCard()
				if len(cards) == s.Num {
					extra.Cards = cards
					extra.Index++
				}
			} else if button.Show == TextCancel { // 直接结束
				extra.Index = MaxIndex // 后面有需求再改
			}
			return true
		}
	}
	return false
}

func (s *SelectPlayerCardStep) Init() {
	if s.Init0 {
		return
	}
	s.Init0 = true
	s.Buttons = NewButtons(TextConfirm, TextCancel)
	for _, button := range s.Buttons {
		s.UIs.UIs = append(s.UIs.UIs, button)
	}
	s.UIs.UIs = append(s.UIs.UIs, s.AllCard)
}

//===================QiLinGongExecuteStep移除选择的牌=================

type QiLinGongExecuteStep struct {
}

func (q *QiLinGongExecuteStep) Update(event *Event, extra *StepExtra) { // 到这里是肯定要移除的了
	// 注意我们使用的是受到伤害事件，这里对面是事件源
	event.Desc.RemoveEquip(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	extra.Index = MaxIndex
}

func NewQiLinGongExecuteStep() *QiLinGongExecuteStep {
	return &QiLinGongExecuteStep{}
}

//========================ButtonSelectStep玩家进行选择，并把选择的选项向下传递===========================

type ButtonSelectStep struct {
	UIs     *EffectWithUI
	Init0   bool
	Buttons []*Button
}

func (b *ButtonSelectStep) Update(event *Event, extra *StepExtra) {
	b.Init()
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	for _, button := range b.Buttons {
		if button.Click(x, y) {
			b.UIs.UIs = make([]IDraw, 0)
			b.Init0 = false
			extra.Select = button.Show
			extra.Index++
			return
		}
	}
}

func (b *ButtonSelectStep) Init() {
	if b.Init0 {
		return
	}
	b.Init0 = true
	for _, button := range b.Buttons {
		b.UIs.UIs = append(b.UIs.UIs, button)
	}
}

func NewButtonSelectStep(UIs *EffectWithUI, shows ...string) *ButtonSelectStep {
	return &ButtonSelectStep{UIs: UIs, Init0: false, Buttons: NewButtons(shows...)}
}

//===================SelectCancelStep=====================

type SelectCancelStep struct {
}

func NewSelectCancelStep() *SelectCancelStep {
	return &SelectCancelStep{}
}

func (s *SelectCancelStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextCancel {
		extra.Index = MaxIndex
	} else {
		extra.Index++
	}
}

//====================CiXiongShuangGuJianReqStep判断用户选择===================

type CiXiongShuangGuJianAskStep struct {
}

func (c *CiXiongShuangGuJianAskStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextConfirm {
		extra.Result1 = &Event{Type: EventAskCard, Src: event.Src, Desc: event.Desc, AskNum: 1, Filter: c.AnyFilter}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	} else if extra.Select == TextCancel {
		extra.Index = MaxIndex
	}
}

func (c *CiXiongShuangGuJianAskStep) AnyFilter(card *Card) bool {
	return true
}

func NewCiXiongShuangGuJianReqStep() *CiXiongShuangGuJianAskStep {
	return &CiXiongShuangGuJianAskStep{}
}

//=======================CiXiongShuangGuJianCheckSkill检测对面选择===========================

type CiXiongShuangGuJianCheckSkill struct {
}

func (c *CiXiongShuangGuJianCheckSkill) Update(event *Event, extra *StepExtra) {
	temp := extra.Result1
	if len(temp.Resps) == 1 { // 对面选择弃牌
		event.Desc.RemoveCard(temp.Resps...)
		MainGame.AddToDesktop(temp.Resps...)
		MainGame.DiscardFromDesktop(temp.Resps...)
	} else { // 对面不弃牌
		event.Src.DrawCard(1)
	}
	extra.Index = MaxIndex
}

func NewCiXiongShuangGuJianCheckSkill() *CiXiongShuangGuJianCheckSkill {
	return &CiXiongShuangGuJianCheckSkill{}
}

//======================BotAskCardStep===================

type BotAskCardStep struct {
	Timer int
}

func (b *BotAskCardStep) Update(event *Event, extra *StepExtra) {
	if b.Timer > 0 {
		b.Timer--
	} else {
		// 收集全部的，会尽量给到满足，否则一个也不给
		cards := Map(event.Desc.Cards, func(item *CardUI) *Card {
			return item.Card
		})
		if event.WithEquip {
			for _, equip := range event.Desc.Equips {
				cards = append(cards, equip.Card)
			}
		}
		for _, card := range cards {
			if event.Filter(card) {
				event.Resps = append(event.Resps, card)
				if len(event.Resps) >= event.AskNum { // 收集完了
					extra.Index = MaxIndex
					event.Abort = true
					return
				}
			}
		}
		event.Resps = nil // 还是不够，全部抛弃
		extra.Index = MaxIndex
	}
}

func NewBotAskCardStep() *BotAskCardStep {
	return &BotAskCardStep{Timer: BotTimer}
}

//==================PlayerAskCardStep用户在这里肯定是选牌了，放到对应位置就行了==================

type PlayerAskCardStep struct {
}

func (p *PlayerAskCardStep) Update(event *Event, extra *StepExtra) {
	event.Resps = extra.Cards
	event.Abort = true
	extra.Index = MaxIndex
}

func NewPlayerAskCardStep() *PlayerAskCardStep {
	return &PlayerAskCardStep{}
}

//=====================QingLongYanYueDaoCheckStep这里是已经选择出杀了=====================

type QingLongYanYueDaoCheckStep struct {
}

func NewQingLongYanYueDaoCheckStep() *QingLongYanYueDaoCheckStep {
	return &QingLongYanYueDaoCheckStep{}
}

func (q *QingLongYanYueDaoCheckStep) Update(event *Event, extra *StepExtra) {
	player := event.Desc
	card := extra.Cards[0] // 这个就是「杀」   这里传递的不是出牌阶段的StepExtra，出杀次数不受影响
	temp := &Event{Type: EventUseCard, Src: player, Descs: []*Player{event.Src}, Card: NewSimpleCardWrap(card), StepExtra: extra}
	effect := card.Skill.CreateEffect(temp)
	MainGame.PushAction(NewEffectGroup(temp, []IEffect{effect}))
	player.RemoveCard(card)
	MainGame.AddToDesktop(card)
	extra.Index = MaxIndex
}

//========================HanBingJianCheckStep=========================

type HanBingJianCheckStep struct { // 到这里就是选了
}

func (h *HanBingJianCheckStep) Update(event *Event, extra *StepExtra) {
	event.ShaHit = false // 免伤
	event.Desc.RemoveCard(extra.Cards...)
	event.Desc.RemoveEquip(extra.Cards...)
	event.Desc.RemoveDelayKit(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	extra.Index = MaxIndex
}

func NewHanBingJianCheckStep() *HanBingJianCheckStep {
	return &HanBingJianCheckStep{}
}

//=====================BaGuaZhenCheckStep=======================

type BaGuaZhenCheckStep struct {
}

func (b *BaGuaZhenCheckStep) Update(event *Event, extra *StepExtra) {
	if IsRedSuit(extra.JudgeCard.Desc.Suit) {
		event.Resp = NewVirtualCardWrap(&Card{Name: "闪", Point: PointNone, Suit: SuitNone, Type: CardBasic})
		MainGame.AddToDesktopRaw(event.Resp)                        // 也展示一下
		event.Abort = true                                          // 已经响应了，进行终止
		MainGame.DiscardFromDesktopRaw(extra.JudgeCard, event.Resp) // 可以丢弃了判定牌了
		MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: event.Desc, Desc: event.Src, Card: event.Card, Event: event})
	}
	extra.Index = MaxIndex
}

func NewBaGuaZhenCheckStep() *BaGuaZhenCheckStep {
	return &BaGuaZhenCheckStep{}
}

//=====================RenWangDunCheckStep=====================

type RenWangDunCheckStep struct {
}

func (r *RenWangDunCheckStep) Update(event *Event, extra *StepExtra) {
	if IsBlackSuit(event.Card.Desc.Suit) {
		event.Abort = true
		event.Invalid = true // 该结算过程无效，也没有后续事件了
	}
	extra.Index = MaxIndex
}

func NewRenWangDunCheckStep() *RenWangDunCheckStep {
	return &RenWangDunCheckStep{}
}

//====================LoopTriggerUseKitStep轮流触发响应无懈可击======================

type LoopTriggerUseKitStep struct {
	Players []*Player // 从事件源开始的列表
	Index   int
}

func (c *LoopTriggerUseKitStep) Update(event *Event, extra *StepExtra) {
	if c.Index < len(c.Players) {
		extra.Result1 = &Event{Type: EventRespCard, Src: event.Src, Card: event.Card, Desc: c.Players[c.Index],
			WrapFilter: c.WuXieKeJiFilter}
		MainGame.TriggerEvent(extra.Result1)
		c.Index++
		extra.Index++ // 查看要牌的结果
	} else {
		c.Index = 0      // 复原属性，可能会重用的
		extra.Index += 2 // 结束要牌
	}
	MainGame.TriggerEvent(&Event{})
}

func (c *LoopTriggerUseKitStep) WuXieKeJiFilter(card *CardWrap) bool {
	return card.Desc.Name == "无懈可击"
}

func NewLoopTriggerUseKitStep(src *Player) *LoopTriggerUseKitStep {
	return &LoopTriggerUseKitStep{Players: MainGame.GetSortPlayer(src), Index: 0}
}

//================CheckRespKitStep=================

type CheckRespKitStep struct {
}

func (c *CheckRespKitStep) Update(event *Event, extra *StepExtra) {
	resp := extra.Result1.Resp
	if resp != nil { // 有响应，触发对应效果
		// 这里使用的牌的目标是一张牌，暂时没有传递
		temp := &Event{Type: EventUseCard, Src: extra.Desc, Card: resp, Event: event}
		// 无懈可击比较特殊，虽然是响应但是还是要触发效果，严格来说这个效果并不是无懈可击的
		// 而是这个效果放在无懈可击上比较方便也可以独立存在，在这里再触发一下，要保证效果不会为空
		effect := resp.Desc.Skill.CreateEffect(temp)
		MainGame.PushAction(NewEffectGroup(temp, []IEffect{effect}))
		extra.Index++ // 只要有一人响应就可以提前跳出循环了
	} else {
		extra.Index-- // 当前用户没有响应，下一个用户
	}
}

func NewCheckRespKitStep() *CheckRespKitStep {
	return &CheckRespKitStep{}
}

//==================WuZhongShengYouStep无中生有核心效果=====================

type WuZhongShengYouStep struct {
}

func (w *WuZhongShengYouStep) Update(event *Event, extra *StepExtra) {
	if !event.Invalid {
		event.Src.DrawCard(2)
	}
	MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
	extra.Index = MaxIndex
}

func NewWuZhongShengYouStep() *WuZhongShengYouStep {
	return &WuZhongShengYouStep{}
}

//=======================WuXieKeJiStep========================

type WuXieKeJiStep struct {
}

func NewWuXieKeJiStep() *WuXieKeJiStep {
	return &WuXieKeJiStep{}
}

func (w *WuXieKeJiStep) Update(event *Event, extra *StepExtra) {
	if !event.Invalid { // 自己没有失效就让上一张牌失效  这里是响应已经弃了
		event.Event.Invalid = true
	}
	extra.Index = MaxIndex
}

//=====================CardCheckInvalidStep通用判断是否是否生效=======================

type CardCheckInvalidStep struct {
}

func (k *CardCheckInvalidStep) Update(event *Event, extra *StepExtra) {
	if event.Invalid { // 违规直接结束
		MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
		extra.Index = MaxIndex
	} else { // 否则下一步
		extra.Index++
	}
}

func NewCardCheckInvalidStep() *CardCheckInvalidStep {
	return &CardCheckInvalidStep{}
}

//================GuoHeChaiQiaoStep====================

type GuoHeChaiQiaoStep struct {
}

func (g *GuoHeChaiQiaoStep) Update(event *Event, extra *StepExtra) {
	event.Descs[0].RemoveCard(extra.Cards...)
	event.Descs[0].RemoveEquip(extra.Cards...)
	event.Descs[0].RemoveDelayKit(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
	extra.Index = MaxIndex
}

func NewGuoHeChaiQiaoStep() *GuoHeChaiQiaoStep {
	return &GuoHeChaiQiaoStep{}
}

//===================ShunShouQianYangStep======================

type ShunShouQianYangStep struct {
}

func (s *ShunShouQianYangStep) Update(event *Event, extra *StepExtra) {
	event.Descs[0].RemoveCard(extra.Cards...)
	event.Descs[0].RemoveEquip(extra.Cards...)
	event.Descs[0].RemoveDelayKit(extra.Cards...)
	event.Src.AddCard(extra.Cards...)
	MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
	extra.Index = MaxIndex
}

func NewShunShouQianYangStep() *ShunShouQianYangStep {
	return &ShunShouQianYangStep{}
}

//======================JieDaoShaRenAskStep======================

type JieDaoShaRenAskStep struct {
}

func (j *JieDaoShaRenAskStep) Update(event *Event, extra *StepExtra) {
	if event.Invalid {
		MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
		extra.Index = MaxIndex
	} else {
		extra.Result1 = &Event{Type: EventAskCard, Src: event.Src, Desc: event.Descs[0], AskNum: 1, Filter: j.ShaFilter}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	}
}

func (j *JieDaoShaRenAskStep) ShaFilter(card *Card) bool {
	return card.Name == "杀" // 这里应该可以要虚拟牌的，至少借刀杀人是可以的
}

func NewJieDaoShaRenAskStep() *JieDaoShaRenAskStep {
	return &JieDaoShaRenAskStep{}
}

//===============JieDaoShaRenCheckStep=================

type JieDaoShaRenCheckStep struct {
}

func NewJieDaoShaRenCheckStep() *JieDaoShaRenCheckStep {
	return &JieDaoShaRenCheckStep{}
}

func (j *JieDaoShaRenCheckStep) Update(event *Event, extra *StepExtra) {
	first := event.Descs[0]
	last := event.Descs[1]
	if len(extra.Result1.Resps) == 1 { // 对面要杀了
		card := extra.Result1.Resps[0]
		temp := &Event{Type: EventUseCard, Src: first, Descs: []*Player{last}, Card: NewSimpleCardWrap(card), StepExtra: extra}
		effect := card.Skill.CreateEffect(temp)
		MainGame.PushAction(NewEffectGroup(temp, []IEffect{effect}))
		first.RemoveCard(card)
		MainGame.AddToDesktop(card)
	} else { // 对面不杀拿装备
		equip := first.Equips[EquipWeapon]
		first.RemoveEquip(equip.Card)
		event.Src.AddCard(equip.Card)
	}
	extra.Index++
}

//==================CardEndStep===================

type CardEndStep struct { // 有些事件为了保持执行顺序，最后一个 step 若是创建了事件不能直接结束必须等一下，否则弃牌堆处理会出问题
}

func (c *CardEndStep) Update(event *Event, extra *StepExtra) {
	MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
	extra.Index = MaxIndex
}

func NewCardEndStep() *CardEndStep {
	return &CardEndStep{}
}

//=================WuGuFengDengPrepareStep=================

type WuGuFengDengPrepareStep struct {
	Descs []*Player
	Index int
}

func (w *WuGuFengDengPrepareStep) Update(event *Event, extra *StepExtra) {
	if w.Index < len(w.Descs) {
		event.Desc = w.Descs[w.Index] // 以某人为目标
		extra.Index++
		w.Index++
	} else { // 结算完毕 丢弃没人要的牌
		MainGame.AddToDesktop(event.Cards...)
		MainGame.DiscardFromDesktop(event.Cards...)
		MainGame.DiscardFromDesktopRaw(event.Card)
		extra.Index = MaxIndex
	}
}

func NewWuGuFengDengPrepareStep(src *Player) *WuGuFengDengPrepareStep {
	return &WuGuFengDengPrepareStep{Descs: MainGame.GetSortPlayer(src), Index: 0}
}

//==================WuGuFengDengChooseStep===================

type WuGuFengDengChooseStep struct {
}

func (w *WuGuFengDengChooseStep) Update(event *Event, extra *StepExtra) {
	if !event.Invalid { // 触发选牌
		extra.Result1 = &Event{Type: EventChooseCard, Src: event.Src, Desc: event.Desc, Card: event.Card,
			Cards: event.Cards, ChooseMax: 1, ChooseMin: 1}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	} else {
		extra.Index = 0
		event.Invalid = false // 被前面复用了注意
	}
}

func NewWuGuFengDengChooseStep() *WuGuFengDengChooseStep {
	return &WuGuFengDengChooseStep{}
}

//===============WuGuFengDengExecuteStep==================

type WuGuFengDengExecuteStep struct {
}

func (w *WuGuFengDengExecuteStep) Update(event *Event, extra *StepExtra) {
	event.Cards = SubSlice(event.Cards, extra.Result1.Resps)
	event.Desc.AddCard(extra.Result1.Resps...)
	extra.Index = 0
}

func NewWuGuFengDengExecuteStep() *WuGuFengDengExecuteStep {
	return &WuGuFengDengExecuteStep{}
}

//====================BotChooseCardStep===================

type BotChooseCardStep struct {
	Timer int
}

func (b *BotChooseCardStep) Update(event *Event, extra *StepExtra) {
	if b.Timer > 0 {
		b.Timer--
	} else {
		for _, card := range event.Cards { // 尽量选一下
			if len(event.Resps) >= event.ChooseMin {
				break
			}
			event.Resps = append(event.Resps, card)
		}
		extra.Index = MaxIndex
	}
}

func NewBotChooseCardStep() *BotChooseCardStep {
	return &BotChooseCardStep{Timer: BotTimer}
}

//==================ChooseNumCardStep===================

type ChooseNumCardStep struct { // 选择固定数量卡牌的步骤
	Min, Max   int // 选择固定数量的牌
	UIs        *EffectWithUI
	Init0      bool
	Buttons    []*Button
	ChooseCard *ChooseCard
}

func NewChooseNumCardStep(min int, max int, UIs *EffectWithUI, chooseCard *ChooseCard) *ChooseNumCardStep {
	return &ChooseNumCardStep{Min: min, Max: max, UIs: UIs, Buttons: make([]*Button, 0), ChooseCard: chooseCard}
}

// 「确定」->把选择的卡牌带到下一步(必须数量足够)
// 「取消」->有选择取消选择
// 点击卡牌->选择卡牌
func (s *ChooseNumCardStep) Update(event *Event, extra *StepExtra) {
	s.Init()
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	// 检验按钮
	if s.HandleBtnClick(x, y, extra) {
		return
	}
	// 检验选择牌
	if s.ChooseCard.ToggleCard(x, y) {
		if len(s.ChooseCard.GetSelectCard()) < s.Max {
			s.ChooseCard.SetAllCanSelect()
		} else {
			s.ChooseCard.DarkLastCard()
		}
	}
}

func (s *ChooseNumCardStep) HandleBtnClick(x, y float32, extra *StepExtra) bool {
	for _, button := range s.Buttons {
		if button.Click(x, y) {
			if button.Show == TextConfirm {
				cards := s.ChooseCard.GetSelectCard()
				if len(cards) >= s.Min && len(cards) <= s.Max {
					extra.Cards = cards
					extra.Index++
				}
			} else if button.Show == TextCancel {
				s.ChooseCard.Reset()
			}
			return true
		}
	}
	return false
}

func (s *ChooseNumCardStep) Init() {
	if s.Init0 {
		return
	}
	s.Init0 = true
	s.Buttons = NewButtons(TextConfirm, TextCancel)
	for _, button := range s.Buttons {
		s.UIs.UIs = append(s.UIs.UIs, button)
	}
	s.UIs.UIs = append(s.UIs.UIs, s.ChooseCard)
}

//=======================PlayerChooseCardStep======================

type PlayerChooseCardStep struct {
}

func (p *PlayerChooseCardStep) Update(event *Event, extra *StepExtra) {
	event.Resps = extra.Cards
	extra.Index = MaxIndex
}

func NewPlayerChooseCardStep() *PlayerChooseCardStep {
	return &PlayerChooseCardStep{}
}

//===================JueDouCheckStep=======================

type JueDouCheckStep struct {
}

func (j *JueDouCheckStep) Update(event *Event, extra *StepExtra) {
	if event.Invalid {
		extra.Index = MaxIndex
		MainGame.DiscardFromDesktopRaw(event.Card)
	} else {
		extra.Desc = event.Descs[0]
		extra.Result1 = &Event{Type: EventRespCard, Src: event.Src, Card: event.Card, Desc: extra.Desc,
			WrapFilter: j.ShaFilter, HurtVal: 1}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	}
}

func (j *JueDouCheckStep) ShaFilter(card *CardWrap) bool {
	return card.Desc.Name == "杀"
}

func NewJueDouCheckStep() *JueDouCheckStep {
	return &JueDouCheckStep{}
}

//==================JueDouExecuteStep====================

type JueDouExecuteStep struct {
}

func (j *JueDouExecuteStep) Update(event *Event, extra *StepExtra) {
	if extra.Result1.Resp != nil { // 继续拼杀
		if extra.Desc == event.Src {
			extra.Desc = event.Descs[0]
		} else {
			extra.Desc = event.Src
		}
		extra.Result1 = &Event{Type: EventRespCard, Src: event.Src, Card: event.Card, Desc: extra.Desc,
			WrapFilter: j.ShaFilter, HurtVal: 1}
		MainGame.TriggerEvent(extra.Result1)
	} else { // 一方败了
		if extra.Desc.ChangeHp(-extra.Result1.HurtVal) {
			MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Src: event.Src, Desc: extra.Desc, Card: event.Card, HurtVal: extra.Result1.HurtVal})
		} else {
			MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Src: event.Src, Desc: extra.Desc, Card: event.Card, HurtVal: extra.Result1.HurtVal})
		}
		extra.Index = MaxIndex
		MainGame.DiscardFromDesktopRaw(event.Card)
	}
}

func (j *JueDouExecuteStep) ShaFilter(card *CardWrap) bool {
	return card.Desc.Name == "杀"
}

func NewJueDouExecuteStep() *JueDouExecuteStep {
	return &JueDouExecuteStep{}
}

//=================AoePrepareStep===================

type AoePrepareStep struct { // 南蛮，万剑 通用组件
	Descs []*Player // 排除玩家逆时针
	Index int
}

func (n *AoePrepareStep) Update(event *Event, extra *StepExtra) {
	if n.Index < len(n.Descs) {
		event.Desc = n.Descs[n.Index]
		extra.Index++
		n.Index++
	} else {
		MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
		extra.Index = MaxIndex
	}
}

func NewAoePrepareStep(src *Player) *AoePrepareStep {
	return &AoePrepareStep{Descs: MainGame.GetSortPlayer(src)[1:], Index: 0}
}

//===============AoeRespStep=================

type AoeRespStep struct {
	CardFilter CardWrapFilter
}

func NewAoeRespStep(cardFilter CardWrapFilter) *AoeRespStep {
	return &AoeRespStep{CardFilter: cardFilter}
}

func (a *AoeRespStep) Update(event *Event, extra *StepExtra) {
	if event.Invalid { // 无懈可击了
		extra.Index = 0
		event.Invalid = false
	} else {
		extra.Result1 = &Event{Type: EventRespCard, Src: event.Src, Card: event.Card, Desc: event.Desc,
			WrapFilter: a.CardFilter}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	}
}

//==============AoeExecuteStep=================

type AoeExecuteStep struct {
}

func (a *AoeExecuteStep) Update(event *Event, extra *StepExtra) {
	if extra.Result1.Resp == nil { // 没有响应
		if event.Desc.ChangeHp(-event.HurtVal) {
			MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Src: event.Src, Desc: event.Desc, Card: event.Card, HurtVal: event.HurtVal})
		} else {
			MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Src: event.Src, Desc: event.Desc, Card: event.Card, HurtVal: event.HurtVal})
		}
	}
	extra.Index = 0
}

func NewAoeExecuteStep() *AoeExecuteStep {
	return &AoeExecuteStep{}
}

//===============TaoYuanJieYiPrepareStep=================

type TaoYuanJieYiPrepareStep struct {
	Descs []*Player
	Index int
}

func (t *TaoYuanJieYiPrepareStep) Update(event *Event, extra *StepExtra) {
	if t.Index < len(t.Descs) {
		event.Desc = t.Descs[t.Index]
		extra.Index++
		t.Index++
	} else {
		MainGame.DiscardFromDesktopRaw(event.Card) // 结算完毕
		extra.Index = MaxIndex
	}
}

func NewTaoYuanJieYiPrepareStep(src *Player) *TaoYuanJieYiPrepareStep {
	return &TaoYuanJieYiPrepareStep{Descs: MainGame.GetSortPlayer(src), Index: 0}
}

//=================TaoYuanJieYiExecuteStep===================

type TaoYuanJieYiExecuteStep struct {
}

func (t *TaoYuanJieYiExecuteStep) Update(event *Event, extra *StepExtra) {
	if event.Invalid {
		event.Invalid = false
	} else {
		event.Desc.ChangeHp(1)
	}
	extra.Index = 0
}

func NewTaoYuanJieYiExecuteStep() *TaoYuanJieYiExecuteStep {
	return &TaoYuanJieYiExecuteStep{}
}

//====================DelayKitExecuteStep=======================

type DelayKitExecuteStep struct {
}

func (d *DelayKitExecuteStep) Update(event *Event, extra *StepExtra) {
	player := event.Src
	if len(event.Descs) == 1 {
		player = event.Descs[0]
	} // 延时锦囊会放到判定区而不是弃牌区
	card := event.Card
	MainGame.RemoveFromDesktopRaw(card)
	player.AddDelayKit(card)
	extra.Index = MaxIndex
}

func NewDelayKitExecuteStep() *DelayKitExecuteStep {
	return &DelayKitExecuteStep{}
}

//==================LeBuSiShuStep=====================

type LeBuSiShuStep struct {
}

func (l *LeBuSiShuStep) Update(event *Event, extra *StepExtra) {
	if extra.JudgeCard.Desc.Suit != SuitHeart {
		event.StageExtra.SkipStage |= StagePlay // 跳出牌阶段
	}
	extra.Index = MaxIndex
	MainGame.DiscardFromDesktopRaw(event.Card, extra.JudgeCard)
}

func NewLeBuSiShuStep() *LeBuSiShuStep {
	return &LeBuSiShuStep{}
}

//====================ShanDianStep======================

type ShanDianStep struct {
}

func NewShanDianStep() *ShanDianStep {
	return &ShanDianStep{}
}

func (s *ShanDianStep) Update(event *Event, extra *StepExtra) {
	card := extra.JudgeCard.Desc
	if card.Suit == SuitSpade && card.Point >= Point2 && card.Point <= Point9 {
		if event.Desc.ChangeHp(-3) {
			MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Desc: event.Desc, Card: event.Card, HurtVal: 3})
		} else {
			MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Desc: event.Desc, Card: event.Card, HurtVal: 3})
		}
		MainGame.DiscardFromDesktopRaw(event.Card, extra.JudgeCard) // 应该在下一个 step伤害事件还没有结算完毕
	} else {
		MainGame.RemoveFromDesktopRaw(event.Card, extra.JudgeCard) // 没有生效给一个人
		next := MainGame.GetNextPlayer()
		next.AddDelayKit(event.Card)
	}
	extra.Index = MaxIndex
}

//===============ShanDianInvalidStep===============

type ShanDianInvalidStep struct {
}

func NewShanDianInvalidStep() *ShanDianInvalidStep {
	return &ShanDianInvalidStep{}
}

func (k *ShanDianInvalidStep) Update(event *Event, extra *StepExtra) {
	if event.Invalid { // 违规直接结束
		MainGame.RemoveFromDesktopRaw(event.Card) // 没有生效给一个人
		next := MainGame.GetNextPlayer()
		next.AddDelayKit(event.Card)
		extra.Index = MaxIndex
	} else { // 否则下一步
		extra.Index++
	}
}

//=====================PlayerDyingCheckStep====================

type PlayerDyingCheckStep struct {
}

func (p *PlayerDyingCheckStep) Update(event *Event, extra *StepExtra) {
	if event.Desc.Hp < 1 {
		event.Descs = MainGame.GetSortPlayer(event.Desc)
		extra.Index++
	} else {
		extra.Index = MaxIndex
	}
}

func NewPlayerDyingCheckStep() *PlayerDyingCheckStep {
	return &PlayerDyingCheckStep{}
}

//====================PlayerDyingLoopStep=======================

type PlayerDyingLoopStep struct {
	Index int
}

func (p *PlayerDyingLoopStep) Update(event *Event, extra *StepExtra) {
	if p.Index < len(event.Descs) {
		extra.Result1 = &Event{Type: EventAskCard, Src: event.Desc, Desc: event.Descs[p.Index], AskNum: 1,
			Filter: p.TaoCardFilter, RecoverVal: 1}
		MainGame.TriggerEvent(extra.Result1)
		p.Index++
		extra.Index++
	} else {
		MainGame.TriggerEvent(&Event{Type: EventPlayerDie, Src: event.Src, Desc: event.Desc})
		extra.Index = MaxIndex
	}
}

func (p *PlayerDyingLoopStep) TaoCardFilter(card *Card) bool {
	return card.Name == "桃"
}

func NewPlayerDyingLoopStep() *PlayerDyingLoopStep {
	return &PlayerDyingLoopStep{Index: 0}
}

//=====================PlayerDyingResStep======================

type PlayerDyingResStep struct {
}

func (p *PlayerDyingResStep) Update(event *Event, extra *StepExtra) {
	res := extra.Result1
	if len(res.Resps) > 0 { // 出桃了
		res.Desc.RemoveCard(res.Resps...)
		MainGame.AddToDesktop(res.Resps...)
		MainGame.DiscardFromDesktop(res.Resps...)
		event.Desc.ChangeHp(res.RecoverVal)
		extra.Index = 0 // 出现一个就可以结束循环了
	} else {
		extra.Index--
	}
}

func NewPlayerDyingResStep() *PlayerDyingResStep {
	return &PlayerDyingResStep{}
}

//=====================SysPlayerDieStep=======================

type SysPlayerDieStep struct {
}

func (s *SysPlayerDieStep) Update(event *Event, extra *StepExtra) {
	desc := event.Desc
	// 处理死亡
	desc.IsDie = true
	cards := make([]*Card, 0)
	cards = append(cards, desc.GetEquips()...)
	cards = append(cards, desc.GetDelayKits()...)
	cards = append(cards, desc.GetCards()...)
	desc.RemoveEquip(cards...)
	desc.RemoveDelayKit(cards...)
	desc.RemoveCard(cards...)
	MainGame.AddToDesktop(cards...)
	MainGame.DiscardFromDesktop(cards...)
	// 处理奖惩
	src := event.Src
	if src != nil {
		if desc.Role == RoleFanZei {
			src.DrawCard(3)
		} else if desc.Role == RoleZhongChen && src.Role == RoleZhuGong {
			// 主公需要弃置所有装备与手牌
			cards = make([]*Card, 0)
			cards = append(cards, src.GetEquips()...)
			cards = append(cards, src.GetCards()...)
			src.RemoveEquip(cards...)
			src.RemoveCard(cards...)
			MainGame.AddToDesktop(cards...)
			MainGame.DiscardFromDesktop(cards...)
		}
	}
	MainGame.TriggerEvent(&Event{Type: EventGameOverCheck, Src: desc})
	extra.Index = MaxIndex
}

func NewSysPlayerDieStep() *SysPlayerDieStep {
	return &SysPlayerDieStep{}
}

//===================SysGameOverStep======================

type SysGameOverStep struct {
	UIs   *EffectWithUI
	Init0 bool
	Info  string
}

func (b *SysGameOverStep) Update(event *Event, extra *StepExtra) {
	if b.Init0 {
		return
	}
	b.Init0 = true
	b.UIs.UIs = append(b.UIs.UIs, NewGameOver(b.Info))
}

func NewSysGameOverStep(UIs *EffectWithUI, info string) *SysGameOverStep {
	return &SysGameOverStep{UIs: UIs, Init0: false, Info: info}
}

//====================JianXiongStep=====================

type JianXiongStep struct {
}

func (j *JianXiongStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextConfirm {
		event.Desc.AddCard(event.Card.Src...)
		event.Card.Src = make([]*Card, 0)
	}
	extra.Index = MaxIndex
}

func NewJianXiongStep() *JianXiongStep {
	return &JianXiongStep{}
}

//====================HuJiaLoopStep======================

type HuJiaLoopStep struct {
	Players []*Player
	Index   int
}

func (h *HuJiaLoopStep) Update(event *Event, extra *StepExtra) {
	if h.Index < len(h.Players) {
		extra.Result1 = &Event{Type: EventAskCard, Src: event.Desc, Desc: h.Players[h.Index], AskNum: 1, Filter: h.ShanFilter}
		MainGame.TriggerEvent(extra.Result1)
		h.Index++
		extra.Index++
	} else { // 没人打闪
		extra.Index = MaxIndex
	}
}

func (h *HuJiaLoopStep) ShanFilter(card *Card) bool {
	return card.Name == "闪"
}

func NewHuJiaLoopStep(src *Player) *HuJiaLoopStep {
	return &HuJiaLoopStep{Players: MainGame.GetPlayers(func(player *Player) bool {
		return player != src && player.Force == ForceWei
	}), Index: 0}
}

//===================HuJiaCheckStep===================

type HuJiaCheckStep struct {
}

func (h *HuJiaCheckStep) Update(event *Event, extra *StepExtra) {
	res := extra.Result1
	if len(res.Resps) > 0 {
		res.Desc.RemoveCard(res.Resps...)
		MainGame.AddToDesktop(res.Resps...)
		MainGame.DiscardFromDesktop(res.Resps...)
		event.Resp = NewSimpleCardWrap(extra.Result1.Resps[0])
		event.Abort = true
		MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: event.Desc, Desc: event.Src, Card: event.Card, Event: event})
		extra.Index = MaxIndex
	} else {
		extra.Index--
	}
}

func NewHuJiaCheckStep() *HuJiaCheckStep {
	return &HuJiaCheckStep{}
}

//================FanKuiStep=================

type FanKuiStep struct {
}

func (f *FanKuiStep) Update(event *Event, extra *StepExtra) {
	event.Src.RemoveCard(extra.Cards...)
	event.Src.RemoveEquip(extra.Cards...)
	event.Src.RemoveDelayKit(extra.Cards...)
	event.Desc.AddCard(extra.Cards...)
	extra.Index = MaxIndex
}

func NewFanKuiStep() *FanKuiStep {
	return &FanKuiStep{}
}

//==================GuiCaiStep==================

type GuiCaiStep struct {
	Player *Player
}

func NewGuiCaiStep(player *Player) *GuiCaiStep {
	return &GuiCaiStep{Player: player}
}

func (g *GuiCaiStep) Update(event *Event, extra *StepExtra) {
	temp := event.StepExtra // 丢弃原始判定牌
	MainGame.RemoveFromDesktopRaw(temp.JudgeCard)
	MainGame.DiscardCard(temp.JudgeCard.Src) // 设置新判定牌
	temp.JudgeCard = NewSimpleCardWrap(extra.Cards[0])
	g.Player.RemoveCard(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	extra.Index = MaxIndex
}

//==================GangLieCheckStep====================

type GangLieCheckStep struct {
}

func (g *GangLieCheckStep) Update(event *Event, extra *StepExtra) {
	if extra.JudgeCard.Desc.Suit == SuitHeart {
		extra.Index = MaxIndex
	} else {
		extra.Result1 = &Event{Type: EventAskCard, Src: event.Desc, Desc: event.Src, AskNum: 2, Filter: g.AnyFilter}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	}
}

func (g *GangLieCheckStep) AnyFilter(card *Card) bool {
	return true
}

func NewGangLieCheckStep() *GangLieCheckStep {
	return &GangLieCheckStep{}
}

//=================GangLieExecuteStep==================

type GangLieExecuteStep struct {
}

func (g *GangLieExecuteStep) Update(event *Event, extra *StepExtra) {
	res := extra.Result1
	if len(res.Resps) == 2 {
		event.Src.RemoveCard(res.Resps...)
		MainGame.AddToDesktop(res.Resps...)
		MainGame.DiscardFromDesktop(res.Resps...)
	} else {
		if extra.Desc.ChangeHp(-1) {
			MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Src: event.Desc, Desc: event.Src, HurtVal: 1})
		} else {
			MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Src: event.Desc, Desc: event.Src, HurtVal: 1})
		}
	}
	extra.Index = MaxIndex
}

func NewGangLieExecuteStep() *GangLieExecuteStep {
	return &GangLieExecuteStep{}
}

//================SelectPlayerStep==================

type SelectPlayerStep struct {
	Min, Max int
	Filter   PlayerFilter
	UIs      *EffectWithUI
	Init0    bool
	Buttons  []*Button
}

func (s *SelectPlayerStep) Update(event *Event, extra *StepExtra) {
	s.Init()
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	} // 点击按钮
	if s.HandleBtnClick(x, y, extra) {
		return
	}
	// 处理玩家选择
	if MainGame.TogglePlayer(x, y) {
		if len(MainGame.GetSelectPlayer()) < s.Max {
			for _, player := range MainGame.Players {
				if !player.Select {
					player.CanSelect = s.Filter(player)
				}
			}
		} else {
			MainGame.DarkLastPlayer()
		}
	}
}

func (s *SelectPlayerStep) Init() {
	if s.Init0 {
		return
	}
	s.Init0 = true
	s.Buttons = NewButtons(TextConfirm, TextCancel)
	for _, button := range s.Buttons {
		s.UIs.UIs = append(s.UIs.UIs, button)
	}
}

// 确定 把选择的玩家向下传递
// 取消 取消 player 的选择
func (s *SelectPlayerStep) HandleBtnClick(x float32, y float32, extra *StepExtra) bool {
	for _, button := range s.Buttons {
		if button.Click(x, y) {
			if button.Show == TextConfirm {
				players := MainGame.GetSelectPlayer()
				if len(players) >= s.Min && len(players) <= s.Max {
					extra.Players = players
					extra.Index++
					MainGame.ResetPlayer()
				}
			} else if button.Show == TextCancel {
				MainGame.ResetPlayer()
				MainGame.CheckPlayerByFilter(s.Filter)
			}
			return true
		}
	}
	return false
}

func NewSelectPlayerStep(min, max int, filter PlayerFilter, uis *EffectWithUI) *SelectPlayerStep {
	MainGame.CheckPlayerByFilter(filter)
	return &SelectPlayerStep{Min: min, Max: max, Filter: filter, UIs: uis}
}

//==================TuXiLoopStep===================

type TuXiLoopStep struct {
	Index int
}

func (t *TuXiLoopStep) Update(event *Event, extra *StepExtra) {
	if t.Index < len(extra.Players) {
		// 因为每次都要使用新的信息，且这时才得知目标
		res := NewEffectWithUI()
		cards := extra.Players[t.Index].GetCards()
		res.SetSteps(NewSelectPlayerCardStep(1, res, NewAllCard(cards, nil, nil)), NewTuXiExecuteStep())
		MainGame.PushAction(NewEffectGroup(&Event{Src: event.Src, Desc: extra.Players[t.Index]}, []IEffect{res}))
		t.Index++
	} else {
		event.StepExtra.Index = MaxIndex
		extra.Index = MaxIndex
	}
}

func NewTuXiLoopStep() *TuXiLoopStep {
	return &TuXiLoopStep{}
}

//=============TuXiExecuteStep===============

type TuXiExecuteStep struct {
}

func (t *TuXiExecuteStep) Update(event *Event, extra *StepExtra) {
	event.Desc.RemoveCard(extra.Cards...)
	event.Src.AddCard(extra.Cards...)
	extra.Index = MaxIndex
}

func NewTuXiExecuteStep() *TuXiExecuteStep {
	return &TuXiExecuteStep{}
}

//==================LuoYiStep==================

type LuoYiStep struct {
}

func (l *LuoYiStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextConfirm { // 添加裸衣 buff
		event.Src.AddSkill(NewLuoYiBuffSkill(event.Src))
	}
	extra.Index = MaxIndex
}

func NewLuoYiStep() *LuoYiStep {
	return &LuoYiStep{}
}

//=================LuoYiBuffStep===================

type LuoYiBuffStep struct {
}

func NewLuoYiBuffStep() *LuoYiBuffStep {
	return &LuoYiBuffStep{}
}

func (l *LuoYiBuffStep) Update(event *Event, extra *StepExtra) {
	name := event.Card.Desc.Name
	if name == "杀" || name == "决斗" {
		event.HurtVal++
	}
	extra.Index = MaxIndex
}

//=================RemoveSkillStep====================

type RemoveSkillStep struct {
	Skill  ISkill
	Player *Player
}

func NewRemoveSkillStep(skill ISkill, player *Player) *RemoveSkillStep {
	return &RemoveSkillStep{Skill: skill, Player: player}
}

func (r *RemoveSkillStep) Update(event *Event, extra *StepExtra) {
	r.Player.RemoveSkill(r.Skill)
	extra.Index = MaxIndex
}

//====================TianDuStep====================

type TianDuStep struct {
}

func NewTianDuStep() *TianDuStep {
	return &TianDuStep{}
}

func (t *TianDuStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextConfirm {
		judgeCard := event.StepExtra.JudgeCard
		event.Src.AddCard(judgeCard.Src...)
		judgeCard.Src = make([]*Card, 0)
	}
	extra.Index = MaxIndex
}

//==================YiJiPrepareStep===================

type YiJiPrepareStep struct {
}

func (y *YiJiPrepareStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextConfirm {
		extra.Cards = MainGame.DrawCard(2)
		extra.Result1 = &Event{Type: EventChooseCard, Src: event.Desc, Desc: event.Desc, Cards: extra.Cards, ChooseMax: 2, ChooseMin: 0}
		MainGame.TriggerEvent(extra.Result1)
		extra.Index++
	} else {
		extra.Index = MaxIndex
	}
}

func NewYiJiPrepareStep() *YiJiPrepareStep {
	return &YiJiPrepareStep{}
}

//====================YiJiCheckStep========================

type YiJiCheckStep struct {
}

func (y *YiJiCheckStep) Update(event *Event, extra *StepExtra) {
	res := extra.Result1
	if len(res.Resps) > 0 {
		event.Desc.AddCard(SubSlice(extra.Cards, res.Resps)...) // 先拿自己的
		extra.Cards = res.Resps
		extra.Index++
	} else { // 不分了都是自己的
		event.Desc.AddCard(extra.Cards...)
		extra.Index = MaxIndex
	}
}

func NewYiJiCheckStep() *YiJiCheckStep {
	return &YiJiCheckStep{}
}

//====================YiJiExecuteStep======================

type YiJiExecuteStep struct {
}

func (y *YiJiExecuteStep) Update(event *Event, extra *StepExtra) {
	extra.Players[0].AddCard(extra.Cards...)
	extra.Index = MaxIndex
}

func NewYiJiExecuteStep() *YiJiExecuteStep {
	return &YiJiExecuteStep{}
}

//===================QingGuoStep======================

type QingGuoStep struct {
}

func (q *QingGuoStep) Update(event *Event, extra *StepExtra) {
	event.Resp = NewTransCardWrap(&Card{Name: "闪", Point: PointNone, Suit: SuitNone, Type: CardBasic}, extra.Cards)
	event.Desc.RemoveCard(extra.Cards...)
	MainGame.AddToDesktopRaw(event.Resp)       // 也展示一下
	event.Abort = true                         // 已经响应了，进行终止
	MainGame.DiscardFromDesktopRaw(event.Resp) // 可以丢弃了判定牌了
	MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: event.Desc, Desc: event.Src, Card: event.Card, Event: event})
}

func NewQingGuoStep() *QingGuoStep {
	return &QingGuoStep{}
}

//==================LuoShenStep=====================

type LuoShenStep struct {
}

func NewLuoShenStep() *LuoShenStep {
	return &LuoShenStep{}
}

func (l *LuoShenStep) Update(event *Event, extra *StepExtra) {
	judgeCard := extra.JudgeCard
	if IsBlackSuit(judgeCard.Desc.Suit) {
		event.Src.AddCard(judgeCard.Src...)
		MainGame.RemoveFromDesktopRaw(judgeCard)
		extra.Index = 0
	} else {
		MainGame.DiscardFromDesktopRaw(judgeCard)
		extra.Index = MaxIndex
	}
}

//===================RenDe=====================

type RenDeStep struct {
	RenDeSkill *RenDeSkill
}

func (r *RenDeStep) Update(event *Event, extra *StepExtra) {
	if r.RenDeSkill.CardNum < 2 && r.RenDeSkill.CardNum+len(extra.Cards) >= 2 {
		event.Src.ChangeHp(1)
	}
	r.RenDeSkill.CardNum += len(extra.Cards)
	event.Src.RemoveCard(extra.Cards...)
	extra.Players[0].AddCard(extra.Cards...)
	extra.Index = MaxIndex
}

func NewRenDeStep(renDeSkill *RenDeSkill) *RenDeStep {
	return &RenDeStep{RenDeSkill: renDeSkill}
}

//====================JiJiangLoopStep======================

type JiJiangLoopStep struct {
	Players []*Player
	Index   int
}

func (h *JiJiangLoopStep) Update(event *Event, extra *StepExtra) {
	if h.Index < len(h.Players) {
		extra.Result1 = &Event{Type: EventAskCard, Src: event.Desc, Desc: h.Players[h.Index], AskNum: 1, Filter: h.ShaFilter}
		MainGame.TriggerEvent(extra.Result1)
		h.Index++
		extra.Index++
	} else { // 没人打杀
		extra.Index = MaxIndex
	}
}

func (h *JiJiangLoopStep) ShaFilter(card *Card) bool {
	return card.Name == "杀"
}

func NewJiJiangLoopStep(src *Player) *JiJiangLoopStep {
	return &JiJiangLoopStep{Players: MainGame.GetPlayers(func(player *Player) bool {
		return player != src && player.Force == ForceShu
	}), Index: 0}
}

//===================JiJiangCheck1Step===================

type JiJiangCheck1Step struct {
}

func (h *JiJiangCheck1Step) Update(event *Event, extra *StepExtra) {
	res := extra.Result1
	if len(res.Resps) > 0 {
		res.Desc.RemoveCard(res.Resps...)
		MainGame.AddToDesktop(res.Resps...)
		MainGame.DiscardFromDesktop(res.Resps...)
		event.Resp = NewSimpleCardWrap(extra.Result1.Resps[0])
		event.Abort = true
		MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: event.Desc, Desc: event.Src, Card: event.Card, Event: event})
		extra.Index = MaxIndex
	} else {
		extra.Index--
	}
}

func NewJiJiangCheck1Step() *JiJiangCheck1Step {
	return &JiJiangCheck1Step{}
}

//=================SelectPlayerVerifyStep==================

type SelectPlayerVerifyStep struct {
}

func NewSelectPlayerVerifyStep() *SelectPlayerVerifyStep {
	return &SelectPlayerVerifyStep{}
}

func (j *SelectPlayerVerifyStep) Update(event *Event, extra *StepExtra) {
	if len(extra.Players) > 0 {
		extra.Index++
	} else { // 没有选人直接结束
		extra.Index = MaxIndex
	}
}

//===================JiJiangCheck2Step===================

type JiJiangCheck2Step struct {
}

func (h *JiJiangCheck2Step) Update(event *Event, extra *StepExtra) {
	res := extra.Result1
	if len(res.Resps) > 0 {
		card := res.Resps[0]
		res.Desc.RemoveCard(card)
		MainGame.AddToDesktop(card)
		temp := &Event{Type: EventUseCard, Src: event.Src, Descs: extra.Players, Card: NewSimpleCardWrap(card), StepExtra: event.StepExtra}
		effect := card.Skill.CreateEffect(temp)
		MainGame.PushAction(NewEffectGroup(temp, []IEffect{effect}))
		extra.Index = MaxIndex
	} else {
		extra.Index--
	}
}

func NewJiJiangCheck2Step() *JiJiangCheck2Step {
	return &JiJiangCheck2Step{}
}

//================WuShengCheckStep===================

type WuShengCheckStep struct {
}

func (w *WuShengCheckStep) Update(event *Event, extra *StepExtra) {
	event.Desc.RemoveCard(extra.Cards...)
	event.Desc.RemoveEquip(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	event.Resp = NewTransCardWrap(&Card{Name: "杀", Point: PointNone, Suit: SuitNone, Type: CardBasic}, extra.Cards)
	event.Abort = true
	MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: event.Desc, Desc: event.Src, Card: event.Card, Event: event})
	extra.Index = MaxIndex
}

func NewWuShengCheckStep() *WuShengCheckStep {
	return &WuShengCheckStep{}
}

//================TransCardStep===================

type TransCardStep struct {
	Desc *Card
}

func NewTransCardStep(desc *Card) *TransCardStep {
	return &TransCardStep{Desc: desc}
}

func (w *TransCardStep) Update(event *Event, extra *StepExtra) {
	card := NewTransCardWrap(w.Desc, extra.Cards)
	event.Src.RemoveCard(card.Src...)
	event.Src.RemoveEquip(card.Src...)
	MainGame.AddToDesktopRaw(card)
	temp := &Event{Type: EventUseCard, Src: event.Src, Descs: extra.Players, Card: card, StepExtra: event.StepExtra}
	effect := card.Desc.Skill.CreateEffect(temp)
	MainGame.PushAction(NewEffectGroup(temp, []IEffect{effect}))
	extra.Index = MaxIndex
}

//====================GuanXingStep=====================

type GuanXingStep struct {
	UIs      *EffectWithUI
	Init0    bool
	Buttons  []*Button
	GuanXing *GuanXing
}

func (g *GuanXingStep) Update(event *Event, extra *StepExtra) {
	g.Init()
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	} // 点击按钮
	if g.HandleBtnClick(x, y, extra) {
		return
	} // 点击卡牌
	g.GuanXing.ToggleCard(x, y)
}

func (g *GuanXingStep) Init() {
	if g.Init0 {
		return
	}
	g.Init0 = true
	g.Buttons = NewButtons(TextConfirm, TextCancel)
	for _, button := range g.Buttons {
		g.UIs.UIs = append(g.UIs.UIs, button)
	}
	g.UIs.UIs = append(g.UIs.UIs, g.GuanXing)
}

// 确定，释放Up Down牌
// 取消，释放原始牌
func (g *GuanXingStep) HandleBtnClick(x float32, y float32, extra *StepExtra) bool {
	for _, button := range g.Buttons {
		if button.Click(x, y) {
			if button.Show == TextConfirm {
				MainGame.AddCardToTop(g.GuanXing.GetUpCards()...)
				MainGame.AddCardToBottom(g.GuanXing.GetDownCards()...)
				extra.Index = MaxIndex
			} else if button.Show == TextCancel {
				cards := Map(g.GuanXing.Cards, func(item *CardUI) *Card {
					return item.Card
				})
				MainGame.AddCardToTop(cards...)
				extra.Index = MaxIndex
			}
			return true
		}
	}
	return false
}

func NewGuanXingStep(uis *EffectWithUI, guanXing *GuanXing) *GuanXingStep {
	return &GuanXingStep{UIs: uis, GuanXing: guanXing}
}

//====================LongDanCheckStep========================

type LongDanCheckStep struct {
	CardName string
}

func NewLongDanCheckStep(cardName string) *LongDanCheckStep {
	return &LongDanCheckStep{CardName: cardName}
}

func (w *LongDanCheckStep) Update(event *Event, extra *StepExtra) {
	event.Desc.RemoveCard(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	event.Resp = NewTransCardWrap(&Card{Name: w.CardName, Point: PointNone, Suit: SuitNone, Type: CardBasic}, extra.Cards)
	event.Abort = true
	MainGame.TriggerEvent(&Event{Type: EventRespCardAfter, Src: event.Desc, Desc: event.Src, Card: event.Card, Event: event})
	extra.Index = MaxIndex
}

//====================TieQiStep=======================

type TieQiStep struct {
}

func NewTieQiStep() *TieQiStep {
	return &TieQiStep{}
}

func (t *TieQiStep) Update(event *Event, extra *StepExtra) {
	judgeCard := extra.JudgeCard
	if IsRedSuit(judgeCard.Desc.Suit) {
		event.WrapFilter = t.NoFilter
	}
	extra.Index = MaxIndex
}

func (t *TieQiStep) NoFilter(card *CardWrap) bool {
	return false
}

//=====================ZhiHengStep=======================

type ZhiHengStep struct {
	ZhiHengSkill *ZhiHengSkill
}

func (z *ZhiHengStep) Update(event *Event, extra *StepExtra) {
	event.Src.RemoveCard(extra.Cards...)
	event.Src.RemoveEquip(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	event.Src.DrawCard(len(extra.Cards))
	z.ZhiHengSkill.Used = true
	extra.Index = MaxIndex
}

func NewZhiHengStep(zhiHengSkill *ZhiHengSkill) *ZhiHengStep {
	return &ZhiHengStep{ZhiHengSkill: zhiHengSkill}
}

//======================JiuYuanStep=============================

type JiuYuanStep struct {
}

func (j *JiuYuanStep) Update(event *Event, extra *StepExtra) {
	event.RecoverVal++
}

func NewJiuYuanStep() *JiuYuanStep {
	return &JiuYuanStep{}
}

//=======================KeJiStep=============================

type KeJiStep struct {
}

func (k *KeJiStep) Update(event *Event, extra *StepExtra) {
	event.StepExtra.Index = MaxIndex
	extra.Index = MaxIndex
}

func NewKeJiStep() *KeJiStep {
	return &KeJiStep{}
}

//=====================KuRouStep======================

type KuRouStep struct {
}

func (k *KuRouStep) Update(event *Event, extra *StepExtra) {
	if event.Src.ChangeHp(-1) {
		MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Desc: event.Src, HurtVal: 1})
	} else {
		MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Desc: event.Src, HurtVal: 1})
	}
	event.Src.DrawCard(2)
	extra.Index = MaxIndex
}

func NewKuRouStep() *KuRouStep {
	return &KuRouStep{}
}

//===================FanJianReqStep======================

type FanJianReqStep struct {
	SuitCards []*Card
}

func (f *FanJianReqStep) Update(event *Event, extra *StepExtra) {
	cards := Map(event.Src.Cards, func(item *CardUI) *Card {
		return item.Card
	}) // 花色&牌一块请求了
	extra.Result2 = &Event{Type: EventChooseCard, Src: event.Src, Desc: extra.Players[0],
		Cards: cards, ChooseMax: 1, ChooseMin: 1}
	MainGame.TriggerEvent(extra.Result2)
	extra.Result1 = &Event{Type: EventChooseCard, Src: event.Src, Desc: extra.Players[0],
		Cards: f.SuitCards, ChooseMax: 1, ChooseMin: 1}
	MainGame.TriggerEvent(extra.Result1)
	extra.Index++
}

func NewFanJianReqStep() *FanJianReqStep {
	return &FanJianReqStep{SuitCards: []*Card{{Suit: SuitDiamond}, {Suit: SuitHeart},
		{Suit: SuitSpade}, {Suit: SuitClub}}} // 直接使用牌来指定花色
}

//====================FanJianCheckStep=====================

type FanJianCheckStep struct {
}

func (f *FanJianCheckStep) Update(event *Event, extra *StepExtra) {
	// 完成卡的转移
	event.Src.RemoveCard(extra.Result2.Resps...)
	extra.Players[0].AddCard(extra.Result2.Resps...)
	// 砸一点伤害
	if extra.Result1.Resps[0].Suit != extra.Result2.Resps[0].Suit {
		if extra.Players[0].ChangeHp(-1) {
			MainGame.TriggerEvent(&Event{Type: EventPlayerDying, Desc: extra.Players[0], Src: event.Src, HurtVal: 1})
		} else {
			MainGame.TriggerEvent(&Event{Type: EventPlayerHurt, Desc: extra.Players[0], Src: event.Src, HurtVal: 1})
		}
	}
	extra.Index = MaxIndex
}

func NewFanJianCheckStep() *FanJianCheckStep {
	return &FanJianCheckStep{}
}

//====================LiuLiStep========================

type LiuLiStep struct {
}

func (l *LiuLiStep) Update(event *Event, extra *StepExtra) {
	if len(extra.Players) > 0 { // 有选择人
		event.Desc.RemoveCard(extra.Cards...)
		event.Desc.RemoveEquip(extra.Cards...)
		MainGame.AddToDesktop(extra.Cards...)
		MainGame.RemoveFromDesktop(extra.Cards...)
		// 转换目标
		event.StepExtra.Desc = extra.Players[0]
		event.Desc = extra.Players[0]
	}
	extra.Index = MaxIndex
}

func NewLiuLiStep() *LiuLiStep {
	return &LiuLiStep{}
}

//===================JieYinStep=====================

type JieYinStep struct {
	JieYinSkill *JieYinSkill
}

func (j *JieYinStep) Update(event *Event, extra *StepExtra) {
	// 移出牌
	event.Src.RemoveCard(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	// 恢复体力
	event.Src.ChangeHp(1)
	extra.Players[0].ChangeHp(1)
	j.JieYinSkill.Used = true
	extra.Index = MaxIndex
}

func NewJieYinStep(jieYinSkill *JieYinSkill) *JieYinStep {
	return &JieYinStep{JieYinSkill: jieYinSkill}
}

//=====================QingNangStep======================

type QingNangStep struct {
	QingNangSkill *QingNangSkill
}

func (q *QingNangStep) Update(event *Event, extra *StepExtra) {
	// 移出牌
	event.Src.RemoveCard(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	// 恢复体力
	extra.Players[0].ChangeHp(1)
	q.QingNangSkill.Used = true
	extra.Index = MaxIndex
}

func NewQingNangStep(qingNangSkill *QingNangSkill) *QingNangStep {
	return &QingNangStep{QingNangSkill: qingNangSkill}
}

//===================JiJiuStep========================

type JiJiuStep struct {
}

func (j *JiJiuStep) Update(event *Event, extra *StepExtra) {
	event.Desc.RemoveEquip(extra.Cards...) // 对面回移出手牌，但是不会移出装备牌 ，最好写到一起
	event.Resps = extra.Cards              // TODO 这里并没有转换为桃，对面不会检查 但这种场景没有接受包装牌
	event.Abort = true
	extra.Index = MaxIndex
}

func NewJiJiuStep() *JiJiuStep {
	return &JiJiuStep{}
}

//===================LiJianStep=======================

type LiJianStep struct {
}

func (l *LiJianStep) Update(event *Event, extra *StepExtra) {
	// 移出牌
	event.Src.RemoveCard(extra.Cards...)
	MainGame.AddToDesktop(extra.Cards...)
	MainGame.DiscardFromDesktop(extra.Cards...)
	// 虚拟决斗
	card := NewVirtualCardWrap(&Card{Name: "决斗", Point: PointNone, Suit: SuitNone, Type: CardKit, Skill: NewJueDouSkill(), KitType: KitInstant})
	event = &Event{Type: EventUseCard, Src: extra.Players[0], Descs: []*Player{extra.Players[1]}, Card: card, StepExtra: extra}
	effect := card.Desc.Skill.CreateEffect(event)
	MainGame.PushAction(NewEffectGroup(event, []IEffect{effect}))
	MainGame.AddToDesktopRaw(card)
	extra.Index = MaxIndex
}

func NewLiJianStep() *LiJianStep {
	return &LiJianStep{}
}

//=================DrawCardStep===================

type DrawCardStep struct {
	Player *Player // 通用简单摸个牌
	Num    int
}

func (d *DrawCardStep) Update(event *Event, extra *StepExtra) {
	d.Player.DrawCard(d.Num)
	extra.Index = MaxIndex
}

func NewDrawCardStep(player *Player, num int) *DrawCardStep {
	return &DrawCardStep{Player: player, Num: num}
}

//=================YaoWuStep=====================

type YaoWuStep struct {
}

func (y *YaoWuStep) Update(event *Event, extra *StepExtra) {
	src := event.Src
	if src.Hp < src.MaxHp {
		src.ChangeHp(1)
	} else {
		src.DrawCard(1)
	}
	extra.Index = MaxIndex
}

func NewYaoWuStep() *YaoWuStep {
	return &YaoWuStep{}
}

//====================WuShuangReqStep=====================

type WuShuangReqStep struct {
}

func NewWuShuangReqStep() *WuShuangReqStep {
	return &WuShuangReqStep{}
}

func (w *WuShuangReqStep) Update(event *Event, extra *StepExtra) {
	// 再要一次，这一次不要再设置卡牌了，防止循环
	extra.Result1 = &Event{Type: EventRespCard, Src: event.Desc, Desc: event.Src, WrapFilter: w.ShanFilter}
	MainGame.TriggerEvent(extra.Result1)
	extra.Index++
}

func (w *WuShuangReqStep) ShanFilter(card *CardWrap) bool {
	return card.Desc.Name == "闪"
}

//====================WuShuangCheckStep=====================

type WuShuangCheckStep struct {
}

func (w *WuShuangCheckStep) Update(event *Event, extra *StepExtra) {
	if extra.Result1.Resp == nil {
		event.Event.Resp = nil // 没有要到，移出第一次响应的牌
	} // 要到了无事发生
	extra.Index = MaxIndex
}

func NewWuShuangCheckStep() *WuShuangCheckStep {
	return &WuShuangCheckStep{}
}
