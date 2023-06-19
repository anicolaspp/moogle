// Package tfidf contains a series of functions to process files.
package tfidf

import (
	"fmt"
	"math"
	"strings"
)

type Corpus struct {
	docs []*Document

	vocabolary map[string]bool

	// Term Frequency of each word on each document.
	tf map[string]map[*Document]float64
	// Inverse Document Frequency. Proportion of documents containing each word.
	idf map[string]float64
	// TF_IDF scores of each word on each document.
	tfidf map[string]map[*Document]float64
}

func NewCorpus() *Corpus {
	return &Corpus{
		docs:       []*Document{},
		vocabolary: map[string]bool{},
	}
}

// FitTransform learns the vocabulary and transforms the corpus into tfidf
// scores.
func (c *Corpus) FitTransform(docs []*Document) {
	for _, d := range docs {
		c.Add(d)
	}

	c.TF()
	c.IDF()

	tfidf := map[string]map[*Document]float64{}

	for w := range c.vocabolary {
		for _, d := range docs {
			tfidf[w] = map[*Document]float64{}
			tfidf[w][d] = c.tf[w][d] * c.idf[w]
		}
	}

	c.tfidf = tfidf
}

func (c *Corpus) Transform(query string) {

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
		c.vocabolary[w] = true
	}

	return true
}

func (c *Corpus) TF() map[string]map[*Document]float64 {
	tf := map[string]map[*Document]float64{} // Corpus TF.

	for _, d := range c.docs {
		for w := range c.vocabolary {
			tf[w] = map[*Document]float64{}
			tf[w][d] = d.TF(w)
		}
	}

	c.tf = tf
	return tf
}

func (c *Corpus) IDF() map[string]float64 {
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

	c.idf = idf
	return idf
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
	name    string         // name of the document.
	raw     []string       // document words.
	content string         // oroginal content.
	words   map[string]int // set of unique words in the document.
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
