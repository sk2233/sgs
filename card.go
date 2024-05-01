/*
@author: sk
@date: 2024/5/1
*/
package main

type Card struct {
}

//=======================CardManager==========================

type CardManager struct {
	Cards        []*Card
	DiscardCards []*Card
}

func (m *CardManager) DrawCard(num int) []*Card {
	res := m.Cards[:num]
	m.Cards = m.Cards[num:]
	return res
}

func NewCardManager() *CardManager {
	return &CardManager{Cards: LoadCard(), DiscardCards: make([]*Card, 0)}
}
