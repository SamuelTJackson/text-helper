package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type geminiClient struct {
	model  *genai.GenerativeModel
	client *genai.Client
}

func newClient(ctx context.Context) (*geminiClient, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil, errors.New("you have to set the env variable 'GEMINI_API_KEY' to make this app work")
	}

	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return nil, fmt.Errorf("could not create genai client: %w", err)
	}

	itr := genaiClient.ListModels(ctx)
	for {
		m, err := itr.Next()
		if err != nil {
			break
		}

		fmt.Println(m.Name)

	}
	model := genaiClient.GenerativeModel("gemini-2.0-flash")

	return &geminiClient{
		model:  model,
		client: genaiClient,
	}, nil

}

func (c *geminiClient) SendMessage(ctx context.Context, text string) <-chan string {
	respChannel := make(chan string)

	if text == "" {
		close(respChannel)

		return respChannel
	}

	go func() {
		defer close(respChannel)

		cs := c.model.StartChat()

		cs.History = []*genai.Content{
			{
				Parts: []genai.Part{
					genai.Text(
						`In the following, I will send you some text in either English or German.
						Your task is to improve the wording, grammar, and overall quality.
						You are part of a program, and you should only respond with the corrected version, no other responses.
						If the text contains colloquial language or informal expressions, keep them as they are.
						The style of the text should not change drastically.`,
					),
				},
				Role: "user",
			},
		}

		iter := cs.SendMessageStream(ctx, genai.Text(text))
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				log.Fatal(err)

				return
			}

			for _, cand := range resp.Candidates {
				if cand.Content == nil {
					continue
				}

				for _, part := range cand.Content.Parts {
					respChannel <- fmt.Sprint(part)
				}
			}
		}
	}()

	return respChannel
}
