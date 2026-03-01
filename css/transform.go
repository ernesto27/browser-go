package css

import "strings"

// ApplyTextTransform transforms text based on CSS text-transform and font-variant.
func ApplyTextTransform(text, transform, variant string) string {
	switch transform {
	case "uppercase":
		text = strings.ToUpper(text)
	case "lowercase":
		text = strings.ToLower(text)
	case "capitalize":
		text = CapitalizeWords(text)
	}

	if strings.ToLower(variant) == "small-caps" {
		return strings.ToUpper(text)
	}

	return text
}

// CapitalizeWords capitalizes the first letter of each word
func CapitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
