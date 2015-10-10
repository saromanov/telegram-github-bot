package telgitbot

import (
	"bytes"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	BEGIN             = "begin"
	AUTH              = "auth"
	REPOS             = "repos"
	COLLABORATORS     = "collaborators"
	DATAAUTH          = "dataauth"
	ISSUES            = "issues"
	ISSUESENTER       = "issues_enter"
	PULLREQUESTS      = "pullrequests"
	PULLREQUESTSENTER = "pullrequests_enter"
	HELP              = "help"
	SEARCH            = "search"
	USER              = "user"
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
	tgb.fsm.AddState(BEGIN, []string{AUTH, "repos", COLLABORATORS, ISSUES, PULLREQUESTS, HELP,
		SEARCH, USER},
		[]string{"/auth", "/repos"})
	tgb.fsm.AddState(AUTH, []string{BEGIN, DATAAUTH}, []string{"", " "})
	tgb.fsm.AddState(REPOS, []string{BEGIN}, []string{""})
	tgb.fsm.AddState(COLLABORATORS, []string{BEGIN}, []string{""})
	tgb.fsm.AddState(DATAAUTH, []string{BEGIN}, []string{""})
	tgb.fsm.AddState(ISSUES, []string{ISSUESENTER, BEGIN}, []string{""})
	tgb.fsm.AddState(ISSUESENTER, []string{BEGIN}, []string{""})
	tgb.fsm.AddState(PULLREQUESTS, []string{PULLREQUESTSENTER, BEGIN}, []string{""})
	tgb.fsm.AddState(PULLREQUESTSENTER, []string{BEGIN}, []string{""})
	tgb.fsm.AddState(HELP, []string{HELP}, []string{})
	tgb.fsm.AddState(SEARCH, []string{BEGIN}, []string{})
	tgb.fsm.AddState(USER, []string{USER}, []string{})
}

func (tgb *Telgitbot) Process(idmsg int, state, text string) {
	if strings.HasPrefix(text, "/") && tgb.fsm.ExistNextState(state) {
		tgb.fsm.SetState(state)
	} else {
		state = tgb.fsm.CurrentState()
	}

	switch state {
	case AUTH:
		tgb.auth(idmsg, text)
		tgb.fsm.SetState(DATAAUTH)
	case DATAAUTH:
		tgb.dataauth(idmsg, text)
		tgb.fsm.SetState(BEGIN)
	case REPOS:
		tgb.repos(idmsg, text)
		tgb.fsm.SetState(BEGIN)
	case ISSUES:
		tgb.issues(idmsg)
		tgb.fsm.SetState(ISSUESENTER)
	case ISSUESENTER:
		tgb.issues_enter(idmsg, text)
		tgb.fsm.SetState(BEGIN)
	case PULLREQUESTS:
		tgb.pullRequests(idmsg)
		tgb.fsm.SetState(PULLREQUESTSENTER)
	case PULLREQUESTSENTER:
		tgb.pullRequestsEnter(idmsg, text)
		tgb.fsm.SetState(BEGIN)
	case HELP:
		tgb.help(idmsg)
		tgb.fsm.SetState(BEGIN)
	case SEARCH:
		tgb.search(idmsg, text)
		tgb.fsm.SetState(BEGIN)
	case USER:
		tgb.user(idmsg, text)
		tgb.fsm.SetState(BEGIN)
	}
}

func (tgb *Telgitbot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	err := tgb.botapi.UpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	tgb.fsm.SetState(BEGIN)
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
	tgb.sendMessage(idmsg, "Set your access token")
}

func (tgb *Telgitbot) dataauth(idmsg int, accesstoken string) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "... your access token ..."},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	tgb.client = github.NewClient(tc)
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
	idx := strings.Index(repoinfo, ":")
	if idx == -1 {
		tgb.fsm.SetState(BEGIN)
		msg := tgbotapi.NewMessage(idmsg, "incorrect format")
		tgb.botapi.SendMessage(msg)
		return
	}

	splitter := strings.Split(repoinfo, ":")
	if len(splitter) != 2 {
		tgb.fsm.SetState(BEGIN)
		msg := tgbotapi.NewMessage(idmsg, "incorrect format")
		tgb.botapi.SendMessage(msg)
		return
	}

	owner := splitter[0]
	repo := splitter[1]
	items, _, err := tgb.client.Issues.ListByRepo(owner, repo, nil)
	if err != nil {
		fmt.Println("error: %v\n\n", err)
	} else {
		result := ""
		for i, iss := range items {
			result += fmt.Sprintf("%d. %s\n", i+1, *iss.Title)
		}
		msg := tgbotapi.NewMessage(idmsg, result)
		tgb.botapi.SendMessage(msg)
	}
}

func (tgb *Telgitbot) pullRequests(idmsg int) {
	tgb.sendMessage(idmsg, "Enter owner:repo:number of PR")
}

func (tgb *Telgitbot) pullRequestsEnter(idmsg int, repoinfo string) {
	idx := strings.Index(repoinfo, ":")
	if idx == -1 {
		tgb.fsm.SetState(BEGIN)
		tgb.sendMessage(idmsg, "incorrect format")
		return
	}

	splitter := strings.Split(repoinfo, ":")
	if len(splitter) != 3 {
		tgb.fsm.SetState(BEGIN)
		tgb.sendMessage(idmsg, "incorrect format")
		return
	}
	owner := splitter[0]
	repo := splitter[1]
	num, err := strconv.Atoi(splitter[2])
	if err != nil {
		tgb.sendMessage(idmsg, "Number of pull requests must be integer")
		return
	}

	items, _, err := tgb.client.PullRequests.ListCommits(owner, repo, num, nil)
	if err != nil {
		fmt.Println("error: %v\n\n", err)
		return
	}

	result := ""
	for i, pr := range items {
		result += fmt.Sprintf("%d. %s:%s\n", i+1, *pr.SHA, *pr.Commit.Message)
	}
	msg := tgbotapi.NewMessage(idmsg, result)
	tgb.botapi.SendMessage(msg)
}

func (tgb *Telgitbot) help(idmsg int) {
	result := bytes.NewBufferString("")
	result.WriteString("/auth - authorization by the token\n")
	result.WriteString("/repos - List of repos by the user\n")
	result.WriteString("/issues - List of issues for project\n")
	result.WriteString("/pullrequests - List of pull requests for project\n")
	result.WriteString("/search - Search repositorires by query\n")
	result.WriteString("/user - Return basic information about user\n")
	tgb.sendMessage(idmsg, result.String())
}

func (tgb *Telgitbot) search(idmsg int, text string) {
	query := strings.Split(text, "/search")
	if len(query) != 2 {
		return
	}

	value := strings.Replace(query[1], " ", "", -1)
	items, _, err := tgb.client.Search.Repositories(value, &github.SearchOptions{})
	if err != nil {
		panic(err)
	}
	result := ""
	for _, item := range items.Repositories {
		result += fmt.Sprintf("%s: %s\n", *item.HTMLURL, *item.Description)
	}

	tgb.sendMessage(idmsg, result)
}

//return to output of telegram list of users
func (tgb *Telgitbot) user(idmsg int, inp string) {
	splitter := strings.Split(inp, "/user")
	if len(splitter) != 2 {
		return
	}
	username := strings.Replace(splitter[1], " ", "", -1)
	user, _, err := tgb.client.Users.Get(username)
	if err != nil {
		if strings.Index(err.Error(), "404") != 0 {
			tgb.sendMessage(idmsg, fmt.Sprintf("User %s is not found", username))
		} else {
			tgb.sendMessage(idmsg, fmt.Sprintf("%v", err))
		}
		return
	}
	msg := tgbotapi.NewMessage(idmsg, fmt.Sprintf("%d", *user.PublicRepos))
	tgb.botapi.SendMessage(msg)
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

func (tgb *Telgitbot) sendMessage(idmsg int, message string) {
	msg := tgbotapi.NewMessage(idmsg, message)
	tgb.botapi.SendMessage(msg)
}

func (tgb *Telgitbot) findByStars(mesg string) {

}
