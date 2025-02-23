package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/aveiga/display-board/pkg/db"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Be careful with this in production
	},
}

// Hub maintains the set of active clients
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

var (
	messagesDb = db.Database{}
	hub        = newHub()
)

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

func main() {
	// Start the hub
	go hub.run()

	// Set up HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/messages", handleMessageCreationPage)
	http.HandleFunc("/message", handleMessageAction)
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	// Register new client
	hub.register <- conn

	// Send initial messages
	messages, _ := messagesDb.GetMessages()
	html := generateMessagesHTML(messages)
	wsMessage := fmt.Sprintf(html)
	conn.WriteMessage(websocket.TextMessage, []byte(wsMessage))
}

// Helper function to generate messages HTML
func generateMessagesHTML(messages []db.Message) string {
	var html string
	for _, msg := range messages {
		html += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%s</td>
				<td>
					<button hx-delete="/message?id=%d"
							hx-confirm="Are you sure you want to delete this message?"
							hx-target="closest tr"
							hx-swap="outerHTML">
						Delete
					</button>
				</td>
			</tr>`,
			msg.Message,
			func() string {
				if !msg.Modified.IsZero() {
					return msg.Modified.Format("2006-01-02 15:04:05")
				}
				return msg.Created.Format("2006-01-02")
			}(),
			msg.ID)
	}
	return html
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	messages, err := messagesDb.GetMessages()
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	tmpl := `<!DOCTYPE html>
<html class="fullPage">
<head>
    <title>Display Board</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <script src="https://unpkg.com/htmx.org/dist/ext/ws.js"></script>
	<style>
		@keyframes flicker {
		  0% {
		    opacity: 0.27861;
		  }
		  5% {
		    opacity: 0.34769;
		  }
		  10% {
		    opacity: 0.23604;
		  }
		  15% {
		    opacity: 0.90626;
		  }
		  20% {
		    opacity: 0.18128;
		  }
		  25% {
		    opacity: 0.83891;
		  }
		  30% {
		    opacity: 0.65583;
		  }
		  35% {
		    opacity: 0.67807;
		  }
		  40% {
		    opacity: 0.26559;
		  }
		  45% {
		    opacity: 0.84693;
		  }
		  50% {
		    opacity: 0.96019;
		  }
		  55% {
		    opacity: 0.08594;
		  }
		  60% {
		    opacity: 0.20313;
		  }
		  65% {
		    opacity: 0.71988;
		  }
		  70% {
		    opacity: 0.53455;
		  }
		  75% {
		    opacity: 0.37288;
		  }
		  80% {
		    opacity: 0.71428;
		  }
		  85% {
		    opacity: 0.70419;
		  }
		  90% {
		    opacity: 0.7003;
		  }
		  95% {
		    opacity: 0.36108;
		  }
		  100% {
		    opacity: 0.24387;
		  }
		}
		@keyframes textShadow {
		  0% {
		    text-shadow: 0.4389924193300864px 0 1px rgba(0,30,255,0.5), -0.4389924193300864px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  5% {
		    text-shadow: 2.7928974010788217px 0 1px rgba(0,30,255,0.5), -2.7928974010788217px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  10% {
		    text-shadow: 0.02956275843481219px 0 1px rgba(0,30,255,0.5), -0.02956275843481219px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  15% {
		    text-shadow: 0.40218538552878136px 0 1px rgba(0,30,255,0.5), -0.40218538552878136px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  20% {
		    text-shadow: 3.4794037899852017px 0 1px rgba(0,30,255,0.5), -3.4794037899852017px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  25% {
		    text-shadow: 1.6125630401149584px 0 1px rgba(0,30,255,0.5), -1.6125630401149584px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  30% {
		    text-shadow: 0.7015590085143956px 0 1px rgba(0,30,255,0.5), -0.7015590085143956px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  35% {
		    text-shadow: 3.896914047650351px 0 1px rgba(0,30,255,0.5), -3.896914047650351px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  40% {
		    text-shadow: 3.870905614848819px 0 1px rgba(0,30,255,0.5), -3.870905614848819px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  45% {
		    text-shadow: 2.231056963361899px 0 1px rgba(0,30,255,0.5), -2.231056963361899px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  50% {
		    text-shadow: 0.08084290417898504px 0 1px rgba(0,30,255,0.5), -0.08084290417898504px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  55% {
		    text-shadow: 2.3758461067427543px 0 1px rgba(0,30,255,0.5), -2.3758461067427543px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  60% {
		    text-shadow: 2.202193051050636px 0 1px rgba(0,30,255,0.5), -2.202193051050636px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  65% {
		    text-shadow: 2.8638780614874975px 0 1px rgba(0,30,255,0.5), -2.8638780614874975px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  70% {
		    text-shadow: 0.48874025155497314px 0 1px rgba(0,30,255,0.5), -0.48874025155497314px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  75% {
		    text-shadow: 1.8948491305757957px 0 1px rgba(0,30,255,0.5), -1.8948491305757957px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  80% {
		    text-shadow: 0.0833037308038857px 0 1px rgba(0,30,255,0.5), -0.0833037308038857px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  85% {
		    text-shadow: 0.09769827255241735px 0 1px rgba(0,30,255,0.5), -0.09769827255241735px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  90% {
		    text-shadow: 3.443339761481782px 0 1px rgba(0,30,255,0.5), -3.443339761481782px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  95% {
		    text-shadow: 2.1841838852799786px 0 1px rgba(0,30,255,0.5), -2.1841838852799786px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		  100% {
		    text-shadow: 2.6208764473832513px 0 1px rgba(0,30,255,0.5), -2.6208764473832513px 0 1px rgba(255,0,80,0.3), 0 0 3px;
		  }
		}
		.crt::after {
		  content: " ";
		  display: block;
		  position: absolute;
		  top: 0;
		  left: 0;
		  bottom: 0;
		  right: 0;
		  background: rgba(18, 16, 16, 0.1);
		  opacity: 0;
		  z-index: 2;
		  pointer-events: none;
		  animation: flicker 0.15s infinite;
		}
		.crt::before {
		  content: " ";
		  display: block;
		  position: absolute;
		  top: 0;
		  left: 0;
		  bottom: 0;
		  right: 0;
		  background: linear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 0, 0, 0.25) 50%), linear-gradient(90deg, rgba(255, 0, 0, 0.06), rgba(0, 255, 0, 0.02), rgba(0, 0, 255, 0.06));
		  z-index: 2;
		  background-size: 100% 2px, 3px 100%;
		  pointer-events: none;
		}
		.crt {
		  animation: textShadow 1.6s infinite;
		}

		.fullPage {
			background-color: #000;
			color: #00B1B7;
			font-size: 50px;
		}

		.h1 {
			text-align: center;
		}

		.table {
			width: 100%;
		}

		.flexContainer {
			display: flex;
			flex-direction: column;
			gap: 10px;
		}

		.flexItem {
			display: flex;
			flex-direction: row;
			gap: 10px;
		}

		.flexItemContent {
			flex: 1;
		}

		
		
		
	</style>
</head>
<body>
    <div id="wrapper" class="crt">
        <h1 class="h1">Display Board</h1>
        <div hx-ext="ws" ws-connect="/ws">
			<div class="flexContainer">
				{{range .}}
				<div class="flexItem">
					<div class="flexItemContent message">
						{{.Message}}
					</div>
					<div class="flexItemContent date">
						{{if .Modified.IsZero}}{{.Created.Format "2006-01-02"}}{{else}}{{.Modified.Format "2006-01-02 15:04:05"}}{{end}}
					</div>
					<div class="flexItemContent actions">
						<button hx-delete="/message?id={{.ID}}"
							hx-confirm="Are you sure you want to delete this message?"
							hx-target="closest tr"
							hx-swap="outerHTML">
							Delete
						</button>
					</div>
				</div>
				{{end}}
			</div>

            <table class="table">
                <thead>
                    <tr>
                        <th>Message</th>
                        <th>Date</th>
                        <th></th>
                    </tr>
                </thead>
                <tbody id="messages-list">
                    {{range .}}
                    <tr>
                        <td>{{.Message}}</td>
                        <td>{{if .Modified.IsZero}}{{.Created.Format "2006-01-02"}}{{else}}{{.Modified.Format "2006-01-02 15:04:05"}}{{end}}</td>
                        <td>
                            <button hx-delete="/message?id={{.ID}}"
                                    hx-confirm="Are you sure you want to delete this message?"
                                    hx-target="closest tr"
                                    hx-swap="outerHTML">
                                Delete
                            </button>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
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

		_, err = messagesDb.AddMessage(db.Message{Message: message})
		if err != nil {
			http.Error(w, "Failed to create message", http.StatusInternalServerError)
			return
		}

		// Broadcast updated messages
		messages, _ := messagesDb.GetMessages()
		html := generateMessagesHTML(messages)
		wsMessage := fmt.Sprintf(`<div hx-swap-oob="innerHTML:#messages-list">%s</div>`, html)
		hub.broadcast <- []byte(wsMessage)

		http.Redirect(w, r, "/", http.StatusSeeOther)
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

		// Broadcast updated messages
		messages, _ := messagesDb.GetMessages()
		html := generateMessagesHTML(messages)
		wsMessage := fmt.Sprintf(`<div hx-swap-oob="innerHTML:#messages-list">%s</div>`, html)
		hub.broadcast <- []byte(wsMessage)

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
