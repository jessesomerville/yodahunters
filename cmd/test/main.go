package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/tjarratt/babble"
)

// Create a bunch of dummy data by sending requests to the API
func main() {
	registerURL := "http://localhost:8080/api/register"
	loginURL := "http://localhost:8080/api/login"
	postThreadURL := "http://localhost:8080/api/threads"
	postCommentURL := "http://localhost:8080/api/comments"

	var threadIDs []int

	babbler := babble.NewBabbler()
	babbler.Count = 3

	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println("Error creating cookie jar:", err)
		return
	}

	client := &http.Client{
		Jar: jar,
	}

	for i := 0; i < 100; i++ {
		username := babbler.Babble()
		password := babbler.Babble()
		email := username + "@example.com"

		// Register, login, post some threads, post some comments
		fmt.Println("Registering user:", username)
		registerBody := fmt.Sprintf(`{"username": "%s", "email": "%s", "password": "%s"}`, username, email, password)
		registerResp, err := client.Post(registerURL, "application/json", strings.NewReader(registerBody))
		if err != nil {
			fmt.Println("Error registering:", err)
			continue
		} else {
			fmt.Println("Registration successful")
		}
		defer registerResp.Body.Close()

		fmt.Println("Logging in user:", username)
		loginBody := fmt.Sprintf(`{"username": "%s", "email": "%s", "password": "%s"}`, username, email, password)
		loginResp, err := client.Post(loginURL, "application/json", strings.NewReader(loginBody))
		if err != nil {
			fmt.Println("Error registering:", err)
			continue
		} else {
			fmt.Println("Login successful")
		}
		defer loginResp.Body.Close()

		numThreads := rand.Intn(5) + 1
		fmt.Println("Posting", numThreads, "threads")
		for j := 0; j < numThreads; j++ {
			title := babbler.Babble()
			body := babbler.Babble()
			threadBody := fmt.Sprintf(`{"title": "%s", "body": "%s"}`, title, body)
			threadResp, err := client.Post(postThreadURL, "application/json", strings.NewReader(threadBody))
			if err != nil {
				fmt.Println("Error posting thread:", err)
				continue
			}
			defer threadResp.Body.Close()

			var thread struct {
				ThreadID int `json:"thread_id"`
			}
			if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
				fmt.Println("Error decoding thread response:", err)
				continue
			}
			threadIDs = append(threadIDs, thread.ThreadID)
		}

		numComments := rand.Intn(10) + 1
		fmt.Println("Posting", numComments, "comments")
		for j := 0; j < numComments; j++ {
			threadID := threadIDs[rand.Intn(len(threadIDs))]
			body := babbler.Babble()
			commentBody := fmt.Sprintf(`{"thread_id": %d, "body": "%s"}`, threadID, body)
			commentResp, err := client.Post(postCommentURL, "application/json", strings.NewReader(commentBody))
			if err != nil {
				fmt.Println("Error posting comment:", err)
				continue
			}
			defer commentResp.Body.Close()
		}
	}

}
