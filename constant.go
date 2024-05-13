/*
@author: sk
@date: 2024/5/1
*/
package main

import "fmt"

var (
	Clr65401E   = Hex2Clr("65401E")
	Clr553010   = Hex2Clr("553010")
	ClrFFFFFF   = Hex2Clr("FFFFFF")
	ClrFF0000   = Hex2Clr("FF0000")
	Clr000000   = Hex2Clr("000000")
	Clr00FF00   = Hex2Clr("00FF00")
	Clr0D0D0D   = Hex2Clr("0D0D0D")
	Clr362618   = Hex2Clr("362618")
	ClrDECDBA   = Hex2Clr("DECDBA")
	Clr00000080 = Hex2Clr("00000080")
	ClrD84B3C   = Hex2Clr("D84B3C") // hp 红
	ClrFEF660   = Hex2Clr("FEF660") // hp 黄
	ClrAAD745   = Hex2Clr("AAD745") // hp 绿
	Clr348EBB   = Hex2Clr("348EBB") // 按钮背景色
	ClrA66F3F   = Hex2Clr("A66F3F") // 按钮描边色
	Clr4B403F   = Hex2Clr("4B403F")
)

var (
	Font18 = NewFont(18)
	Font16 = NewFont(16)
)

type Anchor complex64 // 复数类型，非常适合当做向量，简单起见这里不会用到向量 这里把实数当x锚点，虚数当y锚点

const (
	AnchorMidCenter Anchor = 0.5 + 0.5i
	AnchorTopLeft   Anchor = 0 + 0i
	AnchorTopCenter Anchor = 0.5 + 0i
	AnchorBtmCenter Anchor = 0.5 + 1i
	AnchorTopRight  Anchor = 1 + 0i
	AnchorMidLeft   Anchor = 0 + 0.5i
	AnchorMidRight  Anchor = 1 + 0.5i
)

const (
	WinWidth  = 1200
	WinHeight = 720
)

const (
	MaxIndex = 9999
)

type Role string // 身份

const (
	RoleUnknown   Role = "未知"
	RoleZhuGong   Role = "主公"
	RoleZhongChen Role = "忠臣"
	RoleFanZei    Role = "反贼"
	RoleNeiJian   Role = "内奸"
)

type Force string

const (
	ForceWei Force = "魏"
	ForceShu Force = "蜀"
	ForceWu  Force = "吴"
	ForceQun Force = "群"
)

type EventType string

const (
	EventPlayerStage   EventType = "EventPlayerStage"
	EventGameStart     EventType = "EventGameStart"     // 游戏开始事件，有些武将技能在这里发动
	EventStagePrepare  EventType = "EventStagePrepare"  // 准备阶段事件
	EventStageEnd      EventType = "EventStageEnd"      // 回合结束阶段事件
	EventJudgeCard     EventType = "EventJudgeCard"     // 判定事件发生后已经拿到判定牌了，但是还没有生效
	EventJudgeEnd      EventType = "EventJudgeEnd"      // 判定牌生效后
	EventUseCard       EventType = "EventUseCard"       // 调用卡牌效果
	EventCardPoint     EventType = "EventCardPoint"     // 使用卡牌指定时
	EventRespCard      EventType = "EventRespCard"      // 要求响应某些牌，是专门响应牌的，可以使用各种虚拟牌等且只要一张
	EventAskCard       EventType = "EventAskCard"       // 直接要牌的，不能使用转换牌(实际可以要例如借刀杀人)，可以要多张
	EventShaHit        EventType = "EventShaHit"        // 杀命中事件
	EventRespCardAfter EventType = "EventRespCardAfter" // 对方响应什么牌后
	EventCardAfter     EventType = "EventCardAfter"     // 卡牌结算后
	EventPlayerHurt    EventType = "EventPlayerHurt"    // 玩家受到攻击
	EventPlayerDying   EventType = "EventPlayerDying"   // 玩家濒死
	EventChooseCard    EventType = "EventChooseCard"    // 用户选牌，都是基础类型的牌
)

type ConditionType string

const (
	ConditionInitCard    ConditionType = "ConditionInitCard"    // 初始手牌数量
	ConditionDrawCardNum ConditionType = "ConditionDrawCardNum" // 摸牌阶段摸牌数量
	ConditionMaxCard     ConditionType = "ConditionMaxCard"     // 计算手牌上限
	ConditionGetDist     ConditionType = "ConditionGetDist"     // 计算从src到desc的距离
	ConditionUseCard     ConditionType = "ConditionUseCard"     // 计算使用牌的一些总的条件，不针对具体目标
	ConditionCardMaxDesc ConditionType = "ConditionCardMaxDesc" // 卡牌最多可指定的目标数目
	ConditionUseSha      ConditionType = "ConditionUseSha"      // 计算使用杀的一些条件，针对具体目标
)

type SkillTag int

const (
	TagLock   SkillTag = 1 << iota // 锁定技
	TagActive                      // 可以出牌阶段主动发动的
	TagNone   SkillTag = 0
)

type StageType int

const (
	StagePrepare StageType = 1 << iota
	StageJudge
	StageDraw
	StagePlay
	StageDiscard
	StageEnd
	StageNone StageType = 0
)

type CardPoint int

func (c CardPoint) String() string {
	switch c {
	case PointNone:
		return "无"
	case PointA:
		return "A"
	case Point2:
		return "2"
	case Point3:
		return "3"
	case Point4:
		return "4"
	case Point5:
		return "5"
	case Point6:
		return "6"
	case Point7:
		return "7"
	case Point8:
		return "8"
	case Point9:
		return "9"
	case Point10:
		return "10"
	case PointJ:
		return "J"
	case PointQ:
		return "Q"
	case PointK:
		return "K"
	default:
		panic(fmt.Sprintf("invalid point %d", c))
	}
}

const (
	PointNone CardPoint = iota // 没有点数
	PointA
	Point2
	Point3
	Point4
	Point5
	Point6
	Point7
	Point8
	Point9
	Point10
	PointJ
	PointQ
	PointK
)

type CardSuit int

func (c CardSuit) String() string {
	switch c {
	case SuitNone:
		return "无"
	case SuitHeart:
		return "红"
	case SuitSpade:
		return "黑"
	case SuitClub:
		return "梅"
	case SuitDiamond:
		return "方"
	default:
		panic(fmt.Sprintf("invalid suit %d", c))
	}
}

const (
	SuitNone CardSuit = iota // 没有花色
	SuitHeart
	SuitSpade
	SuitClub
	SuitDiamond
)

type CardType string

const (
	CardBasic CardType = "基本"
	CardEquip CardType = "装备"
	CardKit   CardType = "锦囊"
)

type EquipType string

const (
	EquipWeapon  EquipType = "武器"
	EquipArmor   EquipType = "防具"
	EquipAttack  EquipType = "-1马"
	EquipDefense EquipType = "+1马"
)

type KitType string

const (
	KitInstant KitType = "即时"
	KitDelay   KitType = "延时"
)

type WrapType string

const (
	WrapSimple  WrapType = ""
	WrapTrans   WrapType = "转换"
	WrapVirtual WrapType = "虚拟"
)

const (
	TextPlayCard = "出牌"
	TextCancel   = "取消"
	TextConfirm  = "确定"
)

const (
	MaxTipOffset   = WinHeight / 4
	MinTipInterval = 25
)

const (
	BotTimer = 30 // 单位是帧
)

var (
	EquipIndexes = map[EquipType]float32{
		EquipWeapon:  0,
		EquipArmor:   1,
		EquipAttack:  2,
		EquipDefense: 3,
	}
)

type Gender string

const (
	GenderMan   Gender = "男"
	GenderWoman Gender = "女"
)
