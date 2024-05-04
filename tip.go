/*
@author: sk
@date: 2024/5/3
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Tip struct {
	Info   string
	Offset float32
}

type TipManager struct {
	Tips []*Tip
}

func NewTipManager() *TipManager {
	return &TipManager{Tips: make([]*Tip, 0)}
}

func (m *TipManager) Update() {
	if len(m.Tips) == 0 {
		return
	}
	for _, tip := range m.Tips {
		tip.Offset++
	}
	if m.Tips[0].Offset > MaxTipOffset { // 只有最后一个可能出界
		m.Tips = m.Tips[1:]
	}
}

func (m *TipManager) Draw(screen *ebiten.Image) {
	for _, tip := range m.Tips {
		DrawText(screen, tip.Info, WinWidth/2, WinHeight/2-tip.Offset, AnchorMidCenter, Font18, ClrFF0000)
	}
}

func (m *TipManager) AddTip(info string) {
	m.Tips = append(m.Tips, &Tip{Info: info})
	for i := len(m.Tips) - 2; i >= 0; i-- {
		if m.Tips[i].Offset-m.Tips[i+1].Offset < MinTipInterval {
			m.Tips[i].Offset = m.Tips[i+1].Offset + MinTipInterval
		}
	}
	if m.Tips[0].Offset > MaxTipOffset { // 只有最后一个可能出界
		m.Tips = m.Tips[1:]
	}
}
