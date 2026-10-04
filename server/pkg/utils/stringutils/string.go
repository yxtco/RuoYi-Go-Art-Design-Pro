package stringutils

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// RemoveStart 如果 `str` 以 `remove` 开始，则从 `str` 中去除 `remove` 前缀。
func RemoveStart(str, remove string) string {
	// 检查 `str` 和 `remove` 都不为空且 `str` 以 `remove` 开始
	if str != "" && remove != "" && strings.HasPrefix(str, remove) {
		// 从 `str` 中去除 `remove` 前缀
		return strings.TrimPrefix(str, remove)
	}
	// 如果 `str` 不以 `remove` 开始或者任一为空，则返回原始的 `str`
	return str
}

// SubstringBetween 返回字符串 `str` 中 `open` 和 `close` 之间的子字符串。
// 如果 `str`、`open` 或 `close` 为空，或者找不到这样的子字符串，返回空字符串。
func SubstringBetween(str, open, close string) string {
	// 检查 `str`、`open` 和 `close` 是否为空
	if str == "" || open == "" || close == "" {
		return ""
	}

	// 查找 `open` 在 `str` 中的位置
	start := strings.Index(str, open)
	if start != -1 {
		// 查找 `close` 在 `str` 中的位置，从 `open` 之后的位置开始查找
		end := strings.Index(str[start+len(open):], close)
		if end != -1 {
			// 返回 `open` 和 `close` 之间的子字符串
			return str[start+len(open) : start+len(open)+end]
		}
	}

	// 如果没有找到，返回空字符串
	return ""
}

func Capitalize(str string) string {
	strLen := len(str)
	if strLen == 0 {
		return str
	}

	firstCodepoint, size := utf8.DecodeRuneInString(str)
	newCodePoint := unicode.ToTitle(firstCodepoint)

	if firstCodepoint == newCodePoint {
		return str
	}

	newCodePoints := make([]rune, 0, strLen)
	newCodePoints = append(newCodePoints, newCodePoint)

	for inOffset := size; inOffset < strLen; {
		codepoint, size := utf8.DecodeRuneInString(str[inOffset:])
		newCodePoints = append(newCodePoints, codepoint)
		inOffset += size
	}

	return string(newCodePoints)
}

func StartsWithAny(sequence string, searchStrings ...string) bool {
	if len(sequence) > 0 && len(searchStrings) > 0 {
		for _, searchString := range searchStrings {
			if strings.HasPrefix(sequence, searchString) {
				return true
			}
		}
	}
	return false
}

func ReplaceEach(text string, searchList []string, replacementList []string, repeat bool, timeToLive int) string {
	if timeToLive < 0 {
		searchSet := make(map[string]struct{})
		replacementSet := make(map[string]struct{})
		for _, searchItem := range searchList {
			searchSet[searchItem] = struct{}{}
		}
		for _, replacementItem := range replacementList {
			replacementSet[replacementItem] = struct{}{}
		}

		for k := range searchSet {
			if _, found := replacementSet[k]; found {
				panic("Aborting to protect against StackOverflowError - output of one loop is the input of another")
			}
		}
	}

	if text == "" || len(searchList) == 0 || len(replacementList) == 0 || (len(searchList) > 0 && timeToLive == -1) {
		return text
	}

	searchLength := len(searchList)
	replacementLength := len(replacementList)

	if searchLength != replacementLength {
		panic("Search and Replace array lengths don't match")
	}

	noMoreMatchesForReplIndex := make([]bool, searchLength)
	textIndex := -1
	replaceIndex := -1

	for start := 0; start < searchLength; start++ {
		if !noMoreMatchesForReplIndex[start] && searchList[start] != "" && replacementList[start] != "" {
			tempIndex := strings.Index(text, searchList[start])
			if tempIndex == -1 {
				noMoreMatchesForReplIndex[start] = true
			} else if textIndex == -1 || tempIndex < textIndex {
				textIndex = tempIndex
				replaceIndex = start
			}
		}
	}

	if textIndex == -1 {
		return text
	}

	start := 0
	increase := 0

	for i := 0; i < searchLength; i++ {
		if searchList[i] != "" && replacementList[i] != "" {
			i := len(replacementList[i]) - len(searchList[i])
			if i > 0 {
				increase += 3 * i
			}
		}
	}

	increase = len(text) / 5
	if increase > len(text)/5 {
		increase = len(text) / 5
	}

	var buf strings.Builder
	buf.Grow(len(text) + increase)

	for textIndex != -1 {
		for i := start; i < textIndex; i++ {
			buf.WriteByte(text[i])
		}

		buf.WriteString(replacementList[replaceIndex])
		start = textIndex + len(searchList[replaceIndex])
		textIndex = -1
		replaceIndex = -1

		for i := 0; i < searchLength; i++ {
			if !noMoreMatchesForReplIndex[i] && searchList[i] != "" && replacementList[i] != "" {
				tempIndex := strings.Index(text[start:], searchList[i])
				if tempIndex == -1 {
					noMoreMatchesForReplIndex[i] = true
				} else if textIndex == -1 || tempIndex < textIndex {
					textIndex = tempIndex + start
					replaceIndex = i
				}
			}
		}
	}

	for i := start; i < len(text); i++ {
		buf.WriteByte(text[i])
	}

	result := buf.String()
	if !repeat {
		return result
	} else {
		return ReplaceEach(result, searchList, replacementList, repeat, timeToLive-1)
	}
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StringSliceToUint64Slice(list []string) []uint64 {
	data := make([]uint64, 0)
	for _, s := range list {
		ui64, _ := strconv.ParseUint(s, 10, 64)
		data = append(data, ui64)
	}
	return data
}

func IsHttp(link string) bool {
	return StartsWithAny(link, "http://", "https://")
}

func ContainsIgnoreCase() {

}
