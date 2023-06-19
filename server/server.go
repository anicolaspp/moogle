package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/anicolaspp/moogle/content"
	"github.com/anicolaspp/moogle/tfidf"
)

const (
	serverContentDif = "/content"
)

type Moogle struct {
	fileContents map[string]string
}

func (m *Moogle) Run() error {
	contents, err := libraryContent()
	if err != nil {
		return err
	}

	m.fileContents = contents

	corpus := tfidf.NewCorpus()

	for name, cont := range contents {
		d := tfidf.NewDocument(name, cont)
		corpus.Add(d)
	}

	corpus.TF()

	// start serving RPC request here.

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
}

func (m *Moogle) listFilesHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, _ := range m.fileContents {
			fmt.Fprintln(w, k)
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
