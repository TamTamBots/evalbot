package main

import (
	"fmt"
	"html"
	"log"
	"strings"
	"os"

	"github.com/anonyindian/gottbot"
	"github.com/anonyindian/gottbot/ext"
	"github.com/anonyindian/gottbot/filters"
	"github.com/anonyindian/gottbot/handlers"
	piston "github.com/milindmadhukar/go-piston"
)

var client *piston.Client
var (
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
	dispatcher.AddHandler(handlers.CommandHandler("help", help))
	dispatcher.AddHandler(handlers.CommandHandler("languages", languagesC))
	dispatcher.AddHandler(handlers.BotAddedHandler(botadded))
	dispatcher.AddHandlerToGroup(1, handlers.MessageHandler(filters.Message.Prefix("/"), langfound(eval)))

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
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveUser
	msg.Reply(bot, fmt.Sprintf(`
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
