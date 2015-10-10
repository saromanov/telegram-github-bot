package telgitbot

import
(
	"strings"
	"sort"
)


//https://gist.github.com/ikbear/4038654
type sortedMap struct {
	m map[string]int
	s []string
}

func (sm *sortedMap) Len() int {
	return len(sm.m)
}

func (sm *sortedMap) Less(i, j int) bool {
	return sm.m[sm.s[i]] > sm.m[sm.s[j]]
}

func (sm *sortedMap) Swap(i, j int) {
	sm.s[i], sm.s[j] = sm.s[j], sm.s[i]
}

func sortedKeys(m map[string]int) []string {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]string, len(m))
	i := 0
	for key, _ := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

func IsYes(text string)bool {
	//Repove spaces
	value := strings.Replace(text, " ", "", -1)
	message := strings.ToLower(value)
	msgs := []string{"yes", "yeah", "1", "right", "yep", "ok", "well"}
	for _, item := range msgs {
		if item == message {
			return true
		}
	}
	return false
}

func GetCommonWords(lines []string, oldquery string, num int) string{
	//stopwords:= []string{"with", "this", "like"}
	result:= map[string]int{}
	for _, line := range lines {
		words := strings.Split(line, " ")
		for _, word := range words {
			word = strings.ToLower(word)
			if len(word) > 3 && strings.Index(oldquery, word) == -1{
				_, ok := result[word]
				if !ok {
				result[word] = 1
				} else {
				result[word]++
				}
			}
		}
	}

	sorted := sortedKeys(result)[0:num]
	return strings.Join(sorted, " ")
}