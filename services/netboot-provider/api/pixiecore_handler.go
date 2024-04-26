package api

import (
    "fmt"
    "net/http"
	"sync"
	"github.com/gorilla/mux"
)

// BootHandler handles requests on the /v1/boot/<mac-address> route
func MakeBootHandler(machines *sync.Map) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

		// Get the macaddr variable from the route
		vars := mux.Vars(r)
		macaddr := vars["macaddr"]

		// TODO retry if not found: maybe the DHCP server hasn't processed the request yet
		if archType, ok := machines.Load(macaddr); ok {
			fmt.Println("Value:", archType)
			fmt.Fprintf(w, "The machine with the MAC address: %s is of type: %s", macaddr, archType)
		}
		
		
	}
}
