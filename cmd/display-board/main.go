package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aveiga/display-board/pkg/db"
)

var messagesDb = db.Database{}

func main() {
	// Set up HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/messages", handleMessageCreationPage)
	http.HandleFunc("/message", handleMessageSubmission)

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	messages, err := messagesDb.GetMessages()
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Display Board</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
	<div id="wrapper">
    	<h1>Display Board</h1>
    	<table>
    	    <thead>
    	        <tr>
					<th>Username</th>
    	            <th>Message</th>
    	            <th>Date</th>
    	        </tr>
    	    </thead>
    	    <tbody>
    	        {{range .}}
    	        <tr>
    	            <td>{{.Username}}</td>
    	            <td>{{.Message}}</td>
    	            <td>{{if .Modified.IsZero}}{{.Created.Format "2006-01-02"}}{{else}}{{.Modified.Format "2006-01-02 15:04:05"}}{{end}}</td>
    	        </tr>
    	        {{end}}
    	    </tbody>
    	</table>
		<br/>
	</div>
</body>
</html>`

	t := template.Must(template.New("home").Parse(tmpl))
	t.Execute(w, messages)
}

func handleMessageCreationPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Display Board - Create Message</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
    <h1>Create New Message</h1>
    <form hx-post="/message" hx-swap="none" hx-on::after-request="this.reset()">
        <div>
            <label for="message">Message:</label><br>
            <textarea id="message" name="message" required></textarea><br>
        </div>
        <input type="submit" value="Submit">
    </form>
    <p><a href="/">Back to Messages</a></p>
</body>
</html>`

	t := template.Must(template.New("create").Parse(tmpl))
	t.Execute(w, nil)
}

func handleMessageSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	//username := r.FormValue("username")
	message := r.FormValue("message")

	if message == "" {
		return
	}

	_, err = messagesDb.AddMessage(db.Message{Message: message})

	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
