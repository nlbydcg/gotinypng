package cPrint

import "fmt"

type ColorType uint // 字体颜色

const (
	ColorTypeBlack ColorType = iota + 30
	ColorTypeRed
	ColorTypeGreen
	ColorTypeYellow
	ColorTypeBlue
	ColorTypeMagenta // 紫红色
	ColorTypeCyan    // 青蓝色
	ColorTypeWhite
)

type ShowType uint // 展示类型

const (
	ShowTypeDefault   ShowType   = iota
	ShowTypeHigh                 // 高亮
	ShowTypeUnderline = iota + 2 // 下划线
	ShowTypeFlash                // 闪烁
	ShowTypeAntiWhite = iota + 3 // 反白
	ShowTypeInvisible            // 不可见
)

type BgColorType uint // 背景颜色

const (
	BgColorTypeBlack BgColorType = iota + 40
	BgColorTypeRed
	BgColorTypeGreen
	BgColorTypeYellow
	BgColorTypeBlue
	BgColorTypeMagenta // 紫红色
	BgColorTypeCyan    // 青蓝色
	BgColorTypeWhite
)

func ColorPrint(s string, bgColor BgColorType, color ColorType, st ShowType) {
	fmt.Printf("%c[%d;%d;%dm%s%c[0m", 0x1B, st, bgColor, color, s, 0x1B)
}
func ColorPrintln(s string, bgColor BgColorType, color ColorType, st ShowType) {
	fmt.Printf("%c[%d;%d;%dm%s%c[0m\n\n", 0x1B, st, bgColor, color, s, 0x1B)
}

// 显示错误
func Error(s string) {
	// 高亮红色 黑色底色
	ColorPrintln(s, BgColorTypeBlack, ColorTypeRed, ShowTypeHigh)
}

// 显示成功
func Success(s string) {
	ColorPrintln(s, BgColorTypeBlack, ColorTypeGreen, ShowTypeHigh)
}

type PrintStruct struct {
	Message     string
	ColorType   ColorType
	ShowType    ShowType
	BgColorType BgColorType
}

// 批量提示信息
func PrintList(list []*PrintStruct) {
	for _, v := range list {
		ColorPrint(v.Message, v.BgColorType, v.ColorType, v.ShowType)
	}
	fmt.Println()
}
