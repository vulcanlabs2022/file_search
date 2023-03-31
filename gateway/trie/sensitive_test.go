package trie

import (
	"encoding/csv"
	"fmt"
	"os"
	"testing"
)

var filename = "./sensitive.csv"

func TestReadSensitive(t *testing.T) {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	word := make([]string, 0)
	for _, record := range records {
		word = append(word, record[0])
	}
	for i := 21000; i < 21007; i++ {
		fmt.Println(word[i])
	}
}

func TestSensitive(t *testing.T) {
	err := LoadSensitive(filename)
	if err != nil {
		t.Fatal(err)
	}
	content := "二维切片中。最后，通过循环迭代切片中的每这段ash二维切片中。最后，通过循环迭代切片中的每adfs平郭老 七二维切片中。最后，通过循环迭代二维切片中。最后，通过循环迭代切片中的每切片中的每asdfasd ，adsjfi什么104在邓 的纽约时报二维切片中。最后，通过循环迭代切片中的每二维切片中。最后，通过循环迭代切片中的每二维切片中。最后，通过循环迭代切片中的每二维切片中。最后，通过循环迭代切片中的每"
	fmt.Println(IsSensitive(content))
	fmt.Println(len(content))
}

func BenchmarkSensitive(b *testing.B) {
	err := LoadSensitive(filename)
	if err != nil {
		b.Fatal(err)
	}
	content := "二维切片中。最后，通过循环迭代切片中的每这段ash二维切片中。最后，通过循环迭代切片中的每adfs平郭老 七二维切片中。最后，通过循环迭代二维切片中。最后，通过循环迭代切片中的每切片中的每asdfasd ，adsjfi什么104在邓 的纽约时报二维切片中。最后，通过循环迭代切片中的每二维切片中。最后，通过循环迭代切片中的每二维切片中。最后，通过循环迭代切片中的每二维切片中。最后，通过循环迭代切片中的每"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSensitive(content)
	}
}
