/*
@author: sk
@date: 2024/5/1
*/
package main

type StepExtra struct {
	Index     int // 步骤进行到那里了
	JudgeCard *CardWrap
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
	MainGame.PushAction(NewEffectGroupBySkill(&Event{
		Type:      EventCardSkill,
		Src:       src,   // 这里相当于延时技能自己成来源了
		StepExtra: extra, // 主要为了传递判定牌
	}, card.Desc.Skill))
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
	extra.JudgeCard = NewSimpleCardWrap(MainGame.DrawCard(1)[0]) // 进行判定可能经历修改
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

type PlayStageMainStep struct {
	PlayStage *PlayStage // 需要获取到按钮状态
}

func NewPlayStageMainStep(playStage *PlayStage) *PlayStageMainStep {
	return &PlayStageMainStep{PlayStage: playStage}
}

// 当前设计比较简单，可以任意选择任意张卡，任意名角色，然后点击「出牌」「取消」或技能，「出牌」调用卡牌技能处理
// 「取消」若是由存在选择的手牌，先取消选择，再点击取消流转到下个回合，点击技能触发技能处理
func (p *PlayStageMainStep) Update(event *Event, extra *StepExtra) {
	x, y, ok := MouseClick()
	if !ok { // 点击事件是基础
		return
	}
	player := event.Src
	// 判断点击按钮
	if p.HandleBtnClick(player, x, y, extra) {
		return
	}
	// 判断点击手牌
	if player.ToggleCard(x, y) {
		return
	}
	// TODO 判断点击装备
	// 判断点击player
	MainGame.TogglePlayer(x, y)
}

func (p *PlayStageMainStep) HandleBtnClick(player *Player, x, y float32, extra *StepExtra) bool {
	for _, button := range p.PlayStage.Buttons {
		if button.Click(x, y) {
			if button.Text == TextPlayCard {
				// TODO 触发牌的效果
				cards := player.GetSelectCard() // 后面这里是只能有一张的 这里先不管
				player.DiscardCard(cards)
				MainGame.ResetPlayer()
			} else if button.Text == TextCancel {
				cards := player.GetSelectCard()
				if len(cards) > 0 { // 取消选择
					player.TidyCard()
				} else { // 下一步
					extra.Index = MaxIndex
				}
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
	return &BotPlayStageStep{Timer: 60}
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
	player.ToggleCard(x, y)
}

func (d *DiscardStageMainStep) HandleBtnClick(player *Player, x, y float32, extra *StepExtra) bool {
	for _, button := range d.DiscardStage.Buttons {
		if button.Click(x, y) {
			cards := player.GetSelectCard()
			if len(cards) == 0 { // 选择 0 张牌选什么都没有意义
				return true
			}
			if button.Text == TextConfirm {
				condition := MainGame.ComputeCondition(&Condition{Type: ConditionMaxCard, Src: player})
				if len(cards) <= len(player.Cards)-condition.MaxCard {
					player.DiscardCard(cards)
					extra.Index = 0 // 再去 Check
				}
			} else if button.Text == TextCancel {
				player.TidyCard()
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
	return &BotDiscardStageStep{Timer: 60} // 1s后弃牌，一步到位
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
			player.DiscardCard(cards)
		}
		extra.Index = MaxIndex
	}
}
