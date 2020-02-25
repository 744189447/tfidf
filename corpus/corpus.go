package corpus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math"
	"path"
	"sort"
	"tfidf/seg"
	"tfidf/util"
)

var errCategoryNotFound = errors.New("category not found")

type Corpus struct {
	db        *bolt.DB
	tokenizer seg.Tokenizer
	stopWords map[string]struct{}
}

func NewCorpus(filePath string, categories []string) *Corpus {
	db, err := bolt.Open(path.Join(filePath, "corpus.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	tokenizer := seg.NewJieba(filePath)

	corpus := &Corpus{
		db:        db,
		tokenizer: tokenizer,
		stopWords: make(map[string]struct{}),
	}
	err = corpus.initCorpus(filePath, categories)
	if err != nil {
		log.Fatal(err)
	}
	return corpus
}

func (c *Corpus) Free() error {
	c.tokenizer.Free()
	return c.db.Close()
}

func (c *Corpus) initCorpus(filePath string, categories []string) error {
	err := c.db.Batch(func(tx *bolt.Tx) error {
		for _, category := range categories {
			_, err := tx.CreateBucketIfNotExists([]byte(category))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
		}

		lines, err := util.ReadLines(path.Join(filePath, "stopwords"))
		if err != nil {
			return fmt.Errorf("init stop words: %s", err)
		}

		for _, w := range lines {
			c.stopWords[w] = struct{}{}
		}
		c.stopWords[" "] = struct{}{}
		c.stopWords["\n"] = struct{}{}
		return nil
	})
	return err
}

func (c *Corpus) AddCategory(category string) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(category))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return err
}

func (c *Corpus) DelCategory(category string) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(category)); err != nil {
			return err
		}
		return nil
	})
	return err
}

// AddDocs add train documents
func (c *Corpus) AddDocs(category string, doc string) error {
	doc = util.TrimHtml(doc)
	termFreq := c.termFreq(doc)

	err := c.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(category))
		if b == nil {
			return errCategoryNotFound
		}
		var docCount int
		docCountKey := []byte(category + "~$N^&@")
		v := b.Get(docCountKey)
		if v != nil {
			docCount = btoi(v)
		}
		docCount++
		err := b.Put(docCountKey, itob(docCount))
		if err != nil {
			return err
		}

		for term := range termFreq {
			var termCount int
			termKey := []byte(term)
			v := b.Get(termKey)
			if v != nil {
				termCount = btoi(v)
			}
			termCount++
			err := b.Put(termKey, itob(termCount))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// Cal calculate tf-idf weight for specified document
func (c *Corpus) Cal(category string, doc string) (weight map[string]float64, err error) {
	weight = make(map[string]float64)
	doc = util.TrimHtml(doc)
	termFreq := c.termFreq(doc)
	docTerms := 0
	for _, freq := range termFreq {
		docTerms += freq
	}
	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(category))
		if b == nil {
			return errCategoryNotFound
		}
		var docCount int
		docCountKey := []byte(category + "~$N^&@")
		v := b.Get(docCountKey)
		if v != nil {
			docCount = btoi(v)
		}
		for term, freq := range termFreq {
			var termCount int
			v := b.Get([]byte(term))
			if v != nil {
				termCount = btoi(v)
			}
			weight[term] = tfidf(float64(freq), docTerms, termCount, docCount)
		}
		return nil
	})
	return weight, err
}

type WeightTerm struct {
	Term   string
	Weight float64
}

// ExtractTags Extract tags according to tf-idf weight for specified document
func (c *Corpus) ExtractTags(category string, doc string, topK int) (wTerms []*WeightTerm, err error) {
	wTerms = make([]*WeightTerm, 0)
	terms, err := c.Cal(category, doc)
	if err != nil {
		return
	}
	for k, v := range terms {
		wTerm := &WeightTerm{
			Term:   k,
			Weight: v,
		}
		wTerms = append(wTerms, wTerm)
	}
	sort.Slice(wTerms, func(a, b int) bool {
		return wTerms[a].Weight > wTerms[b].Weight
	})
	if topK == 0 {
		topK = 20
	}
	if len(wTerms) > topK {
		wTerms = wTerms[:topK]
	}
	return wTerms, nil
}

func (c *Corpus) ExtractWithWeight(doc string, topK int) (wTerms []*WeightTerm) {
	doc = util.TrimHtml(doc)
	wTerms = make([]*WeightTerm, 0)
	if topK == 0 {
		topK = 20
	}
	ww := c.tokenizer.ExtractWithWeight(doc, topK)
	for _, v := range ww {
		wt := &WeightTerm{
			Term:   v.Word,
			Weight: v.Weight,
		}
		wTerms = append(wTerms, wt)
	}
	return wTerms
}

func (c *Corpus) termFreq(doc string) (m map[string]int) {
	m = make(map[string]int)
	tokens := c.tokenizer.Seg(doc)
	if len(tokens) == 0 {
		return
	}
	for _, term := range tokens {
		if _, ok := c.stopWords[term]; ok {
			continue
		}
		m[term]++
	}
	return
}

func tfidf(termFreq float64, docTerms, termDocs, N int) float64 {
	tf := termFreq / float64(docTerms)
	idf := math.Log(float64(1+N) / (1 + float64(termDocs)))
	return tf * idf
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}
