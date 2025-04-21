package controller

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	model "go-splitwise/model"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"google.golang.org/api/idtoken"
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

func generateSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
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

func fetchAllGroupsByUserID(userID int64) ([]model.Group, error) {
	rows, err := db.Query("SELECT group_id, name FROM groups WHERE group_id IN (SELECT group_id FROM group_users WHERE user_id = $1)", userID)
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
	var payerShare int64 = expense.Amount
	for _, share := range expense.Shares {
		userID := share.UserID
		if userID != expense.PayerID {
			err := updateBalance(expense.ExpenseID, userID, -shareAmount)
			if err != nil {
				return err
			}
			newShares = append(newShares, model.UserShare{UserID: userID, ShareAmount: -shareAmount})
		} else {
			payerShare -= shareAmount
		}
	}
	err := updateBalance(expense.ExpenseID, expense.PayerID, payerShare)
	if err != nil {
		return err
	}
	newShares = append(newShares, model.UserShare{UserID: expense.PayerID, ShareAmount: payerShare})
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
	var payerShare int64 = expense.Amount
	for i, share := range expense.Shares {
		userID := share.UserID
		if userID != expense.PayerID {
			err := updateBalance(expense.ExpenseID, userID, -int64(expense.Shares[i].ShareAmount))
			if err != nil {
				return err
			}
			newShares = append(newShares, model.UserShare{UserID: userID, ShareAmount: -expense.Shares[i].ShareAmount})
		} else {
			payerShare -= int64(expense.Shares[i].ShareAmount)
		}
	}
	err := updateBalance(expense.ExpenseID, expense.PayerID, payerShare)
	if err != nil {
		return err
	}
	newShares = append(newShares, model.UserShare{UserID: expense.PayerID, ShareAmount: payerShare})
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
	var payerShare int64 = expense.Amount
	for i, share := range expense.Shares {
		userID := share.UserID
		if userID != expense.PayerID {
			shareAmount := int64(expense.Shares[i].ShareAmount) * expense.Amount / 100
			err := updateBalance(expense.ExpenseID, userID, -shareAmount)
			if err != nil {
				return err
			}
			newShares = append(newShares, model.UserShare{UserID: userID, ShareAmount: -shareAmount})
		} else {
			shareAmount := int64(expense.Shares[i].ShareAmount) * expense.Amount / 100
			payerShare -= shareAmount
		}
	}
	err := updateBalance(expense.ExpenseID, expense.PayerID, payerShare)
	if err != nil {
		return err
	}
	newShares = append(newShares, model.UserShare{UserID: expense.PayerID, ShareAmount: payerShare})
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

	sessionToken := generateSessionToken()

	// Store the session in the database with expiration time
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3)",
		sessionToken, registeredUser.UserID, expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registeredUser)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user model.UserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var registeredUser model.UserResponse
	err = db.QueryRow("SELECT user_id, name, email FROM users WHERE email = $1 AND password = $2", user.Email, user.Password).Scan(&registeredUser.UserID, &registeredUser.Name, &registeredUser.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	sessionToken := generateSessionToken()

	// Store the session in the database with expiration time
	var expiresAt time.Time
	if user.RememberMe {
		expiresAt = time.Now().Add(7 * 24 * time.Hour) // 1 week
	} else {
		expiresAt = time.Now().Add(24 * time.Hour) // 1 day
	}
	_, err = db.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3)",
		sessionToken, registeredUser.UserID, expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registeredUser)
}

func HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {

	var body struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	payload, err := idtoken.Validate(r.Context(), body.Token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		http.Error(w, "Invalid Google token", http.StatusUnauthorized)
		return
	}

	claims := payload.Claims

	email := claims["email"].(string)
	name := claims["name"].(string)
	googleID := claims["sub"].(string)

	// ⬇️ Check if user exists in DB, else insert
	fmt.Println("Google ID:", googleID)
	var user model.UserResponse
	err = db.QueryRow("SELECT user_id, name FROM users WHERE google_id = $1", googleID).Scan(&user.UserID, &user.Name)
	if err == sql.ErrNoRows {
		fmt.Println("User not found, creating new user")

		err = db.QueryRow("SELECT user_id, name FROM users WHERE email = $1", email).Scan(&user.UserID, &user.Name)
		if err == sql.ErrNoRows {
			err = db.QueryRow("INSERT INTO users (name, email, google_id) VALUES ($1, $2, $3) RETURNING user_id, name", name, email, googleID).Scan(&user.UserID, &user.Name)
			if err != nil {
				http.Error(w, "Error inserting user: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, "Error fetching user: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			_, err = db.Exec("UPDATE users SET google_id = $1 WHERE user_id = $2", googleID, user.UserID)
			if err != nil {
				http.Error(w, "Error updating user: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}

	sessionToken := generateSessionToken()

	// Store the session in the database with expiration time
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3)",
		sessionToken, user.UserID, expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})

	user.Email = email

	json.NewEncoder(w).Encode(user)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete the session from the database
		_, _ = db.Exec("DELETE FROM sessions WHERE session_id = $1", cookie.Value)
	}

	// Remove the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

func GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	// Get the session token cookie
	cookie, err := r.Cookie("session_token")

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not authenticated",
		})
		return
	}

	sessionToken := cookie.Value

	// Look up the session in the database
	var userID int64
	var expiresAt time.Time
	err = db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE session_id = $1",
		sessionToken,
	).Scan(&userID, &expiresAt)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Invalid session",
		})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Server error",
		})
		return
	}

	// Check if session has expired
	if time.Now().After(expiresAt) {
		// Delete the expired session
		_, _ = db.Exec("DELETE FROM sessions WHERE session_id = $1", sessionToken)

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Session expired",
		})
		return
	}

	// Get user data from database
	var user model.UserResponse
	err = db.QueryRow("SELECT user_id, name, email FROM users WHERE user_id = $1", userID).Scan(
		&user.UserID, &user.Name, &user.Email,
	)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User not found",
		})
		return
	}

	// User is authenticated, return user data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func CleanupExpiredSessions() {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
	}
}

func GetGroupDetailsByUserId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		http.Error(w, "Invalid group_id", http.StatusBadRequest)
		return
	}

	groups, err := fetchAllGroupsByUserID(int64(userID))

	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	var group model.Group
	err = json.NewDecoder(r.Body).Decode(&group)
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
	err = insertUserIntoGroup(int(group.GroupID), int64(userID))
	if err != nil {
		http.Error(w, "Error inserting user into group: "+err.Error(), http.StatusInternalServerError)
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

func GetItemsByGroupId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupId"]

	rows, err := db.Query("SELECT item_id, amount, paid_by, description, created_at FROM items WHERE group_id = $1", groupID)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []model.Expense
	for rows.Next() {
		var item model.Expense
		err := rows.Scan(&item.ExpenseID, &item.Amount, &item.PayerID, &item.Description, &item.Created_at)
		if err != nil {
			http.Error(w, "Failed to scan items", http.StatusInternalServerError)
			return
		}
		cols, er := db.Query("SELECT user_id, share FROM item_splits WHERE item_id = $1", item.ExpenseID)
		if er != nil {
			http.Error(w, "Failed to scan items", http.StatusInternalServerError)
			return
		}
		for cols.Next() {
			var userShare model.UserShare
			err := cols.Scan(&userShare.UserID, &userShare.ShareAmount)
			if err != nil {
				http.Error(w, "Failed to scan item splits", http.StatusInternalServerError)
				return
			}
			item.Shares = append(item.Shares, userShare)
		}
		items = append(items, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func GetSettlements(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["groupId"]
	userIDStr := vars["userId"]
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group_id", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	var users model.UserIDsInput
	err = json.NewDecoder(r.Body).Decode(&users)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var settlements []model.UserShare
	for _, uID := range users.Users {
		var totalAmount int64 = 0
		if uID != int64(userID) {
			rows, err := db.Query("SELECT item_id FROM items WHERE group_id = $1 AND paid_by = $2", groupID, uID)
			if err != nil {
				http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var itemID int64
			for rows.Next() {
				err := rows.Scan(&itemID)
				if err != nil {
					http.Error(w, "Failed to scan items", http.StatusInternalServerError)
					return
				}
				var shareAmount int64
				err = db.QueryRow("SELECT share FROM item_splits WHERE item_id = $1 AND user_id = $2", itemID, int64(userID)).Scan(&shareAmount)

				if err == sql.ErrNoRows {
					continue
				} else if err != nil {
					http.Error(w, "Failed to scan item splits", http.StatusInternalServerError)
					return
				}
				totalAmount += shareAmount
			}

			rows, err = db.Query("SELECT item_id FROM items WHERE group_id = $1 AND paid_by = $2", groupID, userID)
			if err != nil {
				http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			for rows.Next() {
				err := rows.Scan(&itemID)
				if err != nil {
					http.Error(w, "Failed to scan items", http.StatusInternalServerError)
					return
				}
				var shareAmount int64
				err = db.QueryRow("SELECT share FROM item_splits WHERE item_id = $1 AND user_id = $2", itemID, uID).Scan(&shareAmount)

				if err == sql.ErrNoRows {
					continue
				} else if err != nil {
					http.Error(w, "Failed to scan item splits", http.StatusInternalServerError)
					return
				}
				totalAmount -= shareAmount
			}
			settlements = append(settlements, model.UserShare{UserID: uID, ShareAmount: totalAmount})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settlements)

}
