/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
}

func NewGame() *Game {
	return &Game{}
}

// 每帧执行的逻辑，error我没用过
func (g *Game) Update() error {
	return nil
}

// 每帧绘制的画面
func (g *Game) Draw(screen *ebiten.Image) {
	FillRect(screen, 100, 100, 100, 100, ClrFF00FF55)
	StoryRect(screen, 200, 200, 100, 100, 2, ClrFFFFFF)
	DrawText(screen, "你好，世界", 300, 300, AnchorMidCenter, Font18, ClrFFFFFF)
}

// 设置画布的大小，入参窗口大小，返回画布大小
func (g *Game) Layout(w, h int) (int, int) {
	return w, h
}
