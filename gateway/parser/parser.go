package parser

import (
	"io"
	"io/ioutil"
	"path"
	"strings"

	"code.sajari.com/docconv"
)

var ParseAble = map[string]bool{
	".doc":  true,
	".docx": true,
	".pdf":  true,
}

func ParseDoc(f io.Reader, filename string) (string, error) {
	if _, ok := ParseAble[strings.ToLower(path.Ext(filename))]; !ok {
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	mimeType := docconv.MimeTypeByExtension(filename)
	res, err := docconv.Convert(f, mimeType, true)
	if err != nil {
		return "", err
	}
	return res.Body, nil
}
