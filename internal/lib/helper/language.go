package helper

import (
	"regexp"
	"strings"
)

// GetIndefiniteArticle returns "a" or "an" depending on the wirst Word in `phrase`
// inspired by https://stackoverflow.com/questions/4558437/programmatically-determine-whether-to-describe-an-object-with-a-or-an
func GetIndefiniteArticle(phrase string) string {
	re := regexp.MustCompile(`\w+`)
	word := re.FindString(phrase)
	if word == "" {
		return "an"
	}

	iWord := strings.ToLower(word)

	if strings.HasPrefix(iWord, "hour") && !strings.HasPrefix(iWord, "houri") {
		return "an"
	}
	for _, anWord := range []string{"euler", "heir", "honest", "hono"} {
		if strings.HasPrefix(iWord, anWord) {
			return "an"
		}
	}

	if len(iWord) == 1 {
		if strings.ContainsAny(iWord, "aedhilmnorsx") {
			return "an"
		}
		return "a"
	}
	if matched, _ := regexp.MatchString(
		"(?!FJO|[HLMNS]Y.|RY[EO]|SQU|(F[LR]?|[HL]|MN?|N|RH?|S[CHKLMNPTVW]?|X(YL)?)[AEIOU])[FHLMNRSX][A-Z]",
		word,
	); matched {
		return "an"
	}

	for _, pattern := range []string{"^e[uw]", "^onc?e\b", "^uni([^nmd]|mo)", "^u[bcfhjkqrst][aeiou]"} {
		if matched, _ := regexp.MatchString(pattern, iWord); matched {
			return "a"
		}
	}

	if matched, _ := regexp.MatchString("^U[NK][AIEO]", word); matched {
		return "a"
	} else if word == strings.ToUpper(word) {
		if strings.ContainsAny(string(iWord[0]), "aedhilmnorsx") {
			return "an"
		}
		return "a"
	}

	if strings.ContainsAny(string(iWord[0]), "aeiou") {
		return "an"
	}

	if matched, _ := regexp.MatchString("^y(b[lor]|cl[ea]|fere|gg|p[ios]|rou|tt)", iWord); matched {
		return "an"
	}
	return "a"
}
