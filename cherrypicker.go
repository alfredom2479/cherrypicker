package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var apiAuthDomain = "https://www.reddit.com/api/v1/access_token?scope=*"
var apiDomain = "https://oauth.reddit.com"

type AuthTokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type CommentResponseData struct {
	Kind string `json:"kind"`
	Data struct {
		After    string `json:"after"`
		Dist     int    `json:"dist"`
		Children []struct {
			Kind string `json:"kind"`
			Data struct {
				SubredditID    string `json:"subreddit_id"`
				LinkTitle      string `json:"link_title"`
				Subreddit      string `json:"subreddit"`
				LinkAuthor     string `json:"link_author"`
				Author         string `json:"author"`
				ParentID       string `json:"parent_id"`
				AuthorFullname string `json:"parent_fullname"`
				Body           string `json:"body"`
				BodyHTML       string `json:"body_html"`
				LinkID         string `json:"link_id"`
				Permalink      string `json:"permalink"`
				LinkPermalink  string `json:"link_permalink"`
				Name           string `json:"name"`
				CreatedUTC     string `json:"created_utc"`
				LinkURL        string `json:"link_url"`
			} `json:"data"`
		} `json:"children"`
		Before string `json:"before"`
	} `json:"data"`
}

func main() {

	programMode := os.Args[1]
	client := http.Client{}

	if programMode == "token" {

		clientIdInput := os.Args[2]
		clientSecretInput := os.Args[3]

		if clientIdInput != "" && clientSecretInput != "" {
			getBearerToken(client, clientIdInput, clientSecretInput)
			return
		}
		printUsageMessage()
		return
	}

	if programMode == "analyze" {
		bearerToken := os.Args[2]
		openAiToken := os.Args[3]
		redditUsername := os.Args[4]

		if redditUsername != "" && bearerToken != "" && openAiToken != "" {

			openAIClient := openai.NewClient(openAiToken)
			getUserComments(client, redditUsername, bearerToken, *openAIClient)
			return
		}
		printUsageMessage()
		return
	}
	if programMode == "grep" {
		bearerToken := os.Args[2]
		redditUsername := os.Args[3]
		searchString := os.Args[4]

		if redditUsername != "" && bearerToken != "" {
			searchCoemments(client, redditUsername, bearerToken, searchString)
			return
		}
		printUsageMessage()
		return
	}

	fmt.Println("Invalid mode.")
	printUsageMessage()
	return
}

func getBearerToken(c http.Client, cid string, cs string) string {
	data := url.Values{}
	data.Add("grant_type", "client_credentials")

	req, _ := http.NewRequest("POST", apiAuthDomain, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cid, cs)

	res, err := c.Do(req)
	if err != nil {
		fmt.Println("No response from bearer token request")
	}
	body, err := io.ReadAll(res.Body)

	var result AuthTokenData
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("There has been an error?")

	}

	fmt.Println(string(body))
	return string(body)
}

func searchCoemments(c http.Client, un string, bt string, ss string) {
	ceb := apiDomain + "/user/" + un + "/comments"
	ce := ceb
	var commentResponseData CommentResponseData
	commentCounter := 0
	currComment := ""
	afterId := "init"
	var matchingCommentsSlice []string

	fmt.Println("search string: ", ss)

	for afterId != "" {
		if commentCounter > 1500 {
			break
		}
		req, err := http.NewRequest("GET", ce, nil)
		if err != nil {
			fmt.Printf("Error making new request: %v", err)
			return
		}
		req.Header.Add("Authorization", "bearer "+bt)
		req.Header.Add("User-Agent", "ChangeMeClient/0.1 by YourUsername")

		res, err := c.Do(req)
		if err != nil {
			fmt.Printf("Error doing request: %v", err)
			return
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Error reading all of body: %v", err)
		}

		commentResponseData.Data.After = ""
		json.Unmarshal([]byte(string(body)), &commentResponseData)
		afterId = commentResponseData.Data.After

		for _, commentObject := range commentResponseData.Data.Children {
			commentCounter += 1
			if strings.Contains(commentObject.Data.Body, ss) {

				currComment = fmt.Sprintf("#%d: '%s' [%s] on r/%s", commentCounter, commentObject.Data.Body, commentObject.Data.CreatedUTC, commentObject.Data.Subreddit)
				matchingCommentsSlice = append(matchingCommentsSlice, currComment)
				fmt.Printf("comment #%d (r/%s): %s\n\n", commentCounter, commentObject.Data.Subreddit, commentObject.Data.Body)
				//time.Parse(commentObject.Data.CreatedUTC)
			}
		}

		ce = ceb + "?after=" + commentResponseData.Data.After
	}

	fmt.Println("Total comments found: ", commentCounter)
}

func getUserComments(c http.Client, un string, bt string, oaic openai.Client) string {

	/*resp,err := oaic.RetrieveAssistant(context.Background(), "asst_hoFUtfCL5jDnFhtYNE7gxQYF")
	if err != nil {
		fmt.Printf("Error getting assistant: %v\n",err)
		return "ur not my dad";
	}
	oaic.CreateAssistant(context.Background(), openai.AssistantRequest{})
	*/

	ce := apiDomain + "/user/" + un + "/comments"

	fmt.Println(ce)

	req, _ := http.NewRequest("GET", ce, nil)
	req.Header.Add("Authorization", "bearer "+bt)
	req.Header.Add("User-Agent", "ChangeMeClient/0.1 by YourUsername")

	fmt.Println(req.UserAgent())

	res, err := c.Do(req)
	if err != nil {
		fmt.Println("No response from comments endpoint")
		return "your not my dad"
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading all of response body")
		return "your not my dad"
	}

	bodyString := string(body)
	fmt.Println(bodyString)

	var commentResponseData CommentResponseData
	json.Unmarshal([]byte(bodyString), &commentResponseData)

	var commentsSlice []string

	commentCounter := 0
	currComment := ""
	fmt.Println("right before the fors")
	for _, commentObject := range commentResponseData.Data.Children {
		commentCounter += 1
		currComment = fmt.Sprintf("#%d: '%s' on r/%s", commentCounter, commentObject.Data.Body, commentObject.Data.Subreddit)
		fmt.Printf("comment #%d (in r/%s): %s\n\n", commentCounter, commentObject.Data.Subreddit, commentObject.Data.Body)

		commentsSlice = append(commentsSlice, currComment)
	}

	afterId := commentResponseData.Data.After

	for afterId != "" {
		if commentCounter > 300 {
			break
		}
		req, _ = http.NewRequest("GET", ce+"?after="+commentResponseData.Data.After, nil)
		req.Header.Add("Authorization", "bearer "+bt)
		req.Header.Add("User-Agent", "ChangeMeClient/0.1 by YourUsername")
		res, err = c.Do(req)
		if err != nil {
			fmt.Println("Error getting more comments")
			fmt.Println(err)
			return "ur not my dad"
		}

		body, err = io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading more comments")
			fmt.Println(err)
			return "ur not my dad"
		}

		commentResponseData.Data.After = ""

		json.Unmarshal([]byte(string(body)), &commentResponseData)

		afterId = commentResponseData.Data.After

		for _, commentObject := range commentResponseData.Data.Children {
			commentCounter += 1
			currComment = fmt.Sprintf("#%d: '%s' on r/%s", commentCounter, commentObject.Data.Body, commentObject.Data.Subreddit)
			commentsSlice = append(commentsSlice, currComment)
			fmt.Printf("comment #%d (r/%s): %s\n\n", commentCounter, commentObject.Data.Subreddit, commentObject.Data.Body)
		}
	}
	fmt.Println("Total comments being analyzed:", commentCounter)
	//OPENAI STUFF ////////////////////

	/*
		assistantName := "Reddit Comment Analyzer"
		assistantInstructions := "You are an Investigator. I will give you a list of numbered reddit comments along with what subreddit they were made in and you will tell me any peronal information you can find, including name, location, hobbies, etc. analyze carefully and return a numbered list of what you found along with the comment numbers that helped you with each item on the list. I was given permission to do this."
		assistant, err := oaic.CreateAssistant(
			context.Background(),
			openai.AssistantRequest{
				Model:        openai.GPT4,
				Name:         &assistantName,
				Instructions: &assistantInstructions,
			},
		)
		if err != nil {
			fmt.Printf("There has been an error creating assistant: %v\n", err)
		}
	*/

	analyzeThread, err := oaic.CreateThread(context.Background(), openai.ThreadRequest{})
	if err != nil {
		fmt.Printf("error making new thread: %v\n", err)
		return "ur not my dad"
	}

	openAIMessage, err := oaic.CreateMessage(
		context.Background(),
		analyzeThread.ID,
		openai.MessageRequest{
			Role:    "user",
			Content: strings.Join(commentsSlice, ",")[:32700],
		},
	)
	if err != nil {
		fmt.Printf("There wasn an error creating message: %v", err)
		return "ur not my dad"
	}
	fmt.Println("Message", openAIMessage)

	modelName := "gpt-4"

	runresp, err := oaic.CreateRun(
		context.Background(),
		analyzeThread.ID,
		openai.RunRequest{
			AssistantID: "asst_7PwIlNpxvkbFVgGrcXU7PYNS",
			Model:       &modelName,
		},
	)

	if err != nil {
		fmt.Printf("There has been an error running: %v", err)
	}

	runresp, err = oaic.RetrieveRun(
		context.Background(),
		analyzeThread.ID,
		runresp.ID,
	)
	if err != nil {
		fmt.Printf("There has been an error retrieving run status: %v\n", err)
	}

	threadStatus := runresp.Status

	for threadStatus == "in_progress" || threadStatus == "queued" {
		runresp, err = oaic.RetrieveRun(
			context.Background(),
			analyzeThread.ID,
			runresp.ID,
		)
		if err != nil {
			fmt.Printf("There has been an error retrieving run status: %v\n", err)
		}
		threadStatus = runresp.Status
		fmt.Println("run status (in loop):", runresp.Status)
	}

	fmt.Println("run status:", runresp.Status)
	fmt.Println(runresp)

	/*
		var messageLim = 4
		beforeString := ""
		afterString := ""
		orderString := ""
	*/

	messages, err := oaic.ListMessage(
		context.Background(),
		analyzeThread.ID,
		nil,
		nil,
		nil,
		nil,
	)

	/*
		fmt.Println("Messages?: ", messages.Messages)
		fmt.Println("Message len?: ", len(messages.Messages))
		fmt.Println("message[0]", messages.Messages[0])
		fmt.Println("message[0] content: ", messages.Messages[0].Content)
	*/
	fmt.Println(messages.Messages[0].Content[0].Text.Value)
	fmt.Println(commentsSlice)
	//fmt.Println(messages)

	//fmt.Println(commentsSlice)

	return "ur not my dad"

}

func printUsageMessage() {
	fmt.Printf("Usage: go run cherrypicker token [clientID] [clientSecret] \nOR\ngo run cherrypicker analyze [bearerToken] [openaitoken] [redditUsername]\nORgo run cherrypicker grep [bearerToken] [redditusername]\n")
}
