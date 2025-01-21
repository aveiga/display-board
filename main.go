package main

import (
	"fmt"
	"html/template"
	"net/http"

	"rsc.io/quote"
)

func main() {
	// Set up HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/quote", handleQuote)

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>HTMX Demo</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <h1>HTMX Demo</h1>
    <div hx-get="/quote" hx-trigger="click">
        Click me to get a quote!
    </div>
</body>
</html>`

	t := template.Must(template.New("home").Parse(tmpl))
	t.Execute(w, nil)
}

func handleQuote(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<div>%s</div>", quote.Go())
}
