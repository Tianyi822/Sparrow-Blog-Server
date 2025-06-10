package tokenizer

import (
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/yanyiwu/gojieba"
	"strings"
)

type ChineseTokenizer struct {
	handle *gojieba.Jieba
}

func NewChineseTokenizer() *ChineseTokenizer {
	return &ChineseTokenizer{
		// 使用精简词典减少内存占用
		handle: gojieba.NewJieba(gojieba.DICT_PATH, gojieba.HMM_PATH, gojieba.USER_DICT_PATH),
	}
}

func (t *ChineseTokenizer) Tokenize(sentence []byte) analysis.TokenStream {
	result := make(analysis.TokenStream, 0, 20)
	words := t.handle.Cut(string(sentence), true)

	currentPosition := 0 // 当前字节位置
	position := 1        // 词语位置计数器

	for _, word := range words {
		trimmed := strings.TrimSpace(word)
		if trimmed == "" {
			currentPosition += len(word)
			continue
		}

		// 计算实际位置（需要知道词语在原文中的偏移量）
		start := currentPosition
		end := start + len(trimmed)
		currentPosition = end + (len(word) - len(trimmed)) // 调整空白字符

		token := analysis.Token{
			Term:     []byte(trimmed),
			Start:    start,
			End:      end,
			Position: position,
			Type:     analysis.Ideographic,
		}
		result = append(result, &token)
		position++
	}
	return result
}
