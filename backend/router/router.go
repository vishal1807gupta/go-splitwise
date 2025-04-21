package router

import (
	"go-splitwise/controller"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/register", controller.RegisterUser).Methods("POST")
	r.HandleFunc("/api/login", controller.LoginUser).Methods("POST")
	r.HandleFunc("/api/groupdetails/{userId}", controller.GetGroupDetailsByUserId).Methods("GET")
	r.HandleFunc("/api/creategroup/{userId}", controller.CreateGroup).Methods("POST")
	r.HandleFunc("/api/addUsersToGroup/{groupId}", controller.AddUsersToGroup).Methods("POST")
	r.HandleFunc("/api/groupUsers/{groupId}", controller.GetGroupUsers).Methods("GET")
	r.HandleFunc("/api/notGroupUsers/{groupId}", controller.GetNotGroupUsers).Methods("GET")
	r.HandleFunc("/api/addExpense/{groupId}", controller.AddExpense).Methods("POST")
	r.HandleFunc("/api/items/{groupId}", controller.GetItemsByGroupId).Methods("GET")
	r.HandleFunc("/api/settlements/{groupId}/{userId}", controller.GetSettlements).Methods("POST")
	r.HandleFunc("/api/auth/google", controller.HandleGoogleAuth).Methods("POST")
	r.HandleFunc("/api/me", controller.GetLoggedInUser).Methods("GET")
	r.HandleFunc("/api/logout", controller.Logout).Methods("POST")

	return r
}
