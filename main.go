package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	goopenai "github.com/CasualCodersProjects/gopenai"
	"github.com/CasualCodersProjects/gopenai/types"
)

func main() {
	godotenv.Load()
	dg, err := discordgo.New(os.Getenv("DISCORD_KEY"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		res, err := AIResponse(m.Content)
		if err != nil {
			panic(err)
		}

		if ch, err := s.State.Channel(m.ChannelID); err != nil || !ch.IsThread() {
			thread, err := s.MessageThreadStartComplex(m.ChannelID, m.ID, &discordgo.ThreadStart{
				Name:                "Halo " + m.Author.Username + ", Berikut jawaban dari AI:",
				AutoArchiveDuration: 60,
				Invitable:           false,
				RateLimitPerUser:    10,
			})
			if err != nil {
				panic(err)
			}

			_, _ = s.ChannelMessageSend(thread.ID, res)
			m.ChannelID = thread.ID
		} else {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, res, m.Reference())
		}
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func AIResponse(question string) (response string, err error) {
	godotenv.Load(".env")
	openAI := goopenai.NewOpenAI(&goopenai.OpenAIOpts{
		APIKey: os.Getenv("AI_KEY"),
	})

	request := types.NewDefaultCompletionRequest("The following is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly.\n\nHuman: Hello, who are you?\nAI: I am an AI created by OpenAI. How can I help you today?\nHuman: " + question + "\nAI:")
	request.Model = "text-davinci-003"
	request.Temperature = 0.9
	request.MaxTokens = 150
	request.TopP = 1
	request.FrequencyPenalty = 0
	request.PresencePenalty = 0.6
	request.Stop = []string{" Human:", " AI:"}

	resp, err := openAI.CreateCompletion(request)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("Response not Found!")
	}

	return resp.Choices[0].Text, nil
}

// Reference Discord Go: https://github.com/bwmarrin/discordgo
