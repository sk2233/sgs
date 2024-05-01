/*
@author: sk
@date: 2024/5/1
*/
package main

var (
	Clr65401E = Hex2Clr("65401E")
	Clr553010 = Hex2Clr("553010")
	ClrFFFFFF = Hex2Clr("FFFFFF")
	Clr000000 = Hex2Clr("000000")
	Clr0D0D0D = Hex2Clr("0D0D0D")
	Clr362618 = Hex2Clr("362618")
	ClrD84B3C = Hex2Clr("D84B3C") // hp 红
	ClrFEF660 = Hex2Clr("FEF660") // hp 黄
	ClrAAD745 = Hex2Clr("AAD745") // hp 绿
)

var (
	Font18 = NewFont(18)
)

type Anchor complex64 // 复数类型，非常适合当做向量，简单起见这里不会用到向量 这里把实数当x锚点，虚数当y锚点

const (
	AnchorMidCenter Anchor = 0.5 + 0.5i
	AnchorTopLeft   Anchor = 0 + 0i
	AnchorTopCenter Anchor = 0.5 + 0i
)

const (
	WinWidth  = 1200
	WinHeight = 720
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
	ForceWei = "魏"
	ForceShu = "蜀"
	ForceWu  = "吴"
	ForceQun = "群"
)
