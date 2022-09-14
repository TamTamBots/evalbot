package main

import (
	"fmt"
	"html"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/anonyindian/gottbot"
	"github.com/anonyindian/gottbot/ext"
	"github.com/anonyindian/gottbot/filters"
	"github.com/anonyindian/gottbot/handlers"
	piston "github.com/milindmadhukar/go-piston"
)

const (
	LogsGroup = -86066006261922
	OwnerId   = 590383618466
)

var (
	client *piston.Client

	languages      *[]string
	availableLangs string
)

func main() {
	client = piston.CreateDefaultClient()
	languages = client.GetLanguages()
	initLanguagesString()

	bot, err := gottbot.NewBot(os.Getenv("TT_BOT_TOKEN"), nil)
	if err != nil {
		panic(err)
	}
	updater := ext.NewUpdater(nil)
	updater.StartPolling(bot, nil)

	dispatcher := updater.Dispatcher

	dispatcher.AddHandler(handlers.CommandHandler("start", start))
	dispatcher.AddHandler(handlers.CommandHandler("id", id))
	dispatcher.AddHandler(handlers.CommandHandler("help", help))
	dispatcher.AddHandler(handlers.CommandHandler("languages", languagesC))
	dispatcher.AddHandler(handlers.CommandHandler("stats", ownerOnly(stats)))
	dispatcher.AddHandler(handlers.BotStartedHandler(botstarted))
	dispatcher.AddHandler(handlers.BotAddedHandler(botadded))
	dispatcher.AddHandlerToGroup(1, &handlers.Message{
		Response:    langfound(eval),
		Filter:      filters.Message.Prefix("/"),
		AllowEdited: true,
	})

	fmt.Println("Started eval bot with long polling...")

	updater.Idle()
}

func initLanguagesString() {
	availableLangs = "Here is the list of available languages:"
	for _, lang := range *languages {
		availableLangs += fmt.Sprintf("\n- <code>%s</code>", lang)
	}
	availableLangs += "\n\n<b>Usage</b>: <code>/python print(\"hello\")</code>"
}

func languagesC(bot *gottbot.Bot, ctx *ext.Context) error {
	_, _ = ctx.EffectiveMessage.Reply(bot, availableLangs, &gottbot.SendMessageOpts{
		DisableLinkPreview: true,
		Format:             gottbot.Html,
	})
	return ext.EndGroups
}

func botstarted(bot *gottbot.Bot, ctx *ext.Context) error {
	_, _ = bot.SendMessage(ctx.EffectiveChatId, "", &gottbot.SendMessageOpts{
		Attachments: []gottbot.AttachmentRequest{{
			Payload: &gottbot.StickerPayload{
				Code: "d7c1fd51c4",
			},
		}},
	})

	user := ctx.EffectiveUser
	var text string
	if user.Username != "" {
		text = fmt.Sprintf(`<a href="https://tamtam.chat/%s">%s</a> started the bot.`, user.Username, html.EscapeString(user.Name))
	} else {
		text = fmt.Sprintf(`<a href="tamtam://user/%d">%s</a> started the bot.`, user.UserId, html.EscapeString(user.Name))
	}
	_, _ = bot.SendMessage(LogsGroup, text, &gottbot.SendMessageOpts{
		Format: gottbot.Html,
	})
	return start(bot, ctx)
}

func botadded(bot *gottbot.Bot, ctx *ext.Context) error {
	if ctx.BotAdded.IsChannel {
		return ext.EndGroups
	}
	_, err := bot.SendMessage(ctx.BotAdded.ChatId,
		`
Hello! I'm Eval Bot. Thanks for adding me to this group.
Hit <code>/help</code> for more info.
`,
		&gottbot.SendMessageOpts{
			Format: gottbot.Html,
		})
	if err != nil {
		fmt.Println("failed to greet:", err.Error())
	}
	return ext.EndGroups
}

func start(bot *gottbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveUser
	if user == nil {
		user = new(gottbot.User)
	}
	_, _ = bot.SendMessage(ctx.EffectiveChatId, fmt.Sprintf(`
Hey %s, I am a bot to eval codes of various languages.
I'm built using the piston eval engine and the gottbot library.
Hit <code>/help</code> for info related to commands.   
	`, html.EscapeString(user.Name)), &gottbot.SendMessageOpts{
		Attachments: []gottbot.AttachmentRequest{{
			Payload: &gottbot.ButtonsPayload{
				Buttons: [][]gottbot.Button{{
					&gottbot.LinkButton{
						Text: "Piston",
						Url:  "https://github.com/engineer-man/piston",
					},
					&gottbot.LinkButton{
						Text: "GoTTBot",
						Url:  "https://github.com/anonyindian/gottbot",
					},
				}},
			},
		}},
		Format: gottbot.Html,
	})
	return ext.EndGroups
}

func help(bot *gottbot.Bot, ctx *ext.Context) error {
	_, _ = ctx.EffectiveMessage.Reply(bot, `
Here is the list of all commands:

- <code>/start</code>: Start the bot
- <code>/help</code>: Prints this message
- <code>/id</code>: Get id of current chat and user
- <code>/languages</code>: List all supported languages
- <code>/{language} eval code...</code> : Eval a code in 'language'

<b>Example:</b>
<- To print "Hello World" in python:
-> <code>/python print("Hello World")</code> 
	`, &gottbot.SendMessageOpts{
		Format: gottbot.Html,
	})
	return ext.EndGroups
}

func langfound(callback handlers.Callback) handlers.Callback {
	return func(bot *gottbot.Bot, ctx *ext.Context) error {
		msg := ctx.EffectiveMessage
		lang := strings.Fields(msg.Body.Text)[0][1:]
		for _, lang0 := range *languages {
			if strings.ToLower(lang) == lang0 {
				ctx.Data = map[string]any{
					"lang": lang,
				}
				return callback(bot, ctx)
			}
		}
		return ext.EndGroups
	}
}

func id(bot *gottbot.Bot, ctx *ext.Context) error {
	var text string
	if user := ctx.EffectiveUser; user != nil {
		text = fmt.Sprintf("**User ID**: `%d`\n", user.UserId)
	}
	text += fmt.Sprintf("**Chat ID**: `%d`", ctx.EffectiveChatId)
	_, _ = ctx.EffectiveMessage.Reply(bot,
		text,
		&gottbot.SendMessageOpts{
			Format: gottbot.Markdown,
		})
	return ext.EndGroups
}

func eval(bot *gottbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	lang := ctx.Data["lang"].(string)
	text := msg.Body.Text
	text = strings.TrimPrefix(text, "/"+lang)
	text = strings.TrimSpace(text)

	if text == "" {
		_, _ = msg.Reply(bot, "You need to provide me some code to eval.", nil)
		return ext.EndGroups
	}

	output, err := client.Execute(strings.ToLower(lang), "",
		[]piston.Code{
			{Content: text},
		},
	)
	if err != nil {
		_, _ = msg.Reply(bot, fmt.Sprintf("failed to eval: %s", err.Error()), nil)
		return ext.EndGroups
	}

	out := output.GetOutput()

	if out == "" {
		out = "No Output"
	}

	_, err = msg.Reply(bot, output.GetOutput(), nil)
	if err != nil {
		log.Println("failed to send message:", err.Error())
	}
	return ext.EndGroups
}

func ownerOnly(cb handlers.Callback) handlers.Callback {
	return func(bot *gottbot.Bot, ctx *ext.Context) error {
		if ctx.EffectiveUser != nil && ctx.EffectiveUser.UserId == OwnerId {
			return cb(bot, ctx)
		}
		return ext.EndGroups
	}
}

func stats(bot *gottbot.Bot, ctx *ext.Context) error {
	text := fmt.Sprintf(`
**Stats:**
Go Version: %s
Goroutines: %d
CPUs: %d
	`, runtime.Version(), runtime.NumGoroutine(), runtime.NumCPU())
	_, _ = ctx.EffectiveMessage.Reply(bot, text, &gottbot.SendMessageOpts{
		Format: gottbot.Markdown,
	})
	return ext.EndGroups
}
