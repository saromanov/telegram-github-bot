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
	botapi        *tgbotapi.BotAPI
	client        *github.Client
	fsm           *FSM
	updatemessage int
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
	tgb.fsm = NewFSM()
	tgb.registerStates()
	return tgb
}

func (tgb *Telgitbot) registerStates() {
	tgb.fsm.AddState("begin", []string{"auth", "repos", "collaborators", "issues"},
		[]string{"/auth", "/repos"})
	tgb.fsm.AddState("auth", []string{"begin", "dataauth"}, []string{"", " "})
	tgb.fsm.AddState("repos", []string{"begin"}, []string{""})
	tgb.fsm.AddState("collaborators", []string{"begin"}, []string{""})
	tgb.fsm.AddState("dataauth", []string{"begin"}, []string{""})
	tgb.fsm.AddState("issues", []string{"issues_enter"}, []string{""})
	tgb.fsm.AddState("issues_enter", []string{""}, []string{""})
}

func (tgb *Telgitbot) Process(idmsg int, state, text string) {
	if strings.HasPrefix(text, "/") && tgb.fsm.ExistNextState(state) {
		tgb.fsm.SetState(state)
	} else {
		state = tgb.fsm.CurrentState()
	}

	switch state {
	case "auth":
		tgb.auth(idmsg, text)
		tgb.fsm.SetState("dataauth")
	case "dataauth":
		tgb.dataauth(idmsg, text)
		tgb.fsm.SetState("begin")
	case "repos":
		tgb.repos(idmsg, text)
		tgb.fsm.SetState("begin")
	case "issues":
		tgb.issues(idmsg)
		tgb.fsm.SetState("issues_enter")
	case "issues_enter":
		tgb.issues_enter(idmsg, text)
		tgb.fsm.SetState("begin")
	}
}

func (tgb *Telgitbot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	err := tgb.botapi.UpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	tgb.fsm.SetState("begin")
	for {
		for update := range tgb.botapi.Updates {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID
			text := tgb.prepareInput(update.Message.Text)
			state := tgb.prepareState(text)

			//tgb.fsm.SetState(state)
			tgb.Process(update.Message.Chat.ID, state, text)
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
			} else {
				fmt.Println(update.Message.Text)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (tgb *Telgitbot) auth(idmsg int, inp string) {
	msg := tgbotapi.NewMessage(idmsg, "Set your username and password")
	tgb.botapi.SendMessage(msg)
}

func (tgb *Telgitbot) dataauth(idmsg int, text string) {
	check := strings.Index(text, ":")
	if check == -1 {
		msg := tgbotapi.NewMessage(idmsg,
			"username and password must be in the format Username:Passord")
		tgb.botapi.SendMessage(msg)
	}
}

func (tgb *Telgitbot) repos(idmsg int, reponame string) {
	username := strings.Split(reponame, "_")[1]
	if len(username) == 0 {
		return
	}

	//opt := &github.RepositoryListOptions{Sort: "updated"}
	repos, _, err := tgb.client.Repositories.List(username, nil)
	if err != nil {
		fmt.Println("error: %v\n\n", err)
	} else {
		result := ""
		for _, repo := range repos {
			result += *repo.FullName + "\n"
		}
		msg := tgbotapi.NewMessage(idmsg, result)
		tgb.botapi.SendMessage(msg)
	}
}

func (tgb *Telgitbot) issues(idmsg int) {
	msg := tgbotapi.NewMessage(idmsg, "Enter owner:repo")
	tgb.botapi.SendMessage(msg)
}

func (tgb *Telgitbot) issues_enter(idmsg int, repoinfo string) {
	splitter := strings.Split(repoinfo, ":")
	if len(splitter) != 2 {
		return
	}

	owner := splitter[0]
	repo := splitter[1]
	items, _, err := tgb.client.Issues.ListByRepo(owner, repo, nil)
	if err != nil {
		fmt.Println("error: %v\n\n", err)
	} else {
		result := ""
		for _, iss := range items {
			result += *iss.Title + "\n"
		}
		msg := tgbotapi.NewMessage(idmsg, result)
		tgb.botapi.SendMessage(msg)
	}
}

//prepareInput provides getting "standard" data from request
func (tgb *Telgitbot) prepareInput(inp string) string {
	return strings.ToLower(inp)
}

//Get "clean" inpute command for state of FSM
func (tgb *Telgitbot) prepareState(inp string) string {
	result := inp
	if strings.HasPrefix(inp, "/") {
		result = result[1:]
	}

	if strings.Index(result, "_") != -1 {
		splitter := strings.Split(result, "_")
		result = splitter[0]
	}

	if strings.Index(result, " ") != -1 {
		splitter := strings.Split(result, " ")
		result = splitter[0]
	}
	return result
}

func (tgb *Telgitbot) findByStars(title string) {

}
