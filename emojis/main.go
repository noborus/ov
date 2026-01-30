package main

import (
	"fmt"

	"github.com/mattn/go-runewidth"
)

func main() {
	// 実際のあいまい幅文字
	ambiguousChars := []rune{
		'§', // Section Sign
		'±', // Plus-Minus
		'°', // Degree Sign
		'×', // Multiplication Sign
		'÷', // Division Sign
		'·', // Middle Dot
		'¿', // Inverted Question Mark
		'¡', // Inverted Exclamation Mark
		'ª', // Feminine Ordinal Indicator
		'º', // Masculine Ordinal Indicator
	}

	fmt.Println("Ambiguous Width Characters:")
	fmt.Println("Char | Unicode | Width | IsAmbiguous")
	fmt.Println("-----|---------|-------|------------")

	for _, r := range ambiguousChars {
		width := runewidth.RuneWidth(r)
		isAmb := runewidth.IsAmbiguousWidth(r)
		fmt.Printf("%-4c | U+%04X  | %-5d | %v\n", r, r, width, isAmb)
	}

	// 絵文字は Ambiguous ではない
	fmt.Println("\nEmoji Characters:")

	emojis := []rune{'🛤', '🚏', '🚃', '🚌'}
	for _, r := range emojis {
		width := runewidth.RuneWidth(r)
		isAmb := runewidth.IsAmbiguousWidth(r)
		fmt.Printf("%-4c | U+%04X  | %-5d | %v\n", r, r, width, isAmb)
	}

	fmt.Println(runewidth.IsAmbiguousWidth('é'))
	fmt.Println(runewidth.RuneWidth('é'))

	latinChars := []rune{'à', 'á', 'â', 'ã', 'ä', 'å', 'æ', 'ç', 'è', 'é', 'ê', 'ë'}

	for _, r := range latinChars {
		isAmb := runewidth.IsAmbiguousWidth(r)
		width := runewidth.RuneWidth(r)
		fmt.Printf("'%c' (U+%04X): IsAmbiguous=%v, Width=%d\n", r, r, isAmb, width)
	}
}
