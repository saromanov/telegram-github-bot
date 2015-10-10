package telgitbot

import
(
	"strings"
)

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