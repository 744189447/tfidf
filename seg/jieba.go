package seg

import (
	"github.com/yanyiwu/gojieba"
)

type JiebaTokenizer struct {
	x *gojieba.Jieba
}

func NewJieba(filePath string) *JiebaTokenizer {
	return &JiebaTokenizer{
		x: gojieba.NewJieba(),
	}
}

func (j *JiebaTokenizer) Seg(text string) []string {
	return j.x.Cut(text, true)
}

func (j *JiebaTokenizer) Free() {
	if j.x != nil {
		j.x.Free()
	}
}

func (j *JiebaTokenizer) ExtractWithWeight(text string, topK int) []gojieba.WordWeight {
	return j.x.ExtractWithWeight(text, topK)
}
