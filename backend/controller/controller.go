package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	model "go-splitwise/model"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Vishal#2002"
	dbname   = "splitwise"
)

var db *sql.DB

func init() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlconn)

	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")

	// defer db.Close()
}

func isValidEmail(email string) bool {
	regex := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	return regex.MatchString(email)
}

func validateUserInput(user model.UserRequest, w http.ResponseWriter) bool {
	if len(strings.TrimSpace(user.Name)) < 3 {
		http.Error(w, "Name must be at least 3 characters", http.StatusBadRequest)
		return false
	}
	if len(user.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return false
	}
	if !isValidEmail(user.Email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return false
	}
	return true
}

func fetchAllUsers() ([]model.UserResponse, error) {
	rows, err := db.Query("SELECT user_id, name, email FROM users")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var u model.UserResponse
		err := rows.Scan(&u.UserID, &u.Name, &u.Email)
		if err != nil {
			panic(err)
		}
		users = append(users, u)
	}
	return users, nil
}

func fetchAllGroups() ([]model.Group, error) {
	rows, err := db.Query("SELECT group_id, name FROM groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.Group
	for rows.Next() {
		var g model.Group
		if err := rows.Scan(&g.GroupID, &g.GroupName); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	users, err := fetchAllUsers()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user model.UserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if !validateUserInput(user, w) {
		return
	}

	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING user_id`
	err = db.QueryRow(query, user.Name, user.Email, user.Password).Scan(&user.UserID)
	if err != nil {
		http.Error(w, "Error inserting user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var registeredUser model.UserResponse
	registeredUser.UserID = user.UserID
	registeredUser.Name = user.Name
	registeredUser.Email = user.Email

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registeredUser)
}

func GetGroupDetails(w http.ResponseWriter, r *http.Request) {
	groups, err := fetchAllGroups()
	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group model.Group
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if len(strings.TrimSpace(group.GroupName)) < 3 {
		http.Error(w, "GroupName must be at least 3 characters", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO groups (name) VALUES ($1) RETURNING group_id`
	err = db.QueryRow(query, group.GroupName).Scan(&group.GroupID)
	if err != nil {
		http.Error(w, "Error inserting group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func GetGroupDetailsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupId"]

	var group model.Group
	err := db.QueryRow("SELECT group_id, name FROM groups WHERE group_id = $1", groupID).Scan(&group.GroupID, &group.GroupName)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}
