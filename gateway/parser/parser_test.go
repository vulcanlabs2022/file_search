package parser

import (
	"fmt"
	"os"
	"testing"

	"code.sajari.com/docconv"
)

func TestParseDoc(t *testing.T) {
	file, err := os.Open("../testfile/test.docx")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	s, err := ParseDoc(file, "test.docx")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}

func TestParsePDF(t *testing.T) {
	file, err := os.Open("../testfile/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	s, err := ParseDoc(file, "test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}

func TestDocx(t *testing.T) {
	res, err := docconv.ConvertPath("../testfile/test.docx")
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Body)
}

func TestPdf(t *testing.T) {
	res, err := docconv.ConvertPath("../testfile/test.pdf")
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Body)
}
