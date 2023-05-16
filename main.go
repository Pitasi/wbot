package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing environment variable %v", key)
	}
	return value
}

var (
	BotToken  = getEnv("BOT_TOKEN")
	OpenAIKey = getEnv("OPENAI_KEY")
	allowlist = map[int64]struct{}{}
)

func init() {
	chatids := getEnv("ALLOWLIST")
	for _, chatid := range strings.Split(chatids, ",") {
		chatid, err := strconv.ParseInt(chatid, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing chatid %v: %v", chatid, err)
		}
		allowlist[chatid] = struct{}{}
	}
}

func main() {
	client := openai.NewClient(OpenAIKey)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !chatAllowed(update.Message.Chat.ID) {
			continue
		}

		for _, newMember := range update.Message.NewChatMembers {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			welcomeMsg, err := generateWelcome(ctx, client, name(newMember))
			if err != nil {
				log.Printf("Error generating welcome message: %v", err)
				continue
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMsg))
		}
	}
}

func chatAllowed(chatID int64) bool {
	if _, found := allowlist[chatID]; found {
		return true
	}
	return false
}

func name(user tgbotapi.User) string {
	if user.UserName != "" {
		return user.UserName
	}
	return user.FirstName
}

func generateWelcome(ctx context.Context, c *openai.Client, name string) (string, error) {
	resp, err := c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			N:     1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are the greeter in a chat group called \"pisa.dev\". The messages are the names or usernames of new members joining the group. Your response will be a message in italian to welcome them to pisa.dev in a friendly way, followed by a question to encourage them to not be shy and present themselves. Sometimes you can ask questions or tell jokes about software engineer, hardware, or tech in general.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: name,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
