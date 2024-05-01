/*
@author: sk
@date: 2024/5/1
*/
package main

var (
	ClrFFFFFF   = Hex2Clr("FFFFFF")
	ClrFF00FF55 = Hex2Clr("FF00FF55")
)

var (
	Font18 = NewFont(18)
)

type Anchor complex64 // 复数类型，非常适合当做向量，简单起见这里不会用到向量 这里把实数当x锚点，虚数当y锚点

const (
	AnchorMidCenter Anchor = 0.5 + 0.5i
)
