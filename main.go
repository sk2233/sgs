/*
@author: sk
@date: 2024/5/1
*/
package sgs_study

import "github.com/hajimehoshi/ebiten/v2"

func main() {
	err := ebiten.RunGame(NewGame())
	if err != nil {
		return
	}
}
