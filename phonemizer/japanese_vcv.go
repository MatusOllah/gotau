package phonemizer

import (
	"iter"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/SladkyCitron/gotau/voicebank"
)

var _ Phonemizer = (*JapaneseVCV)(nil)

// JapaneseVCV is a Japanese vowel+consonant+vowel (VCV) [Phonemizer].
//
// It extracts the vowel from the final kana of the previous lyric.
//
// [JapaneseVCV] emits candidates based on the following order:
//
//  1. prefix.map + lyric with VCV prefix (if [JapaneseVCV.PrefixMap] and [ResolveConfig.PrevLyric] are present)
//  2. whitespace-trimmed lyric with VCV prefix (if [ResolveConfig.PrevLyric] is present)
//  3. raw lyric with VCV prefix (if [ResolveConfig.PrevLyric] is present)
type JapaneseVCV struct {
	// PrefixMap contains the prefix.map rules for note-based prefix / suffix lookup.
	// Optional.
	PrefixMap voicebank.PrefixMap
}

// Resolve satisfies the [Phonemizer] interface.
func (p *JapaneseVCV) Resolve(cfg ResolveConfig) iter.Seq[string] {
	return func(yield func(string) bool) {
		prev := norm.NFC.String(cfg.PrevLyric)
		lyric := norm.NFC.String(cfg.Lyric)

		vowel := ""
		if cfg.PrevLyric != "" {
			vowel = getLastKanaVowel(prev)
		}

		vcvPrefix := "- "
		if vowel != "" {
			vcvPrefix = vowel + " "
		}

		vcvCombo := vcvPrefix + lyric

		// prefix.map
		if entry, ok := p.PrefixMap[cfg.Note]; ok {
			if !yield(entry.Prefix + vcvCombo + entry.Suffix) {
				return
			}
		}

		// trimmed lyric
		if !yield(vcvPrefix + strings.TrimSpace(lyric)) {
			return
		}

		// raw lyric
		// no need to check yield result; this is the final candidate
		yield(vcvCombo)
	}
}

func getLastKanaVowel(lyric string) string {
	r, _ := utf8.DecodeLastRuneInString(lyric)
	switch r {
	case 'a', 'A',
		'ぁ', 'あ', 'か', 'が', 'さ', 'ざ', 'た', 'だ', 'な', 'は', 'ば', 'ぱ', 'ま', 'ゃ', 'や', 'ら', 'わ',
		'ァ', 'ア', 'カ', 'ガ', 'サ', 'ザ', 'タ', 'ダ', 'ナ', 'ハ', 'バ', 'パ', 'マ', 'ャ', 'ヤ', 'ラ', 'ワ':
		return "a"
	case 'i', 'I',
		'ぃ', 'い', 'き', 'ぎ', 'し', 'じ', 'ち', 'ぢ', 'に', 'ひ', 'び', 'ぴ', 'み', 'り', 'ゐ',
		'ィ', 'イ', 'キ', 'ギ', 'シ', 'ジ', 'チ', 'ヂ', 'ニ', 'ヒ', 'ビ', 'ピ', 'ミ', 'リ', 'ヰ':
		return "i"
	case 'u', 'U',
		'ぅ', 'う', 'く', 'ぐ', 'す', 'ず', 'つ', 'づ', 'ぬ', 'ふ', 'ぶ', 'ぷ', 'む', 'ゅ', 'ゆ', 'る',
		'ゥ', 'ウ', 'ク', 'グ', 'ス', 'ズ', 'ツ', 'ヅ', 'ヌ', 'フ', 'ブ', 'プ', 'ム', 'ュ', 'ユ', 'ル', 'ヴ':
		return "u"
	case 'e', 'E',
		'ぇ', 'え', 'け', 'げ', 'せ', 'ぜ', 'て', 'で', 'ね', 'へ', 'べ', 'ぺ', 'め', 'れ', 'ゑ',
		'ェ', 'エ', 'ケ', 'ゲ', 'セ', 'ゼ', 'テ', 'デ', 'ネ', 'ヘ', 'ベ', 'ペ', 'メ', 'レ', 'ヱ':
		return "e"
	case 'o', 'O',
		'ぉ', 'お', 'こ', 'ご', 'そ', 'ぞ', 'と', 'ど', 'の', 'ほ', 'ぼ', 'ぽ', 'も', 'ょ', 'よ', 'ろ', 'を',
		'ォ', 'オ', 'コ', 'ゴ', 'ソ', 'ゾ', 'ト', 'ド', 'ノ', 'ホ', 'ボ', 'ポ', 'モ', 'ョ', 'ヨ', 'ロ', 'ヲ':
		return "o"
	case 'n', 'N', 'ん', 'ン':
		return "n"
	case 'ー':
		return getLastKanaVowel(lyric[:len(lyric)-1])
	default:
		return ""
	}
}
