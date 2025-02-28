package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/aveiga/display-board/pkg/db"
)

var messagesDb = db.Database{}
var currentScrollIndex = 0

func main() {
	// Start HTTP server for message creation
	go startWebServer()

	// Run framebuffer-based display
	go runConsoleBased()

	// Keep the main thread alive
	select {}
}

func runConsoleBased() {
	// Clear the console
	clearCommand := exec.Command("clear")
	clearCommand.Stdout = os.Stdout

	for {
		messages, _ := messagesDb.GetMessages()

		// Clear the screen
		clearCommand.Run()

		// Print title
		fmt.Println("\033[1;36m======== REMINDINTOSH ========\033[0m")
		fmt.Println()

		// Display messages
		maxShow := 10 // Maximum messages to show
		if len(messages) > 0 {
			count := len(messages)
			if count > maxShow {
				count = maxShow
			}

			for i := 0; i < count; i++ {
				idx := (currentScrollIndex + i) % len(messages)
				if idx < len(messages) {
					msg := messages[idx].Message
					if len(msg) > 40 {
						msg = msg[:37] + "..."
					}
					fmt.Printf("\033[1;32m%d:\033[0m %s\n", idx+1, msg)
				}
			}
		} else {
			fmt.Println("\033[1;33mNo messages yet\033[0m")
		}

		// Print footer
		fmt.Println()
		fmt.Println("\033[1;36m==============================\033[0m")

		// Increment scroll position every 3 seconds
		if len(messages) > maxShow {
			currentScrollIndex = (currentScrollIndex + 1) % len(messages)
		}

		// Wait before refresh
		time.Sleep(3 * time.Second)
	}
}

func startWebServer() {
	// Set up HTTP routes
	http.HandleFunc("/messages", handleMessageCreationPage)
	http.HandleFunc("/messages/data", handleMessageData)
	http.HandleFunc("/message", handleMessageAction)

	fmt.Println("Web server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}

func handleMessageCreationPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html class="fullPage">
<head>
    <title>Remindintosh - Create Message</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=VT323&display=swap" rel="stylesheet">
    <style>
        .fullPage {
            background-color: #000;
            color: #00B1B7;
            font-size: 40px;
            font-family: "VT323", serif;
            font-weight: 400;
            font-style: normal;
        }
        .button-terminal {
            background-color: black;
            color: #00B1B7;
            font-size: 40px;
            border: 5px solid #00B1B7;
            padding: 10px 20px;
            text-transform: uppercase;
            cursor: pointer;
            outline: none;
            transition: all 0.2s ease-in-out;
        }
        textarea {
            background-color: #000;
            color: #00B1B7;
            border: 5px solid #00B1B7;
            font-family: "VT323", serif;
            font-size: 30px;
            padding: 10px;
            width: 80%;
            height: 150px;
        }
        .message-list {
            border: 3px solid #00B1B7;
            margin-top: 20px;
            padding: 10px;
        }
        .message-item {
            margin-bottom: 10px;
            padding: 10px;
            border-bottom: 2px dashed #00B1B7;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .delete-btn {
            background-color: black;
            color: #ff0000;
            border: 3px solid #ff0000;
            font-family: "VT323", serif;
            font-size: 30px;
            cursor: pointer;
            padding: 5px 15px;
        }
        h2 {
            border-bottom: 3px solid #00B1B7;
            padding-bottom: 10px;
        }
        .header-with-button {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .refresh-btn {
            background-color: black;
            color: #00B1B7;
            border: 3px solid #00B1B7;
            font-family: "VT323", serif;
            font-size: 28px;
            cursor: pointer;
            padding: 5px 15px;
        }
    </style>
</head>
<body>
    <h1>REMINDINTOSH</h1>
    
    <h2>Create New Message</h2>
    <form hx-post="/message" hx-swap="none" hx-on::after-request="this.reset(); document.getElementById('refresh-button').click();">
        <div>
            <label for="message">Message:</label><br>
            <textarea id="message" name="message" required></textarea><br>
        </div>
        <input type="submit" value="Submit" class="button-terminal">
    </form>
    
    <div class="header-with-button">
        <h2>Current Messages</h2>
        <button id="refresh-button" class="refresh-btn" hx-get="/messages/data" hx-target="#message-list">REFRESH</button>
    </div>
    <div id="message-list" class="message-list" hx-get="/messages/data" hx-trigger="load">
        Loading messages...
    </div>
</body>
</html>`

	t := template.Must(template.New("create").Parse(tmpl))
	t.Execute(w, nil)
}

func handleMessageData(w http.ResponseWriter, r *http.Request) {
	messages, err := messagesDb.GetMessages()
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	if len(messages) == 0 {
		fmt.Fprint(w, "<div class='message-item'>No messages yet</div>")
		return
	}

	for _, msg := range messages {
		// Format the creation time
		timeStr := msg.Created.Format("2006-01-02 15:04")
		fmt.Fprintf(w, `
			<div class="message-item" id="msg-%d">
				<div>%s<br><small>%s</small></div>
				<button class="delete-btn" hx-delete="/message?id=%d" hx-target="#msg-%d" hx-swap="outerHTML">DELETE</button>
			</div>
		`, msg.ID, msg.Message, timeStr, msg.ID, msg.ID)
	}
}

func handleMessageAction(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		message := r.FormValue("message")
		if message == "" {
			return
		}

		_, err = messagesDb.AddMessage(db.Message{
			Message: message,
			Created: time.Now(),
			ID:      time.Now().Nanosecond(),
		})
		if err != nil {
			http.Error(w, "Failed to create message", http.StatusInternalServerError)
			return
		}

		// Return success status
		w.WriteHeader(http.StatusOK)
	} else if r.Method == http.MethodDelete {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Message ID is required", http.StatusBadRequest)
			return
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid message ID", http.StatusBadRequest)
			return
		}

		_, err = messagesDb.DeleteMessage(idInt)
		if err != nil {
			http.Error(w, "Failed to delete message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
