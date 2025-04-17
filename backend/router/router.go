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
	r.HandleFunc("/api/creategroup", controller.CreateGroup).Methods("POST")
	r.HandleFunc("/api/addUsersToGroup/{groupId}", controller.AddUsersToGroup).Methods("POST")
	r.HandleFunc("/api/groupUsers/{groupId}", controller.GetGroupUsers).Methods("GET")
	r.HandleFunc("/api/notGroupUsers/{groupId}", controller.GetNotGroupUsers).Methods("GET")
	r.HandleFunc("/api/addExpense/{groupId}", controller.AddExpense).Methods("POST")

	return r
}
