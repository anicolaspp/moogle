// Package tfidf contains a series of functions to process files.
package tfidf

import (
	"fmt"
	"math"
	"strings"

	sim "github.com/khaibin/go-cosinesimilarity"
	"golang.org/x/exp/slices"
)

type wordSet map[string]bool

func (ws wordSet) add(w string) bool {
	_, ok := ws[w]
	if ok {
		return false
	}

	ws[w] = true
	return true
}

// Corspus represents the entire library and the corresponding transformations
// we can apply to it.
type Corpus struct {
	docs []*Document

	// unique set of words in the corpus.
	vocabolary wordSet

	// TF_IDF scores of each word on each document.
	tfidf map[string]map[*Document]float64
}

// NewCorpus creates a new, empty Corpus.
func NewCorpus() *Corpus {
	return &Corpus{
		docs:       []*Document{},
		vocabolary: wordSet{},
	}
}

// Add adds a document to the library if it does not exist and returns true,
// returns false if the document is already in the library.
func (c *Corpus) Add(d *Document) bool {
	for _, doc := range c.docs {
		if d.name == doc.name {
			return false
		}
	}

	c.docs = append(c.docs, d)
	for w := range d.words {
		c.vocabolary.add(w)
	}

	return true
}

// Transform calculate the tf-idf in the corpus.
func (c *Corpus) Transform() {
	c.transform()
}

// FitTransform learns the vocabulary and transforms the corpus into tfidf
// scores based on the given documents.
//
// This is the same as adding each doc using
// corpus.Add then calling corpus.Transform().
func (c *Corpus) FitTransform(docs []*Document) {
	// Fit the documents.
	for _, d := range docs {
		c.Add(d)
	}

	// calculate tf-idf.
	c.Transform()
}

// VectorScore is the scores of corpus' word in a particular document.
type VectorScore map[string][]Score

// RankDocs returns the list of documents similar to the query order by
// similarities.
func (c *Corpus) RankDocs(query VectorScore) []string {
	vects := c.Vectorize()
	drank := []rank{}

	for _, d := range c.docs {
		v := VectorScore{}
		v[d.name] = vects[d.name]

		rank := rank{document: d.name, rank: c.cosineSimilarities(v, query)}
		drank = append(drank, rank)
	}

	// order the docs based on their similaries to the query.
	slices.SortFunc[rank](drank, func(a, b rank) bool { return a.rank > b.rank })

	res := []string{}
	for _, r := range drank {
		if r.rank > 0 { // remove docs that are completely different.
			res = append(res, r.document)
		}
	}

	return res
}

// Words returns all the words in the Corpus. Basically all words in all
// documents.
func (c *Corpus) Words() []string {
	ws := []string{}
	for _, d := range c.docs {
		ws = append(ws, d.raw...)
	}

	return ws
}

// Documents returns the set of the documents in the library.
func (c *Corpus) Documents() []*Document {
	return c.docs
}

// Transform returns the tf-idf vector of the query document.
func Transform(corpus *Corpus, query string) VectorScore {
	qd := NewDocument("moogle_query", query)
	cc := corpus.clone()
	cc.FitTransform([]*Document{qd})

	v := VectorScore{}
	for _, s := range cc.AsVector() {
		if s.Document == qd.name {
			v[s.Document] = append(v[s.Document], s)
		}
	}

	return v
}

// calculates the fi-idf of the corpus.
func (c *Corpus) transform() {
	tf := c.calculateTF()
	idf := c.calculateIDF()

	tfidf := map[string]map[*Document]float64{}

	for w := range c.vocabolary {
		for _, d := range c.docs {
			tfidf[w] = map[*Document]float64{}
			tfidf[w][d] = tf[w][d] * idf[w]
		}
	}

	c.tfidf = tfidf
}

// clones the corpus without modifying it.
func (c *Corpus) clone() *Corpus {
	cc := NewCorpus()
	cc.docs = c.docs
	cc.vocabolary = c.vocabolary

	return cc
}

func (c *Corpus) Vectorize() VectorScore {
	res := VectorScore{}

	for _, s := range c.AsVector() {
		res[s.Document] = append(res[s.Document], s)
	}

	return res
}

// how a particular document ranks agains the other ones when compared to a
// particular query.
type rank struct {
	document string
	rank     float64
}

// cosineSimilarities returns how similar both vectors are.
func (c *Corpus) cosineSimilarities(a, b VectorScore) float64 {
	xscores := []float64{}
	for _, scores := range a {
		for _, s := range scores {
			xscores = append(xscores, s.Score)
		}
	}

	yscores := []float64{}
	for _, scores := range b {
		for _, s := range scores {
			yscores = append(yscores, s.Score)
		}
	}

	y := [][]float64{yscores}
	x := [][]float64{xscores}
	res := sim.Compute(x, y)

	return res[0][0]
}

func (c *Corpus) calculateTF() map[string]map[*Document]float64 {
	tf := map[string]map[*Document]float64{} // Corpus TF.

	for _, d := range c.docs {
		for w := range c.vocabolary {
			tf[w] = map[*Document]float64{}
			tf[w][d] = d.TF(w)
		}
	}

	return tf
}

func (c *Corpus) calculateIDF() map[string]float64 {
	df := map[string]int{} // number of documents containing each word.

	for w := range c.vocabolary {
		for _, d := range c.docs {
			if d.Contains(w) {
				df[w]++
			}
		}
	}

	idf := map[string]float64{}

	for w := range c.vocabolary {
		idf[w] = math.Log10(float64(len(c.docs) / df[w]))
	}

	return idf
}

// Score reprensets the TF-IDF score of a word-document pair.
type Score struct {
	Word     string
	Document string
	Score    float64
}

// String returns the string representation of `Score`.
func (s *Score) String() string {
	return fmt.Sprintf("word: %v, doc: %v, score: %v", s.Word, s.Document, s.Score)
}

// AsVector retuns the tf-idf score of each word for each document.
func (c *Corpus) AsVector() []Score {
	res := []Score{}

	for w := range c.vocabolary {
		for _, d := range c.docs {
			score := Score{Word: w, Document: d.name, Score: c.tfidf[w][d]}

			res = append(res, score)
		}
	}

	return res
}

// Document represents a document in the library.
type Document struct {
	name    string   // name of the document.
	raw     []string // document words.
	content string   // oroginal content.
	// set of unique words in the document with their corresponding frequency.
	words map[string]int
}

// NewDocument creates a `Document`.
func NewDocument(name, content string) *Document {
	words := words(content)

	// set of unique words.
	uwords := map[string]int{}
	for _, w := range words {
		uwords[w]++
	}

	return &Document{name: name, raw: words, words: uwords, content: content}
}

// TF returns the tf value of the word in the document.
//
// The number of times the `word` appears in the document diveded by the number
// of words in the document.
func (d *Document) TF(word string) float64 {
	if count, ok := d.words[word]; ok {
		return float64(count) / float64(len(d.raw))
	}

	return 1
}

// Contains returns true if the word is in the document, otherwise false.
func (d *Document) Contains(word string) bool {
	_, ok := d.words[word]

	return ok
}

// String returns the document name.
func (d *Document) String() string {
	return d.name
}

// Content returns the original content of the `Document`.
func (d *Document) Content() string {
	return d.content
}

// words extracts all the words of the given document.
func words(document string) []string {
	symbols := "!\"#$%&()*+-./:;<=>?@[]^_`{|}~\n"

	pieces := strings.ToLower(document)
	for s := range symbols {
		pieces = strings.ReplaceAll(pieces, string(rune(s)), " ")
	}

	pieces = strings.ReplaceAll(pieces, "'", " ")

	return strings.Split(pieces, " ")
}
