package shared

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func httpHandler(sm *sync.Map) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			sm.Range(func(key, value interface{}) bool {
				fmt.Fprintf(w, "%s: %s\n", key, value)
				return true
			})
		case "POST":
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}
			key := r.Form.Get("key")
			value := r.Form.Get("value")
			sm.Store(key, value)
			fmt.Fprintf(w, "Added %s: %s\n", key, value)
		default:
			http.Error(w, "Unsupported HTTP method.", http.StatusBadRequest)
		}
	})
}

func startHttpServer(sm *sync.Map) {
	http.Handle("/", httpHandler(sm))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dhcpServer(sm *sync.Map) {
	for {
		// Simulated DHCP operation: just display all items every 10 seconds
		fmt.Println("DHCP Server access:")
		sm.Range(func(key, value interface{}) bool {
			fmt.Printf("%s: %s\n", key, value)
			return true
		})
		time.Sleep(10 * time.Second)
	}
}

func main() {
	sm := &sync.Map{}

	// Start the HTTP server in its own goroutine
	go startHttpServer(sm)

	// Start the simulated DHCP server in its own goroutine
	go dhcpServer(sm)

	// Block main goroutine indefinitely
	select {}
}
