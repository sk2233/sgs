/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	rand.Seed(time.Now().Unix())
	ebiten.SetWindowSize(WinWidth, WinHeight)
	InitGeneral()
	err := ebiten.RunGame(NewGame()) // 阻塞运行项目
	HandleErr(err)
}
