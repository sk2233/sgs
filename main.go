/*
@author: sk
@date: 2024/5/1
*/
package main

import "github.com/hajimehoshi/ebiten/v2"

func main() {
	ebiten.SetWindowSize(WinWidth, WinHeight)
	InitGeneral()
	err := ebiten.RunGame(NewGame()) // 阻塞运行项目
	HandleErr(err)
}
