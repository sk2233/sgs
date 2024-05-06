/*
@author: sk
@date: 2024/5/4
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Desktop struct { // 实际就是卡牌的处理区，大部分卡牌都是先进入处理区，最终结算完毕再移除到弃牌堆
	Cards   []*CardWrap // 用来记录的
	CardUIs []*CardUI   // 用来绘制的
}

func NewDesktop() *Desktop {
	return &Desktop{Cards: make([]*CardWrap, 0)}
}

func (d *Desktop) Draw(screen *ebiten.Image) {
	for _, card := range d.CardUIs {
		card.Draw(screen)
	}
}

// x 240~1200-240 y 280+(280-160)/2 居中展示 卡牌：宽 110 高 160
func (d *Desktop) AddCard(cards []*CardWrap) {
	if len(cards) == 0 {
		return
	}
	if len(d.Cards) == 0 {
		d.Cards = cards
		d.CardUIs = make([]*CardUI, 0) // 主要是清空作用，可能上次的已经全部处理完了，为了防止处理的太快没有绘制，会保留UI,直到下次新增
	} else {
		d.Cards = append(d.Cards, cards...)
	}
	for _, card := range cards {
		d.CardUIs = append(d.CardUIs, NewCardUI(card.Desc))
	}
	d.TidyCard()
}

func (d *Desktop) DiscardCard(cards []*CardWrap) {
	if len(cards) == 0 {
		return
	}
	over := false
	for _, card := range cards {
		if card.Desc == d.Cards[0].Desc {
			over = true // 队列的第一张牌也处理完毕了，可以全部弃之了
			break
		}
	}
	if !over {
		return
	} // 全部处理掉，但是继续保持绘制，防止有时太快什么都没有展示
	for _, card := range d.Cards {
		MainGame.DiscardCard(card.Src)
	}
	d.Cards = make([]*CardWrap, 0)
}

func (d *Desktop) RemoveCard(cards []*CardWrap) {
	if len(cards) == 0 {
		return
	}
	temp := Map(cards, func(item *CardWrap) *Card {
		return item.Desc
	})
	set := NewSet(temp...)
	d.CardUIs = Filter(d.CardUIs, func(ui *CardUI) bool {
		return !set.Contain(ui.Card)
	})
	d.Cards = Filter(d.Cards, func(card *CardWrap) bool {
		return !set.Contain(card.Desc)
	})
	d.TidyCard()
}

func (d *Desktop) TidyCard() {
	w := float32(Min(WinWidth-240-240, 110*len(d.CardUIs)))
	offset := (w - 110) / float32(len(d.CardUIs)-1)
	x := (WinWidth - w) / 2
	y := float32(280 + (280-160)/2)
	for _, card := range d.CardUIs {
		card.X, card.Y = x, y
		x += offset
	}
}
