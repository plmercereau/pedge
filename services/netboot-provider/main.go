package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/mux"

    "netboot-provider/api"
)

func main() {
    machines := &sync.Map{}

    // Initialize HTTP server
    go func() {
        r := mux.NewRouter()
        r.HandleFunc("/v1/boot/{macaddr}", api.MakeBootHandler(machines))
        port := "3000"
        log.Printf("HTTP server starting on port %s\n", port)
        if err := http.ListenAndServe(":"+port, r); err != nil {
            log.Fatal("HTTP server failed: ", err)
        }
    }()

    // Initialize DHCP server
    go api.StartDHCPServer(machines) 

    // Block forever
    select {}
    
}
