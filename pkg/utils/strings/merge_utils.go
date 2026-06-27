// Package strings @Author larry
// File merge_utils.go
// @Date 2024/4/23 16:15:00
// @Desc
package strings

import (
	"strings"
)

const (
	BlankSplitStr     = ""
	WaveSplitStr      = "~"
	DotSplitChar      = "."
	ColonSplitChar    = ":"
	PlusSplitChar     = "+"
	DollarSplitChar   = "$"
	WellSplitStr      = "#"
	UnderlineSplitStr = "_"
	HyphenSplitStr    = "-"
	SlashSplitStr     = "/"
	ArrowSplitStr     = "->"
	SplitSplitStr     = "|"
	SpaceSplitStr     = " "
	CommaSplitStr     = ","
)

func MergeStr(split string, strs ...string) string {
	return strings.Join(strs, split)
}

func MergeWithSplitter(left, splitStr, right string) string {
	return left + splitStr + right
}

func MergeBlank(left, right string) string {
	return MergeWithSplitter(left, BlankSplitStr, right)
}

func MergeWave(left, right string) string {
	return MergeWithSplitter(left, WaveSplitStr, right)
}

func MergeWell(left, right string) string {
	return MergeWithSplitter(left, WellSplitStr, right)
}

func MergeDot(left, right string) string {
	return MergeWithSplitter(left, DotSplitChar, right)
}

func MergeUnder(left, right string) string {
	return MergeWithSplitter(left, UnderlineSplitStr, right)
}

func MergeHyphen(left, right string) string {
	return MergeWithSplitter(left, HyphenSplitStr, right)
}

func MergePlus(left, right string) string {
	return MergeWithSplitter(left, PlusSplitChar, right)
}

func MergeColon(left, right string) string {
	return MergeWithSplitter(left, ColonSplitChar, right)
}

func MergeSlash(left, right string) string {
	return MergeWithSplitter(left, SlashSplitStr, right)
}

func MergeArrow(left, right string) string {
	return MergeWithSplitter(left, ArrowSplitStr, right)
}

func MergeSpace(left, right string) string {
	return MergeWithSplitter(left, SpaceSplitStr, right)
}

func MergeSplit(left, right string) string {
	return MergeWithSplitter(left, SplitSplitStr, right)
}

func LeftStr(mergeStr, splitStr string) string {
	split := strings.Split(mergeStr, splitStr)
	if len(split) > 0 {
		return split[0]
	}
	return ""
}

func RightStr(mergeStr, splitStr string) string {
	split := strings.Split(mergeStr, splitStr)
	if len(split) > 1 {
		return split[len(split)-1]
	}
	return ""
}
