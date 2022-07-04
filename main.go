package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	// Load enviroment variables
	godotenv.Load(".env")
	// Build slack bot connection
	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))
	// Build wit ai client
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	// Build wolfram client
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	// A "go" statement starts the execution of a function call as an independent concurrent thread of control, or goroutine, within the same address space.
	go printCommandEvents(bot.CommandEvents())
	// Define bot command
	bot.Command("query for bot - <message>", &slacker.CommandDefinition{
		Description: "send any question to wolfram",
		Example:     "what is the fastes car on the planet",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")
			// Create wolfram query
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})

			// Format json and get value from WIT AI
			data, _ := json.MarshalIndent(msg, "", "     ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			// Get response from wolfram
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				fmt.Println("there is an error")
			}
			// Send response from wolfram to slack
			response.Reply(res)
		},
	})

	// Release recources
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for errors
	err := bot.Listen(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
