package telgitbot

import (
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/google/go-github/github"
	"log"
	"strings"
	"time"
)

type Telgitbot struct {
	botapi *tgbotapi.BotAPI
	client *github.Client
}

func New(token string) *Telgitbot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	tgb := new(Telgitbot)
	tgb.botapi = bot
	tgb.client = github.NewClient(nil)
	return tgb
}

func (tgb *Telgitbot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	err := tgb.botapi.UpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for {
		for update := range tgb.botapi.Updates {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID
			text := update.Message.Text
			if strings.HasPrefix(text, "/repos_") {
				username := strings.Split(text, "_")[1]
				if len(username) > 0 {
					//opt := &github.RepositoryListOptions{Sort: "updated"}
					repos, _, err := tgb.client.Repositories.List(username, nil)
					if err != nil {
						fmt.Println("error: %v\n\n", err)
					} else {
						result := ""
						for _, repo := range repos {
							result += *repo.FullName + "\n"
						}
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
						tgb.botapi.SendMessage(msg)
					}
				}
			}

			if strings.HasPrefix(text, "/collaborators_") {
				repo := strings.Split(text, "_")[1]
				if len(repo) > 0 {
					repos, _, err := tgb.client.Repositories.ListCollaborators("saromanov", repo, nil)
					if err != nil {
						fmt.Println("error: %v\n\n", err)
					} else {
						result := ""
						for _, repo := range repos {
							result += *repo.Name + "\n"
						}
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
						tgb.botapi.SendMessage(msg)
					}
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (tgb *Telgitbot) findByStars(title string) {

}
