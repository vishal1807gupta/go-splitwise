package router

import (
	"go-splitwise/controller"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()

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
	r.HandleFunc("/api/update-password", controller.UpdatePassword).Methods("POST")
	r.HandleFunc("/api/auth/request-password-reset", controller.RequestPasswordResetHandler).Methods("POST")
	r.HandleFunc("/api/auth/reset-password-complete", controller.ResetPasswordCompleteHandler).Methods("POST")
	r.HandleFunc("/api/memories/{groupId}", controller.GetMemoriesHandler).Methods("GET")
	r.HandleFunc("/api/memories/upload", controller.UploadMemoryHandler).Methods("POST")
	r.HandleFunc("/api/memories/{memoryId}", controller.DeleteMemoryHandler).Methods("DELETE")
	r.HandleFunc("/api/getTransactions/{groupId}", controller.GetTransactions).Methods("GET")
	r.HandleFunc("/api/insertTransactions/{groupId}", controller.InsertTransactions).Methods("POST")
	r.HandleFunc("/api/trigger-monthly-reminders", controller.TriggerMonthlyReminders).Methods("POST")
	r.HandleFunc("/api/wakeup", controller.WakeUp).Methods("POST")

	return r
}
