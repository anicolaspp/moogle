package content

import (
	"io/ioutil"
	"os"
)

type Loader struct {
	basePath string
}

func NewLoader(basePath string) *Loader {
	return &Loader{basePath: basePath}
}

// ReadContent returns the content of each file or an error if it cannot read a
// file.
func (l *Loader) ReadContent() (map[string]string, error) {
	res := map[string]string{}

	infos, err := ioutil.ReadDir(l.basePath)
	if err != nil {
		return res, err
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		path := l.basePath + "/" + info.Name()
		bytes, err := os.ReadFile(path)
		if err != nil {
			return res, err
		}

		content := string(bytes)
		res[info.Name()] = content
	}

	return res, nil
}
