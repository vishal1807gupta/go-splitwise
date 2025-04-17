package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	model "go-splitwise/model"
	"net/http"
	"regexp"
	"strconv"
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

func insertUserIntoGroup(groupID int, userID int64) error {
	query := `INSERT INTO group_users (group_id, user_id)
	          VALUES ($1, $2)
	          ON CONFLICT DO NOTHING`
	_, err := db.Exec(query, groupID, userID)
	return err
}

func updateBalance(itemID int64, receiverID int64, amount int64) error {
	query := `INSERT INTO item_splits (item_id, user_id, share)
              VALUES ($1, $2, $3)
              ON CONFLICT (item_id, user_id)
              DO UPDATE SET share = EXCLUDED.share;`
	_, err := db.Exec(query, itemID, receiverID, amount)
	return err
}

func splitEqually(expense *model.Expense) error {
	shareAmount := expense.Amount / int64(len(expense.Shares))
	var newShares []model.UserShare
	err := updateBalance(expense.ExpenseID, expense.PayerID, expense.Amount)
	if err != nil {
		return err
	}
	newShares = append(newShares, model.UserShare{UserID: expense.PayerID, ShareAmount: expense.Amount})
	for _, share := range expense.Shares {
		userID := share.UserID
		if userID != expense.PayerID {
			err = updateBalance(expense.ExpenseID, userID, -shareAmount)
			if err != nil {
				return err
			}
			newShares = append(newShares, model.UserShare{UserID: userID, ShareAmount: -shareAmount})
		}
	}
	expense.Shares = newShares
	return nil
}

func splitExactAmount(expense *model.Expense) error {
	var sum int64 = 0
	for _, share := range expense.Shares {
		fmt.Println(share)
		sum += int64(share.ShareAmount)
	}
	// fmt.Println(sum, expense.Amount)
	if int64(sum) != expense.Amount {
		return fmt.Errorf("sum of shares is not equal to the amount")
	}
	var newShares []model.UserShare
	err := updateBalance(expense.ExpenseID, expense.PayerID, expense.Amount)
	if err != nil {
		return err
	}
	newShares = append(newShares, model.UserShare{UserID: expense.PayerID, ShareAmount: expense.Amount})
	for i, share := range expense.Shares {
		userID := share.UserID
		if userID != expense.PayerID {
			err = updateBalance(expense.ExpenseID, userID, -int64(expense.Shares[i].ShareAmount))
			if err != nil {
				return err
			}
			newShares = append(newShares, model.UserShare{UserID: userID, ShareAmount: -expense.Shares[i].ShareAmount})
		}
	}
	expense.Shares = newShares
	return nil
}

func splitByPercentage(expense *model.Expense) error {
	var percentSum int64 = 0
	for _, share := range expense.Shares {
		percentSum += share.ShareAmount
	}
	if percentSum != 100 {
		return fmt.Errorf("sum of shares is not equal to 100")
	}
	var newShares []model.UserShare
	err := updateBalance(expense.ExpenseID, expense.PayerID, expense.Amount)
	if err != nil {
		return err
	}
	newShares = append(newShares, model.UserShare{UserID: expense.PayerID, ShareAmount: expense.Amount})
	for i, share := range expense.Shares {
		userID := share.UserID
		if userID != expense.PayerID {
			shareAmount := int64(expense.Shares[i].ShareAmount) * expense.Amount / 100
			err = updateBalance(expense.ExpenseID, userID, -shareAmount)
			if err != nil {
				return err
			}
			newShares = append(newShares, model.UserShare{UserID: userID, ShareAmount: -shareAmount})
		}
	}
	expense.Shares = newShares
	return nil
}

func calculateBalances(expense *model.Expense) error {
	// Calculate balances based on the expense type and update the balances table
	switch expense.ExpenseType {
	case "EQUAL":
		return splitEqually(expense)
	case "EXACT":
		return splitExactAmount(expense)
	case "PERCENTAGE":
		return splitByPercentage(expense)
	}
	return nil
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

func AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	var groupUsers model.GroupUsers

	vars := mux.Vars(r)
	groupIDStr := vars["groupId"]

	groupID, err := strconv.Atoi(groupIDStr)

	if err != nil {
		http.Error(w, "Invalid group_id", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&groupUsers); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	for _, userID := range groupUsers.UserIDs {
		err := insertUserIntoGroup(groupID, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error adding user %d to group", userID), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Users added to group successfully",
		"group_id": groupID,
		"user_ids": groupUsers.UserIDs,
	})
}

func GetGroupUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupId"]

	rows, err := db.Query("SELECT u.user_id, u.name, u.email FROM users u JOIN group_users gu ON u.user_id = gu.user_id WHERE gu.group_id = $1", groupID)
	if err != nil {
		http.Error(w, "Failed to fetch group users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var u model.UserResponse
		err := rows.Scan(&u.UserID, &u.Name, &u.Email)
		if err != nil {
			http.Error(w, "Failed to scan group users", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetNotGroupUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupId"]

	rows, err := db.Query(`SELECT u.user_id, u.name, u.email
	FROM users u
	WHERE u.user_id NOT IN (
		SELECT user_id
		FROM group_users
		WHERE group_id = $1
	)`, groupID)
	if err != nil {
		http.Error(w, "Failed to fetch group users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var u model.UserResponse
		err := rows.Scan(&u.UserID, &u.Name, &u.Email)
		if err != nil {
			http.Error(w, "Failed to scan group users", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func AddExpense(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Hello")
	var expense model.Expense
	err := json.NewDecoder(r.Body).Decode(&expense)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// fmt.Println(err)
	vars := mux.Vars(r)
	groupID := vars["groupId"]

	query := `INSERT INTO items (group_id, amount, paid_by, description) VALUES ($1, $2, $3, $4) RETURNING item_id`
	err = db.QueryRow(query, groupID, expense.Amount, expense.PayerID, expense.Description).Scan(&expense.ExpenseID)
	if err != nil {
		http.Error(w, "Error inserting expense: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Println(err)
	err = calculateBalances(&expense)
	// fmt.Println(err)
	if err != nil {
		query = `DELETE FROM items WHERE item_id = $1`
		_, newErr := db.Exec(query, expense.ExpenseID)
		if newErr != nil {
			http.Error(w, "Error inserting expense: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, "Error calculating balances: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Println(err)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}
