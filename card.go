/*
@author: sk
@date: 2024/5/1
*/
package main

type Card struct { // 暂时牌上无需携带标记，不需要额外字段存储
	Name  string
	Point CardPoint
	Suit  CardSuit
	Type  CardType
	Skill ICheckSkill // 处理目标是否合法与发动最终效果
	// 冗余参数
	EquipType EquipType
	KitType   KitType
}

type CardWrap struct { // 转换牌，打出的牌都是这个
	Desc *Card    // 作为什么牌使用的，使用这张牌对应的Skill发动
	Type WrapType // 怎么来的
	Src  []*Card  // 由什么牌转换来的，也可能只有一个不是转换来的，也可能一个牌也没有虚拟的
}

func NewSimpleCardWrap(card *Card) *CardWrap {
	return &CardWrap{Desc: card, Type: WrapSimple, Src: []*Card{card}}
}

//=======================CardManager==========================

type CardManager struct {
	Cards        []*Card
	DiscardCards []*Card
}

func (m *CardManager) DrawCard(num int) []*Card { // 先不考虑平局
	if num > len(m.Cards) {
		RandSlice(m.DiscardCards)
		m.Cards = append(m.Cards, m.DiscardCards...)
		m.DiscardCards = make([]*Card, 0)
	}

	res := m.Cards[:num]
	m.Cards = m.Cards[num:]
	return res
}

func (m *CardManager) DiscardCard(cards []*Card) {
	m.DiscardCards = append(m.DiscardCards, cards...)
}

func NewCardManager() *CardManager {
	return &CardManager{Cards: LoadCard(), DiscardCards: make([]*Card, 0)}
}
