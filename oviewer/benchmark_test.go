package oviewer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParseString_Normal(b *testing.B) {
	Parse_Helper(b, filepath.Join(testdata, "normal.txt"))
}

func BenchmarkParseString_AnsiEscape(b *testing.B) {
	Parse_Helper(b, filepath.Join(testdata, "ansiescape.txt"))
}

func BenchmarkParseString_ChromaTerm(b *testing.B) {
	b.Helper()
	Parse_Helper(b, filepath.Join(testdata, "ct.log"))
}

func Parse_Helper(b *testing.B, fileName string) {
	b.Helper()
	f, err := os.ReadFile(fileName)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StrToContents(string(f), 8)
	}
}

func BenchmarkDraw_Normal(b *testing.B) {
	Draw_Helper(b, filepath.Join(testdata, "normal.txt"))
}

func BenchmarkDraw_AnsiEscape(b *testing.B) {
	Draw_Helper(b, filepath.Join(testdata, "ansiescape.txt"))
}

func BenchmarkDraw_ChromaTerm(b *testing.B) {
	Draw_Helper(b, filepath.Join(testdata, "ct.log"))
}

func Draw_Helper(b *testing.B, fileName string) {
	root, err := Open(fileName)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	root.ViewSync(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.draw(ctx)
		root.Doc.ClearCache()
	}
	root.Close()
}
