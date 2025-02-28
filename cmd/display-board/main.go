package main

import (
	"fmt"
	"html/template"
	"image/color"
	"net/http"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/aveiga/display-board/pkg/db"
)

// RetroTheme implements a custom theme that mimics old terminal displays
type RetroTheme struct {
	fyne.Theme
}

func (r RetroTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.Black
	case theme.ColorNameForeground, theme.ColorNamePrimary:
		return color.RGBA{0x00, 0xB1, 0xB7, 0xFF} // #00B1B7 - the terminal green-blue color
	case theme.ColorNameButton, theme.ColorNameDisabled:
		return color.RGBA{0x00, 0x80, 0x80, 0xFF}
	default:
		return color.RGBA{0x00, 0xB1, 0xB7, 0xFF}
	}
}

func (r RetroTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (r RetroTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (r RetroTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 18 // Larger text size
	case theme.SizeNamePadding:
		return 4 // Reduced padding for compact look
	default:
		return theme.DefaultTheme().Size(name)
	}
}

var messagesDb = db.Database{}
var currentScrollIndex = 0

func main() {
	// Start HTTP server for message creation
	go startWebServer()

	// Start Fyne app for message display
	myApp := app.New()
	myApp.Settings().SetTheme(&RetroTheme{})
	window := myApp.NewWindow("Remindintosh")

	// Set up fullscreen mode
	window.SetFullScreen(true)
	window.CenterOnScreen()

	// Create title label with custom styling
	title := widget.NewLabel("Remindintosh")
	title.TextStyle = fyne.TextStyle{
		Bold:      true,
		Monospace: true,
	}
	// Center the title and wrap it in padding
	titleContainer := container.NewCenter(
		container.NewPadded(title),
	)

	// Define messageList here so it's in scope for the refresh function
	var messageList *widget.List
	messageList = widget.NewList(
		func() int {
			messages, _ := messagesDb.GetMessages()
			return len(messages)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				// Make the label take most of the space
				container.NewGridWrap(fyne.NewSize(240, 30),
					widget.NewLabel(""),
				),
				// Small delete button on the right
				container.NewGridWrap(fyne.NewSize(50, 30),
					widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			messages, _ := messagesDb.GetMessages()
			if id < len(messages) {
				box := item.(*fyne.Container)
				labelContainer := box.Objects[0].(*fyne.Container)
				buttonContainer := box.Objects[1].(*fyne.Container)

				label := labelContainer.Objects[0].(*widget.Label)
				button := buttonContainer.Objects[0].(*widget.Button)

				// Truncate long messages to fit the screen
				msg := messages[id].Message
				if len(msg) > 30 {
					msg = msg[:27] + "..."
				}
				label.SetText(msg)
				label.Wrapping = fyne.TextTruncate

				button.OnTapped = func() {
					messagesDb.DeleteMessage(messages[id].ID)
					messageList.Refresh()
				}
			}
		},
	)

	// Make the list items more compact
	messageList.CreateItem = func() fyne.CanvasObject {
		return container.NewHBox(
			container.NewGridWrap(fyne.NewSize(240, 30),
				widget.NewLabel("Template"),
			),
			container.NewGridWrap(fyne.NewSize(50, 30),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
			),
		)
	}

	// Layout
	content := container.NewBorder(
		titleContainer, // Top
		nil,            // Bottom
		nil,            // Left
		nil,            // Right
		messageList,    // Center
	)

	// Start periodic refresh (every 5 seconds)
	go func() {
		for range time.Tick(5 * time.Second) {
			messageList.Refresh()
		}
	}()

	// Auto-scroll functionality
	go func() {
		for range time.Tick(3 * time.Second) { // Adjust scroll speed here
			messages, _ := messagesDb.GetMessages()
			maxVisible := 6 // Approximate number of visible messages

			if len(messages) > maxVisible {
				currentScrollIndex = (currentScrollIndex + 1) % len(messages)
				messageList.ScrollTo(currentScrollIndex)
			}
		}
	}()

	window.SetContent(content)
	// Remove the Resize call since we're in fullscreen
	// window.Resize(fyne.NewSize(320, 240))
	// window.SetFixedSize(true) // Not needed in fullscreen
	window.ShowAndRun()
}

func startWebServer() {
	// Set up HTTP routes
	http.HandleFunc("/messages", handleMessageCreationPage)
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
    </style>
</head>
<body>
    <h1>Create New Message</h1>
    <form hx-post="/message" hx-swap="none" hx-on::after-request="this.reset()">
        <div>
            <label for="message">Message:</label><br>
            <textarea id="message" name="message" required></textarea><br>
        </div>
        <input type="submit" value="Submit" class="button-terminal">
    </form>
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
