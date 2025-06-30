package tokenizer

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/blevesearch/bleve/v2/analysis"
)

// SafeChineseTokenizer 纯 Go 实现的安全中文分词器
// 不依赖 CGO，避免权限相关的段错误
type SafeChineseTokenizer struct {
	// 中文字符正则表达式
	chineseRegex *regexp.Regexp
	// 英文单词正则表达式
	englishRegex *regexp.Regexp
}

func NewSafeChineseTokenizer() *SafeChineseTokenizer {
	return &SafeChineseTokenizer{
		// 匹配中文字符（包括标点符号）
		chineseRegex: regexp.MustCompile(`[\p{Han}]+`),
		// 匹配英文单词和数字
		englishRegex: regexp.MustCompile(`[a-zA-Z0-9]+`),
	}
}

func (t *SafeChineseTokenizer) Tokenize(sentence []byte) analysis.TokenStream {
	text := string(sentence)
	result := make(analysis.TokenStream, 0, 20)
	position := 1

	// 1. 首先提取英文单词和数字
	englishMatches := t.englishRegex.FindAllStringIndex(text, -1)
	for _, match := range englishMatches {
		word := strings.TrimSpace(text[match[0]:match[1]])
		if len(word) >= 2 { // 最小长度过滤
			token := &analysis.Token{
				Term:     []byte(strings.ToLower(word)),
				Start:    match[0],
				End:      match[1],
				Position: position,
				Type:     analysis.AlphaNumeric,
			}
			result = append(result, token)
			position++
		}
	}

	// 2. 然后处理中文字符
	chineseMatches := t.chineseRegex.FindAllStringIndex(text, -1)
	for _, match := range chineseMatches {
		chineseText := text[match[0]:match[1]]

		// 对中文文本进行简单的字符级别分词
		tokens := t.segmentChinese(chineseText, match[0], &position)
		result = append(result, tokens...)
	}

	return result
}

// segmentChinese 对中文文本进行分词
func (t *SafeChineseTokenizer) segmentChinese(text string, startOffset int, position *int) []*analysis.Token {
	var result []*analysis.Token

	runes := []rune(text)

	for i := 0; i < len(runes); {
		// 尝试找到合适的词汇边界
		wordEnd := t.findWordBoundary(runes, i)

		if wordEnd > i {
			word := string(runes[i:wordEnd])
			if len(word) >= 2 { // 最小长度过滤
				// 计算字节偏移量
				byteStart := startOffset + len(string(runes[:i]))
				byteEnd := startOffset + len(string(runes[:wordEnd]))

				token := &analysis.Token{
					Term:     []byte(word),
					Start:    byteStart,
					End:      byteEnd,
					Position: *position,
					Type:     analysis.Ideographic,
				}
				result = append(result, token)
				*position++
			}
			i = wordEnd
		} else {
			i++
		}
	}

	return result
}

// findWordBoundary 找到词汇边界（简单实现）
func (t *SafeChineseTokenizer) findWordBoundary(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}

	// 简单的规则：
	// 1. 单个字符作为词汇
	// 2. 连续的相同类型字符（如数字、字母）作为词汇
	// 3. 最大词长限制为4个字符

	maxLen := 4
	currentType := t.getCharType(runes[start])

	for i := start + 1; i < len(runes) && i < start+maxLen; i++ {
		charType := t.getCharType(runes[i])

		// 如果字符类型发生变化，在此处断开
		if charType != currentType {
			return i
		}

		// 对于中文字符，通常2-3个字符组成一个词
		if currentType == "chinese" && (i-start) >= 2 {
			// 在这里可以加入更复杂的词典查找逻辑
			return i
		}
	}

	// 如果没有找到合适的边界，返回下一个字符位置
	if start+1 < len(runes) {
		return start + 1
	}
	return len(runes)
}

// getCharType 获取字符类型
func (t *SafeChineseTokenizer) getCharType(r rune) string {
	if unicode.Is(unicode.Han, r) {
		return "chinese"
	}
	if unicode.IsLetter(r) {
		return "letter"
	}
	if unicode.IsDigit(r) {
		return "digit"
	}
	return "other"
}
