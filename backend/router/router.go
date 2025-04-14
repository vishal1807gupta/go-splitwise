package router

import (
	"go-splitwise/controller"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/userdetails", controller.GetUserDetails).Methods("GET")
	r.HandleFunc("/api/registeruser", controller.RegisterUser).Methods("POST")
	r.HandleFunc("/api/groupdetails", controller.GetGroupDetails).Methods("GET")
	r.HandleFunc("/api/groupdetails/{groupId}", controller.GetGroupDetailsByID).Methods("GET")
	r.HandleFunc("/api/creategroup", controller.CreateGroup).Methods("POST")

	return r
}
