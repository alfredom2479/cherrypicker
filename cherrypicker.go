package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var apiAuthDomain = "https://www.reddit.com/api/v1/access_token?scope=*"
var apiDomain = "https://oauth.reddit.com/"

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
		redditUsername := os.Args[3]

		if redditUsername != "" && bearerToken != "" {
			getUserComments(client, redditUsername, bearerToken)
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

func getUserComments(c http.Client, un string, bt string) string {

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

	bodyString := string(body)

	var commentResponseData CommentResponseData
	json.Unmarshal([]byte(bodyString), &commentResponseData)

	commentCounter := 0

	for _, commentObject := range commentResponseData.Data.Children {
		commentCounter += 1
		fmt.Printf("comment #%d (r/%s): %s\n\n", commentCounter, commentObject.Data.Subreddit, commentObject.Data.Body)
	}

	afterId := commentResponseData.Data.After

	for afterId != "" {
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
			fmt.Printf("comment #%d (r/%s): %s\n\n", commentCounter, commentObject.Data.Subreddit, commentObject.Data.Body)
		}

	}

	return "ur not my dad"

}

func printUsageMessage() {
	fmt.Printf("Usage: go run cherrypicker token [clientID] [clientSecret] \nOR\ngo run cherrypicker analyze [bearerToken] [redditUsername]")
}
