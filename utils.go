/*
@author: sk
@date: 2024/5/1
*/
package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func RandSlice(arr any) {
	sort.Slice(arr, func(_, _ int) bool {
		return rand.Intn(2) == 0
	})
}

// 获取对于 bot 最珍贵的牌，一般这时候该进入回合外了
func GetBotCardVal(card *Card) int {
	switch card.Name {
	// 保命牌
	case "桃":
		return 10
	case "闪", "无懈可击":
		return 9
	case "杀":
		return 8
		// 高性价比牌
	case "乐不思蜀":
		return 7
		// 无损的锦囊
	case "无中生有", "南蛮入侵", "万箭齐发":
		return 5
		// 有风险的锦囊
	case "决斗", "五谷丰登":
		return 4
		// bot不主动使用的锦囊
	case "过河拆桥", "顺手牵羊", "借刀杀人", "闪电":
		return 0
	default:
		if card.Type == CardEquip && card.EquipType == EquipDefense { // 防御马
			return 6
		} else {
			// 默认牌价值
			return 3
		}
	}
}

func GetEnemy(role Role) []*Player {
	var players []*Player
	var orders map[Role]int
	switch role {
	case RoleZhuGong, RoleZhongChen: // 主公，忠臣目标一致
		players = MainGame.GetPlayers(func(player *Player) bool {
			return !player.IsDie && (player.Role == RoleFanZei || player.Role == RoleNeiJian)
		})
		orders = map[Role]int{
			RoleFanZei:  1, // 先打反
			RoleNeiJian: 2,
		}
	case RoleFanZei:
		players = MainGame.GetPlayers(func(player *Player) bool {
			return !player.IsDie && (player.Role != RoleFanZei)
		})
		orders = map[Role]int{
			RoleZhuGong:   1, // 先打主公，实在打不到打忠臣
			RoleZhongChen: 2,
			RoleNeiJian:   3,
		}
	case RoleNeiJian:
		players = MainGame.GetPlayers(func(player *Player) bool {
			return !player.IsDie && (player.Role != RoleNeiJian)
		})
		orders = map[Role]int{
			RoleFanZei:    1, // 先打反贼，实在打不到打忠臣
			RoleZhongChen: 2,
			RoleZhuGong:   3,
		}
	default:
		panic(fmt.Sprintf("invalid role %v", role))
	}
	sort.Slice(players, func(i, j int) bool {
		return orders[players[i].Role] < orders[players[j].Role]
	})
	return players
}

//=====================绘图========================

// 只支持 6 位或 8 位的颜色转换
func Hex2Clr(hex string) color.Color {
	if len(hex) != 6 && len(hex) != 8 {
		panic(fmt.Sprintf("err hex %s", hex))
	}

	data := []uint8{255, 255, 255, 255}
	for i := 0; i < len(hex); i += 2 {
		temp, err := strconv.ParseInt(hex[i:i+2], 16, 64)
		HandleErr(err)
		data[i/2] = uint8(temp)
	}
	return color.RGBA{R: data[0], G: data[1], B: data[2], A: data[3]}
}

func FillRect(screen *ebiten.Image, x, y, w, h float32, clr color.Color) {
	vector.DrawFilledRect(screen, x, y, w, h, clr, false)
}

func StrokeRect(screen *ebiten.Image, x, y, w, h, sw float32, clr color.Color) {
	vector.StrokeRect(screen, x, y, w, h, sw, clr, false)
}

var defaultFont *opentype.Font

func NewFont(size float64) font.Face {
	if defaultFont == nil {
		// 必须使用支持中文的字体
		// https://github.com/TakWolf/fusion-pixel-font?tab=readme-ov-file
		bs, err := os.ReadFile("res/fusion-pixel-12px-monospaced-zh_hans.ttf")
		HandleErr(err)
		defaultFont, err = opentype.Parse(bs)
		HandleErr(err)
	}

	// 设置字体样式  dpi不清楚什么意思，这里固定为96
	face, err := opentype.NewFace(defaultFont, &opentype.FaceOptions{Size: size, DPI: 96, Hinting: font.HintingFull})
	HandleErr(err)
	return face
}

func DrawText(screen *ebiten.Image, str string, x, y float32, anchor Anchor, face font.Face, clr color.Color) {
	bound := text.BoundString(face, str)
	// 根据测量的结果，先把整体左上角移动到 x,y 处，再根据锚点与测量的宽高进行偏移
	// Anchor 是一个复数类型 实数当x锚点，虚数当y锚点，取值都在 0.0～1.0
	x = x - float32(bound.Min.X) - float32(bound.Dx())*real(anchor)
	y = y - float32(bound.Min.Y) - float32(bound.Dy())*imag(anchor)
	text.Draw(screen, str, face, int(x), int(y), clr)
}

func FillCircle(screen *ebiten.Image, x, y, r float32, clr color.Color) {
	vector.DrawFilledCircle(screen, x, y, r, clr, false)
}

func StrokeCircle(screen *ebiten.Image, x, y, r, sw float32, clr color.Color) {
	vector.StrokeCircle(screen, x, y, r, sw, clr, false)
}

func GetSuitClr(suit CardSuit) color.Color {
	if suit == SuitHeart || suit == SuitDiamond {
		return ClrFF0000
	}
	return Clr000000
}

func IsRedSuit(suit CardSuit) bool {
	return suit == SuitHeart || suit == SuitDiamond
}

func IsBlackSuit(suit CardSuit) bool {
	return suit == SuitClub || suit == SuitSpade
}

func GetHpClr(hp, maxHp int) color.Color {
	if hp > (maxHp+1)/2 { // 大于一半 绿色
		return ClrAAD745
	} else if hp > (maxHp+1)/4 { // 大于1/4 黄色
		return ClrFEF660
	} else { // 否则红色
		return ClrD84B3C
	}
}

//===================文本=======================

func VerticalText(val string) string {
	buff := strings.Builder{}
	items := []rune(val)
	for i := 0; i < len(items); i++ {
		if i > 0 {
			buff.WriteByte('\n')
		}
		buff.WriteRune(items[i])
	}
	return buff.String()
}

func Int2Str(val int) string {
	return strconv.FormatInt(int64(val), 10)
}

//===================Invoke=====================

type IDraw interface {
	Draw(screen *ebiten.Image)
}

func InvokeDraw(src any, screen *ebiten.Image) {
	if tar, ok := src.(IDraw); ok {
		tar.Draw(screen)
	}
}

type IStageDraw interface {
	DrawStage(screen *ebiten.Image, player *Player, extra *StageExtra)
}

func InvokeDrawStage(src any, screen *ebiten.Image, player *Player, extra *StageExtra) {
	if tar, ok := src.(IStageDraw); ok {
		tar.DrawStage(screen, player, extra)
	}
}

type IEffectDraw interface {
	DrawEffect(screen *ebiten.Image, event *Event)
}

func InvokeDrawEffect(src any, screen *ebiten.Image, event *Event) {
	if tar, ok := src.(IEffectDraw); ok {
		tar.DrawEffect(screen, event)
	}
}

type IStageInit interface {
	InitStage(player *Player, extra *StageExtra)
}

func InvokeInitStage(src any, player *Player, extra *StageExtra) {
	if tar, ok := src.(IStageInit); ok {
		tar.InitStage(player, extra)
	}
}

type ITop interface {
	Top()
}

func InvokeTop(src any) {
	if tar, ok := src.(ITop); ok {
		tar.Top()
	}
}

type IStageTop interface {
	TopStage(player *Player, extra *StageExtra)
}

func InvokeTopStage(src any, player *Player, extra *StageExtra) {
	if tar, ok := src.(IStageTop); ok {
		tar.TopStage(player, extra)
	}
}

//==================点击交互===================

// 与游戏的交互只有右键点击
func MouseClick() (float32, float32, bool) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return 0, 0, false
	}
	x, y := ebiten.CursorPosition()
	return float32(x), float32(y), true
}

//==================collection==================

func Filter[T any](data []T, filter func(T) bool) []T {
	res := make([]T, 0)
	for _, item := range data {
		if filter(item) {
			res = append(res, item)
		}
	}
	return res
}

func Map[S any, D any](data []S, trans func(S) D) []D {
	res := make([]D, 0, len(data))
	for _, item := range data {
		res = append(res, trans(item))
	}
	return res
}

func SubSlice[T comparable](all, sub []T) []T {
	set := NewSet[T](sub...)
	res := make([]T, 0)
	for _, item := range all {
		if !set.Contain(item) {
			res = append(res, item)
		}
	}
	return res
}

func HasAny[T any](all []T, filter func(T) bool) bool {
	for _, item := range all {
		if filter(item) {
			return true
		}
	}
	return false
}

func ReverseSlice[T any](data []T) {
	l, r := 0, len(data)-1
	for l < r {
		data[l], data[r] = data[r], data[l]
		l, r = l+1, r-1
	}
}

//===================math==================

func Abs[T int](val T) T {
	if val < 0 {
		return -val
	}
	return val
}

func Min[T int](val1, val2 T) T {
	if val1 < val2 {
		return val1
	}
	return val2
}
