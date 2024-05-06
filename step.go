/*
@author: sk
@date: 2024/5/1
*/
package main

type StepExtra struct {
	Index     int // 步骤进行到那里了
	JudgeCard *CardWrap
	ShaCount  int // 出了几次杀了
	Card      *CardUI
	Cards     []*Card
	MaxDesc   int
	Result    *Event // 事件即是参数，也存储结果
	Desc      *Player
	Select    string
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
	MainGame.TriggerEvent(&Event{Type: t.EventType, Src: event.Src}) // TODO 参数后续可能需要继续补充
	extra.Index++
}

//=====================DrawStageMainStep摸牌阶段的主要步骤========================

type DrawStageMainStep struct {
}

func NewDrawStageMainStep() *DrawStageMainStep {
	return &DrawStageMainStep{}
}

func (d *DrawStageMainStep) Update(event *Event, extra *StepExtra) {
	condition := MainGame.ComputeCondition(&Condition{Type: ConditionDrawCardNum, Src: event.Src})
	event.Src.DrawCard(condition.CardNum)
	// TODO TEST
	//event.Src.AddCard(&Card{Name: "寒冰剑", Point: CardPoint(rand.Intn(13) + 1), EquipAlias: "寒冰剑2",
	//	Suit: CardSuit(rand.Intn(4) + 1), Type: CardEquip, EquipType: EquipWeapon, Skill: NewEquipSkill()})
	extra.Index = MaxIndex
}

//======================JudgeStageCheckStep判定阶段检查步骤=======================

type JudgeStageCheckStep struct {
}

func NewJudgeStageCheckStep() *JudgeStageCheckStep {
	return &JudgeStageCheckStep{}
}

func (j *JudgeStageCheckStep) Update(event *Event, extra *StepExtra) {
	if len(event.Src.JudgeCards) > 0 { // 这时判定牌应该放到处理区了 TODO
		extra.Index++ // 还有判定牌接着判定
	} else {
		extra.Index = MaxIndex // 没有了结束
	}
}

//============================JudgeStageExecuteStep判定阶段判定牌生效完清理步骤=================================

type JudgeStageExecuteStep struct {
}

func NewJudgeStageExecuteStep() *JudgeStageExecuteStep {
	return &JudgeStageExecuteStep{}
}

func (j *JudgeStageExecuteStep) Update(event *Event, extra *StepExtra) { // 普通判定也有这也这一步骤，但是判定阶段需要构成循环
	extra.Index = 0
	src := event.Src
	card := src.JudgeCards[0]
	src.JudgeCards = src.JudgeCards[1:]
	MainGame.PushAction(NewEffectGroupBySkill(&Event{ // 延时锦囊牌的牌面效果与装备牌类似，都是添加一个新对象 TODO 这里是有问题的
		Type:      EventUseCard,
		Src:       src,   // 这里相当于延时技能自己成来源了
		StepExtra: extra, // 主要为了传递判定牌
	}, card.Desc.Skill)) // TODO 判定这里有问题不能直接丢弃
	MainGame.DiscardCard(append(card.Src, extra.JudgeCard.Src...)) // 丢弃延时锦囊牌与判定牌，若是他们还有实体牌的话
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
	for _, card := range player.Cards {
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
	// TODO 判断点击技能
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
		MainGame.TriggerEvent(&Event{Type: EventCardPoint, Src: event.Src, Card: event.Card, Desc: extra.Desc})
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
	extra.Result = &Event{Type: EventRespCard, Src: event.Src, Card: event.Card, Desc: extra.Desc,
		WrapFilter: r.ShanFilter, HurtVal: 1}
	MainGame.TriggerEvent(extra.Result)
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
	temp := extra.Result
	extra.Result = &Event{Type: EventShaHit, Src: temp.Src, Desc: temp.Desc, Card: temp.Card, HurtVal: temp.HurtVal, ShaHit: false}
	if !temp.Invalid && (temp.Resp == nil || temp.Force) { // 没有闪或强制崩血 且改事件有效
		extra.Result.ShaHit = true
		MainGame.TriggerEvent(extra.Result)
	}
	extra.Index++
}

//=======================RespShaExecuteStep=============================

type RespShaExecuteStep struct {
}

func (r *RespShaExecuteStep) Update(event *Event, extra *StepExtra) {
	// 结算后事件
	MainGame.TriggerEvent(&Event{Type: EventCardAfter, Src: event.Src, Card: event.Card, Desc: extra.Desc})
	temp := extra.Result
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
	Num       int        // 选择固定数量的牌，没有其他约束了
	WithEquip bool       // 是否包含装备牌
	Player    *Player    // 为了更通用需要确认当前执行人
	UIs       *EffectWithUI
	Init0     bool
	Buttons   []*Button
}

func NewSelectNumCardStep(filter CardFilter, num int, withEquip bool, player *Player, UIs *EffectWithUI) *SelectNumCardStep {
	return &SelectNumCardStep{Filter: filter, Num: num, WithEquip: withEquip, Player: player, UIs: UIs, Buttons: make([]*Button, 0)}
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
		if len(s.Player.GetSelectCard()) < s.Num {
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
				if len(cards) == s.Num {
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

//====================CiXiongShuangGuJianReqStep判断用户选择===================

type CiXiongShuangGuJianAskStep struct {
}

func (c *CiXiongShuangGuJianAskStep) Update(event *Event, extra *StepExtra) {
	if extra.Select == TextConfirm {
		extra.Result = &Event{Type: EventAskCard, Src: event.Src, Desc: event.Desc, AskNum: 1, Filter: c.AnyFilter}
		MainGame.TriggerEvent(extra.Result)
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
	temp := extra.Result
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
