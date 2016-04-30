package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Data struct {
	WebsiteUrl         string
	SessionId          string
	ResizeFrom         Dimension
	ResizeTo           Dimension
	CopyAndPaste       map[string]bool // map[fieldId]true
	FormCompletionTime int             // Seconds
}

type Dimension struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

type formEventRequest struct {
	EventType  string    `json:"eventType"`
	WebsiteUrl string    `json:"websiteUrl"`
	SessionId  string    `json:"sessionId"`
	Pasted     bool      `json:"pasted"`
	InputId    string    `json:"inputId"`
	Time       int       `json:"time"`
	ResizeFrom Dimension `json:"resizeFrom"`
	ResizeTo   Dimension `json:"resizeTo"`
}

// session ids to data
var formSessions = make(map[string]Data)

func main() {
	http.HandleFunc("/formEvent", formEventHandler)

	fmt.Println("Server now running on localhost:8080")
	fmt.Println(`Try running: curl -X POST -d '{"eventType":"copyAndPaste"}' http://localhost:8080/formEvent`)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func formEventHandler(w http.ResponseWriter, r *http.Request) {
	// Handle cors
	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unable to read body"))
		return
	}

	req := &formEventRequest{}

	if err = json.Unmarshal(body, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unable to unmarshal JSON request"))
		return
	}

	formSession, hasFormSession := formSessions[req.SessionId]

	// If no formsession has been created create one
	if !hasFormSession {
		formSession = Data{
			SessionId:          req.SessionId,
			WebsiteUrl:         req.WebsiteUrl,
			CopyAndPaste:       make(map[string]bool),
			ResizeFrom:         Dimension{},
			ResizeTo:           Dimension{},
			FormCompletionTime: 0,
		}
	}

	switch req.EventType {
	case "copyAndPaste":
		{
			formSession.CopyAndPaste[req.InputId] = req.Pasted
		}
	case "timeTaken":
		{
			formSession.FormCompletionTime = req.Time
		}
	case "resize":
		{
			formSession.ResizeFrom = req.ResizeFrom
			formSession.ResizeTo = req.ResizeTo
		}
	}

	// Save the session
	formSessions[req.SessionId] = formSession
	log.Printf("Form session %+v", formSession)

	w.WriteHeader(http.StatusOK)
}
