//go:generate go run ./internal/genvoweltable/.

package phonemizer

import "unicode/utf8"

func kanaTailVowel(s string) byte {
	if s == "" {
		return 0
	}

	r, size := utf8.DecodeLastRuneInString(s)

	// handle chōonpu
	if r == 'ー' {
		return kanaTailVowel(s[:len(s)-size])
	}

	// handle latin
	switch r {
	case 'a', 'A':
		return 'a'
	case 'i', 'I':
		return 'i'
	case 'u', 'U':
		return 'u'
	case 'e', 'E':
		return 'e'
	case 'o', 'O':
		return 'o'
	}

	if r < kanaTableStart || r > kanaTableEnd {
		return 0
	}
	return kanaVowelTable[r-kanaTableStart]
}
