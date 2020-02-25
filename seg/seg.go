package seg

import "github.com/yanyiwu/gojieba"

type Tokenizer interface {
	Seg(text string) []string
	Free()
	ExtractWithWeight(text string, topK int) []gojieba.WordWeight
}
