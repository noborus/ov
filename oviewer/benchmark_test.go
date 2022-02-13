package oviewer

import (
	"os"
	"testing"
)

func BenchmarkParseString_Normal(b *testing.B) {
	Parse_Helper(b, "../testdata/normal.txt")
}

func BenchmarkParseString_AnsiEscape(b *testing.B) {
	Parse_Helper(b, "../testdata/ansiescape.txt")
}

func BenchmarkParseString_ChromaTerm(b *testing.B) {
	Parse_Helper(b, "../testdata/ct.log")
}

func Parse_Helper(b *testing.B, fileName string) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseString(string(f), 8)
	}
}

func BenchmarkDraw_Normal(b *testing.B) {
	Draw_Helper(b, "../testdata/normal.txt")
}

func BenchmarkDraw_AnsiEscape(b *testing.B) {
	Draw_Helper(b, "../testdata/ansiescape.txt")
}

func BenchmarkDraw_ChromaTerm(b *testing.B) {
	Draw_Helper(b, "../testdata/ct.log")
}

func Draw_Helper(b *testing.B, fileName string) {
	root, err := Open(fileName)
	if err != nil {
		b.Fatal(err)
	}
	root.ViewSync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.draw()
		root.Doc.ClearCache()
	}
	root.Close()
}
