package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/anicolaspp/moogle/content"
	"github.com/anicolaspp/moogle/tfidf"
)

const (
	serverContentDif = "/library"
)

type Moogle struct {
	corpus *tfidf.Corpus
}

func (m *Moogle) Run() error {
	contents, err := libraryContent()
	if err != nil {
		return err
	}

	m.corpus = tfidf.NewCorpus()

	docs := []*tfidf.Document{}
	for name, cont := range contents {
		docs = append(docs, tfidf.NewDocument(name, cont))
	}

	m.corpus.FitTransform(docs)

	// scores := m.corpus.AsVector()
	// for _, s := range scores {
	// 	if s.Score > 0 {
	// 		fmt.Println(s)
	// 	}
	// }

	// start serving request here.
	server := http.NewServeMux()
	m.addHandlers(server)

	if err := http.ListenAndServe(":9090", server); err != nil {
		return err
	}

	return nil
}

func (m *Moogle) addHandlers(s *http.ServeMux) {
	s.HandleFunc("/ls", m.listFilesHandler())
	s.HandleFunc("/content", m.content())
	s.HandleFunc("/search", m.search())
}

func (m *Moogle) search() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		fmt.Fprintf(w, "Query: %v\n", query)

		q := tfidf.NewCorpus()
		d := tfidf.NewDocument("q", query)
		q.Add(d)

		fmt.Fprintf(w, "Query TF: %v, Query IDF: %v", q.TF(), q.IDF())

	})
}

func (m *Moogle) listFilesHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for _, d := range m.corpus.Documents() {
			fmt.Fprintln(w, d.String())
		}
	})
}

func (m *Moogle) content() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
}

func libraryContent() (map[string]string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	loader := content.NewLoader(curDir + serverContentDif)
	return loader.ReadContent()
}
