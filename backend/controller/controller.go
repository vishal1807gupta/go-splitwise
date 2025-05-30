package controller

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-splitwise/cloudfareR2"
	"go-splitwise/email"
	model "go-splitwise/model"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

var db *sql.DB

func configureDatabase() {
	// Set explicit connection pool parameters
	db.SetMaxOpenConns(20)                  // Limit total connections
	db.SetMaxIdleConns(5)                   // Limit idle connections
	db.SetConnMaxLifetime(30 * time.Minute) // Recycle connections periodically
	db.SetConnMaxIdleTime(5 * time.Minute)  // Don't keep idle connections too long

	// Verify connections work
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
}

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using existing environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")

	if !strings.Contains(databaseURL, "?") {
		databaseURL += "?disable_prepared_statements=true"
	} else {
		databaseURL += "&disable_prepared_statements=true"
	}

	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}

	configureDatabase()

	fmt.Println("Successfully connected!")
}

type PasswordResetService struct {
	emailService *email.EmailService
	codeTTL      time.Duration
}

type ReminderService struct {
	emailService *email.EmailService
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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
		jsonError(w, "Name must be at least 3 characters long.", http.StatusBadRequest)
		return false
	}
	if len(user.Password) < 6 {
		jsonError(w, "Password must be at least 6 characters long.", http.StatusBadRequest)
		return false
	}
	if !isValidEmail(user.Email) {
		jsonError(w, "Please enter a valid email address.", http.StatusBadRequest)
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

func getUserByEmail(email string) (*model.UserRequest, error) {
	user := &model.UserRequest{}
	query := "SELECT user_id, email, password FROM users WHERE email = $1"
	err := db.QueryRow(query, email).Scan(&user.UserID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func newPasswordResetService(emailService *email.EmailService) *PasswordResetService {
	return &PasswordResetService{
		emailService: emailService,
		codeTTL:      15 * time.Minute,
	}
}

func newReminderService(emailService *email.EmailService) *ReminderService {
	return &ReminderService{
		emailService: emailService,
	}
}

func savePasswordReset(reset model.PasswordReset) error {
	query := `
		INSERT INTO password_resets (user_id, email, code, expires_at, used)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Exec(query, reset.UserID, reset.Email, reset.Code, reset.ExpiresAt, reset.Used)
	return err
}

func (s *PasswordResetService) RequestPasswordReset(emailAddress string) error {
	user, err := getUserByEmail(emailAddress)
	if err != nil {
		return errors.New("user not found")
	}

	// Generate a verification code
	code := email.GenerateVerificationCode()

	// Store code in database
	reset := model.PasswordReset{
		UserID:    user.UserID,
		Email:     emailAddress,
		Code:      code,
		ExpiresAt: time.Now().Add(s.codeTTL),
		Used:      false,
	}

	if err := savePasswordReset(reset); err != nil {
		return err
	}

	// Send email with code
	if err := s.emailService.SendPasswordResetCode(emailAddress, code); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func getLatestPasswordResetByEmail(emailAddress string) (*model.PasswordReset, error) {
	reset := &model.PasswordReset{}
	query := `
		SELECT id, user_id, email, code, expires_at, used, created_at
		FROM password_resets
		WHERE email = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := db.QueryRow(query, emailAddress).Scan(
		&reset.ID, &reset.UserID, &reset.Email, &reset.Code,
		&reset.ExpiresAt, &reset.Used, &reset.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no reset request found")
		}
		return nil, err
	}
	return reset, nil
}

func updateUserPassword(userID int64, hashedPassword string) error {
	query := "UPDATE users SET password = $1 WHERE user_id = $2"
	result, err := db.Exec(query, hashedPassword, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no user found with that ID")
	}

	return nil
}

func markPasswordResetCodeAsUsed(resetID int64) error {
	query := "UPDATE password_resets SET used = true WHERE id = $1"
	_, err := db.Exec(query, resetID)
	return err
}

func (s *PasswordResetService) ValidateCodeAndResetPassword(email, code, newPassword string) error {
	// Get the most recent reset request for this email
	reset, err := getLatestPasswordResetByEmail(email)
	if err != nil {
		return errors.New("invalid or expired password reset request, please try again")
	}

	// Check if code is valid
	if reset.Code != code {
		return errors.New("the verification code you entered is invalid, please check and try again")
	}

	// Check if code is expired
	if time.Now().After(reset.ExpiresAt) {
		return errors.New("this verification code has expired, please request a new code")
	}

	// Check if code was already used
	if reset.Used {
		return errors.New("this verification code has already been used, please request a new code")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		return errors.New("we couldn't process your password, please try a different password")
	}

	// Update user password in DB
	if err := updateUserPassword(reset.UserID, string(hashedPassword)); err != nil {
		return errors.New("failed to update password, please try again later")
	}

	// Mark code as used
	if err := markPasswordResetCodeAsUsed(reset.ID); err != nil {
		return errors.New("failed to complete password reset, please try again")
	}

	return nil
}

func RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var req model.RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request. Please check your input and try again.", http.StatusBadRequest)
		return
	}

	emailService, err := email.NewEmailService(
		r.Context(),
		os.Getenv("AWS_REGION"),
		os.Getenv("EMAIL_SENDER"),
	)
	if err != nil {
		jsonError(w, "We're experiencing technical difficulties. Please try again later.", http.StatusInternalServerError)
		return
	}

	resetService := newPasswordResetService(emailService)
	err = resetService.RequestPasswordReset(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Error",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If your email is registered, you will receive a verification code",
	})
}

func ResetPasswordCompleteHandler(w http.ResponseWriter, r *http.Request) {
	var req model.ResetPasswordCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request. Please check your input and try again.", http.StatusBadRequest)
		return
	}

	emailService, err := email.NewEmailService(
		r.Context(),
		os.Getenv("AWS_REGION"),
		os.Getenv("EMAIL_SENDER"),
	)
	if err != nil {
		jsonError(w, "We're experiencing technical difficulties. Please try again later.", http.StatusInternalServerError)
		return
	}

	resetService := newPasswordResetService(emailService)
	err = resetService.ValidateCodeAndResetPassword(req.Email, req.Code, req.NewPassword)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password has been reset successfully",
	})
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user model.UserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		jsonError(w, "Invalid input. Please check your information and try again.", http.StatusBadRequest)
		return
	}

	if !validateUserInput(user, w) {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "We're experiencing technical difficulties. Please try again later.", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING user_id`
	err = db.QueryRow(query, user.Name, user.Email, string(hashedPassword)).Scan(&user.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			jsonError(w, "This email address is already registered. Please use a different email or login to your account.", http.StatusInternalServerError)
		} else {
			jsonError(w, "We couldn't create your account. Please try again later.", http.StatusInternalServerError)
		}
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
		jsonError(w, "Failed to create session. Please try logging in again.", http.StatusInternalServerError)
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
		jsonError(w, "Invalid input. Please check your information and try again.", http.StatusBadRequest)
		return
	}

	var registeredUser model.UserResponse
	var hashedPassword string

	err = db.QueryRow("SELECT user_id, name, email, password FROM users WHERE email = $1",
		user.Email).Scan(&registeredUser.UserID, &registeredUser.Name, &registeredUser.Email, &hashedPassword)

	if err != nil {
		jsonError(w, "Invalid email or password. Please check your credentials and try again.", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		jsonError(w, "Invalid email or password. Please check your credentials and try again.", http.StatusUnauthorized)
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
		jsonError(w, "Failed to create session. Please try logging in again.", http.StatusInternalServerError)
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
		jsonError(w, "Invalid Google authentication. Please try again or use a different login method.", http.StatusUnauthorized)
		return
	}

	claims := payload.Claims

	email := claims["email"].(string)
	name := claims["name"].(string)
	googleID := claims["sub"].(string)

	var user model.UserResponse
	err = db.QueryRow("SELECT user_id, name FROM users WHERE google_id = $1", googleID).Scan(&user.UserID, &user.Name)
	if err == sql.ErrNoRows {

		err = db.QueryRow("SELECT user_id, name FROM users WHERE email = $1", email).Scan(&user.UserID, &user.Name)
		if err == sql.ErrNoRows {
			err = db.QueryRow("INSERT INTO users (name, email, google_id) VALUES ($1, $2, $3) RETURNING user_id, name", name, email, googleID).Scan(&user.UserID, &user.Name)
			if err != nil {
				jsonError(w, "Failed to create account. Please try again later.", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			jsonError(w, "Failed to retrieve account information. Please try again later.", http.StatusInternalServerError)
			return
		} else {
			_, err = db.Exec("UPDATE users SET google_id = $1 WHERE user_id = $2", googleID, user.UserID)
			if err != nil {
				jsonError(w, "Failed to update account. Please try again later.", http.StatusInternalServerError)
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
		jsonError(w, "Failed to create session. Please try logging in again.", http.StatusInternalServerError)
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
			"message": "Your session has expired. Please log in again to continue.",
		})
		return
	}

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

func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		jsonError(w, "Invalid input. Please check your information and try again.", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "We're experiencing technical difficulties. Please try again later.", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET password = $1 WHERE email = $2", string(hashedPassword), input.Email)
	if err != nil {
		jsonError(w, "Failed to update password. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully",
	})
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
		jsonError(w, "Invalid user ID. Please try again.", http.StatusBadRequest)
		return
	}

	groups, err := fetchAllGroupsByUserID(int64(userID))

	if err != nil {
		jsonError(w, "Failed to fetch your groups. Please try again later.", http.StatusInternalServerError)
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
		jsonError(w, "Invalid user ID. Please try again.", http.StatusBadRequest)
		return
	}

	var group model.Group
	err = json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		jsonError(w, "Invalid input. Please check your group information and try again.", http.StatusBadRequest)
		return
	}

	if len(strings.TrimSpace(group.GroupName)) < 3 {
		jsonError(w, "Group name must be at least 3 characters long.", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO groups (name) VALUES ($1) RETURNING group_id`
	err = db.QueryRow(query, group.GroupName).Scan(&group.GroupID)
	if err != nil {
		jsonError(w, "Failed to create group. Please try again later.", http.StatusInternalServerError)
		return
	}
	err = insertUserIntoGroup(int(group.GroupID), int64(userID))
	if err != nil {
		jsonError(w, "Failed to add you to the group. Please try again later.", http.StatusInternalServerError)
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
		jsonError(w, "Invalid group ID. Please try again.", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&groupUsers); err != nil {
		jsonError(w, "Invalid input. Please check your information and try again.", http.StatusBadRequest)
		return
	}

	for _, userID := range groupUsers.UserIDs {
		err := insertUserIntoGroup(groupID, userID)
		if err != nil {
			jsonError(w, "Failed to add users to the group. Please try again later.", http.StatusInternalServerError)
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
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	groupID, ok := vars["groupId"]
	if !ok {
		jsonError(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Use direct string formatting instead of parameter binding
	query := fmt.Sprintf("SELECT u.user_id, u.name, u.email FROM users u JOIN group_users gu ON u.user_id = gu.user_id WHERE gu.group_id = %s", groupID)

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying group users: %v", err)
		jsonError(w, "Failed to fetch group members.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var u model.UserResponse
		err := rows.Scan(&u.UserID, &u.Name, &u.Email)
		if err != nil {
			jsonError(w, "Failed to process group members.", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// Return empty array if no users found
	if users == nil {
		users = []model.UserResponse{}
	}

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
		jsonError(w, "Failed to fetch available users. Please try again later.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var u model.UserResponse
		err := rows.Scan(&u.UserID, &u.Name, &u.Email)
		if err != nil {
			jsonError(w, "Failed to process available users. Please try again later.", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func AddExpense(w http.ResponseWriter, r *http.Request) {
	var expense model.Expense
	err := json.NewDecoder(r.Body).Decode(&expense)
	if err != nil {
		jsonError(w, "Invalid expense data. Please check your information and try again.", http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	groupID := vars["groupId"]

	query := `INSERT INTO items (group_id, amount, paid_by, description) VALUES ($1, $2, $3, $4) RETURNING item_id`
	err = db.QueryRow(query, groupID, expense.Amount, expense.PayerID, expense.Description).Scan(&expense.ExpenseID)
	if err != nil {
		jsonError(w, "Failed to create expense. Please try again later.", http.StatusInternalServerError)
		return
	}

	err = calculateBalances(&expense)

	if err != nil {
		query = `DELETE FROM items WHERE item_id = $1`
		_, newErr := db.Exec(query, expense.ExpenseID)
		if newErr != nil {
			jsonError(w, "An error occurred while processing your expense. Please try again later.", http.StatusInternalServerError)
			return
		}

		if strings.Contains(err.Error(), "sum of shares is not equal to the amount") {
			jsonError(w, "The sum of individual shares must equal the total amount.", http.StatusInternalServerError)
		} else if strings.Contains(err.Error(), "sum of shares is not equal to 100") {
			jsonError(w, "When splitting by percentage, all percentages must add up to 100%.", http.StatusInternalServerError)
		} else {
			jsonError(w, "Failed to calculate balances. Please check your expense details and try again.", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

func GetItemsByGroupId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	groupID := vars["groupId"]

	// Use direct string formatting
	query := fmt.Sprintf("SELECT item_id, amount, paid_by, description, created_at FROM items WHERE group_id = %s ORDER BY created_at DESC", groupID)

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying items: %v", err)
		jsonError(w, "Failed to fetch expenses.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []model.Expense
	for rows.Next() {
		var item model.Expense
		err := rows.Scan(&item.ExpenseID, &item.Amount, &item.PayerID, &item.Description, &item.Created_at)
		if err != nil {
			jsonError(w, "Failed to process expenses.", http.StatusInternalServerError)
			return
		}

		// Get shares with direct string formatting
		sharesQuery := fmt.Sprintf("SELECT user_id, share FROM item_splits WHERE item_id = %d", item.ExpenseID)
		shareRows, err := db.Query(sharesQuery)
		if err != nil {
			jsonError(w, "Failed to fetch expense details.", http.StatusInternalServerError)
			return
		}

		for shareRows.Next() {
			var userShare model.UserShare
			err := shareRows.Scan(&userShare.UserID, &userShare.ShareAmount)
			if err != nil {
				shareRows.Close()
				jsonError(w, "Failed to process expense shares.", http.StatusInternalServerError)
				return
			}
			item.Shares = append(item.Shares, userShare)
		}
		shareRows.Close()

		items = append(items, item)
	}

	// Return empty array if no items found
	if items == nil {
		items = []model.Expense{}
	}

	json.NewEncoder(w).Encode(items)
}

func GetSettlements(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	groupIDStr := vars["groupId"]
	userIDStr := vars["userId"]

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		jsonError(w, "Invalid group ID.", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		jsonError(w, "Invalid user ID.", http.StatusBadRequest)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "Failed to read request body.", http.StatusBadRequest)
		return
	}

	// Unmarshal the body
	var requestData struct {
		Users []int64 `json:"users"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		// Log the body content for debugging
		log.Printf("Invalid request body: %s", string(body))
		jsonError(w, "Invalid request format.", http.StatusBadRequest)
		return
	}

	// Simple validation
	if len(requestData.Users) == 0 {
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	var settlements []map[string]interface{}

	for _, otherUserID := range requestData.Users {
		// Skip self
		if otherUserID == int64(userID) {
			continue
		}

		// Calculate settlement using direct queries
		itemsOtherUserPaid := fmt.Sprintf(
			"SELECT item_id FROM items WHERE group_id = %d AND paid_by = %d",
			groupID, otherUserID)

		var totalAmount int64 = 0

		// Process items paid by other user
		itemRows, err := db.Query(itemsOtherUserPaid)
		if err != nil {
			log.Printf("Error fetching items paid by other user: %v", err)
			continue
		}

		for itemRows.Next() {
			var itemID int64
			err := itemRows.Scan(&itemID)
			if err != nil {
				log.Printf("Error scanning item ID: %v", err)
				continue
			}

			// Get user's share of this item
			var shareAmount int64
			shareQuery := fmt.Sprintf(
				"SELECT share FROM item_splits WHERE item_id = %d AND user_id = %d",
				itemID, userID)

			err = db.QueryRow(shareQuery).Scan(&shareAmount)
			if err == sql.ErrNoRows {
				continue
			} else if err != nil {
				log.Printf("Error getting share: %v", err)
				continue
			}

			totalAmount += shareAmount
		}
		itemRows.Close()

		// Process transactions already made
		transQuery := fmt.Sprintf(
			"SELECT amount FROM transactions WHERE group_id = %d AND user_id = %d AND payer_id = %d",
			groupID, userID, otherUserID)

		transRows, err := db.Query(transQuery)
		if err != nil {
			log.Printf("Error fetching transactions: %v", err)
			continue
		}

		for transRows.Next() {
			var amount int64
			err := transRows.Scan(&amount)
			if err != nil {
				log.Printf("Error scanning transaction amount: %v", err)
				continue
			}
			totalAmount -= amount
		}
		transRows.Close()

		// Items paid by current user
		itemsUserPaid := fmt.Sprintf(
			"SELECT item_id FROM items WHERE group_id = %d AND paid_by = %d",
			groupID, userID)

		itemRows, err = db.Query(itemsUserPaid)
		if err != nil {
			log.Printf("Error fetching items paid by user: %v", err)
			continue
		}

		for itemRows.Next() {
			var itemID int64
			err := itemRows.Scan(&itemID)
			if err != nil {
				log.Printf("Error scanning item ID: %v", err)
				continue
			}

			// Get other user's share of this item
			var shareAmount int64
			shareQuery := fmt.Sprintf(
				"SELECT share FROM item_splits WHERE item_id = %d AND user_id = %d",
				itemID, otherUserID)

			err = db.QueryRow(shareQuery).Scan(&shareAmount)
			if err == sql.ErrNoRows {
				continue
			} else if err != nil {
				log.Printf("Error getting share: %v", err)
				continue
			}

			totalAmount -= shareAmount
		}
		itemRows.Close()

		// More transactions
		transQuery = fmt.Sprintf(
			"SELECT amount FROM transactions WHERE group_id = %d AND user_id = %d AND payer_id = %d",
			groupID, otherUserID, userID)

		transRows, err = db.Query(transQuery)
		if err != nil {
			log.Printf("Error fetching transactions: %v", err)
			continue
		}

		for transRows.Next() {
			var amount int64
			err := transRows.Scan(&amount)
			if err != nil {
				log.Printf("Error scanning transaction amount: %v", err)
				continue
			}
			totalAmount += amount
		}
		transRows.Close()

		// Add to settlements if non-zero
		if totalAmount != 0 {
			settlement := map[string]interface{}{
				"user_id":      otherUserID,
				"share_amount": totalAmount,
			}
			settlements = append(settlements, settlement)
		}
	}

	// Return empty array if no settlements
	if settlements == nil {
		settlements = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(settlements)
}

func GetMemoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	groupId, err := strconv.Atoi(vars["groupId"])
	if err != nil {
		jsonError(w, "Invalid group ID. Please try again.", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
		SELECT id, group_id, filename, image_url, created_at 
		FROM memories 
		WHERE group_id = $1 
		ORDER BY created_at DESC`, groupId)
	if err != nil {
		log.Printf("Error querying memories: %v", err)
		jsonError(w, "Failed to fetch memories. Please try again later.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var memories []model.Memory
	for rows.Next() {
		var memory model.Memory
		err := rows.Scan(
			&memory.ID,
			&memory.GroupID,
			&memory.Filename,
			&memory.ImageURL,
			&memory.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning memory row: %v", err)
			continue
		}

		memories = append(memories, memory)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating memory rows: %v", err)
		jsonError(w, "Error fetching memories. Please try again later.", http.StatusInternalServerError)
		return
	}

	response := model.MemoryResponse{
		Success:  true,
		Memories: memories,
	}
	json.NewEncoder(w).Encode(response)
}

func UploadMemoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	r2Storage, err := cloudfareR2.NewR2Storage(
		os.Getenv("R2_ACCESS_KEY"),
		os.Getenv("R2_SECRET_KEY"),
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_BUCKET_NAME"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize R2 storage: %v", err)
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		jsonError(w, "The file you uploaded is too large. Please upload an image smaller than 5MB.", http.StatusBadRequest)
		return
	}

	groupIdStr := r.FormValue("groupId")
	if groupIdStr == "" {
		jsonError(w, "Missing group ID. Please try again.", http.StatusBadRequest)
		return
	}

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		jsonError(w, "Invalid group ID. Please try again.", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		jsonError(w, "Error retrieving file. Please try again.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		jsonError(w, "Please select a valid image file (JPEG, PNG, GIF, WebP, BMP, or TIFF).", http.StatusBadRequest)
		return
	}

	// Upload file to R2
	filename, imageURL, err := r2Storage.UploadFile(file, fileHeader)
	if err != nil {
		log.Printf("Error uploading file to R2: %v", err)
		jsonError(w, "Failed to upload your image. Please try again later.", http.StatusInternalServerError)
		return
	}

	var memory model.Memory
	err = db.QueryRow(`
		INSERT INTO memories (group_id, filename, image_url, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, group_id, filename, image_url, created_at`,
		groupId, filename, imageURL).Scan(
		&memory.ID,
		&memory.GroupID,
		&memory.Filename,
		&memory.ImageURL,
		&memory.CreatedAt,
	)
	if err != nil {
		log.Printf("Error inserting memory: %v", err)
		// Try to delete the file from R2 since the DB insert failed
		if delErr := r2Storage.DeleteFile(filename); delErr != nil {
			log.Printf("Error deleting file from R2 after DB insert failed: %v", delErr)
		}
		jsonError(w, "Failed to save memory. Please try again later.", http.StatusInternalServerError)
		return
	}

	response := model.MemoryResponse{
		Success: true,
		Message: "Memory uploaded successfully",
		Memory:  &memory,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteMemoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	r2Storage, err := cloudfareR2.NewR2Storage(
		os.Getenv("R2_ACCESS_KEY"),
		os.Getenv("R2_SECRET_KEY"),
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_BUCKET_NAME"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize R2 storage: %v", err)
	}

	vars := mux.Vars(r)
	memoryId, err := strconv.Atoi(vars["memoryId"])
	if err != nil {
		jsonError(w, "Invalid memory ID. Please try again.", http.StatusBadRequest)
		return
	}

	var filename string
	err = db.QueryRow("SELECT filename FROM memories WHERE id = $1", memoryId).Scan(&filename)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, "Memory not found or already deleted.", http.StatusNotFound)
		} else {
			log.Printf("Error fetching memory: %v", err)
			jsonError(w, "Failed to retrieve memory information. Please try again later.", http.StatusInternalServerError)
		}
		return
	}

	result, err := db.Exec("DELETE FROM memories WHERE id = $1", memoryId)
	if err != nil {
		log.Printf("Error deleting memory from database: %v", err)
		jsonError(w, "Failed to delete memory. Please try again later.", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		jsonError(w, "Memory not found or already deleted.", http.StatusNotFound)
		return
	}

	// Delete the file from R2
	if err := r2Storage.DeleteFile(filename); err != nil {
		log.Printf("Warning: Could not delete file from R2: %v", err)
		// Continue anyway, as the database record is already deleted
	}

	response := model.MemoryResponse{
		Success: true,
		Message: "Memory deleted successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
		"image/tiff": true,
	}
	return validTypes[contentType]
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	groupIDStr, ok := vars["groupId"]
	if !ok {
		jsonError(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Use string formatting to avoid prepared statements
	query := fmt.Sprintf(`
        SELECT id, user_id, payer_id, group_id, amount, created_at 
        FROM transactions 
        WHERE group_id = %s 
        ORDER BY created_at DESC`, groupIDStr)

	// Log the query for debugging
	log.Printf("Executing query: %s", query)

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying transactions: %v", err)
		// Return empty array on error
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}

	for rows.Next() {
		var id, userID, payerID, groupID, amount int64
		var createdAt time.Time

		err := rows.Scan(&id, &userID, &payerID, &groupID, &amount, &createdAt)
		if err != nil {
			log.Printf("Error scanning transaction row: %v", err)
			continue
		}

		transaction := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"payer_id":   payerID,
			"group_id":   groupID,
			"amount":     amount,
			"created_at": createdAt,
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
	}

	// Always return an array (even if empty)
	if transactions == nil {
		transactions = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(transactions)
}

func InsertTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	groupId, err := strconv.Atoi(vars["groupId"])
	if err != nil {
		jsonError(w, "Invalid group ID. Please try again.", http.StatusBadRequest)
		return
	}
	var transaction model.Transactions
	err = json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		jsonError(w, "Invalid transaction data. Please check your information and try again.", http.StatusBadRequest)
		return
	}
	transaction.GroupID = int64(groupId)
	query := `INSERT INTO transactions (user_id, payer_id, group_id, amount) VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(query, transaction.UserID, transaction.PayerID, groupId, transaction.Amount).Scan(&transaction.ID)
	if err != nil {
		jsonError(w, "Failed to record transaction. Please try again later.", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(transaction)
}

func (s *ReminderService) SendMonthlyBalanceReminders() {
	log.Println("Starting monthly balance reminder job")
	startTime := time.Now()

	// Create a job record
	jobID, err := logReminderJob("monthly_balance", "running") // inserting new job in db (jobtype, status)
	if err != nil {
		log.Printf("Error logging reminder job: %v", err)
	}

	rows, err := db.Query("SELECT user_id, name, email FROM users")
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		updateReminderJob(jobID, "failed", err.Error(), 0, 0)
		return
	}
	defer rows.Close()

	var users []model.UserResponse

	for rows.Next() {
		var user model.UserResponse
		if err := rows.Scan(&user.UserID, &user.Name, &user.Email); err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		users = append(users, user)
	}

	successCount := 0
	errorCount := 0

	for _, user := range users {
		groups, err := fetchAllGroupsByUserID(user.UserID)
		if err != nil {
			log.Printf("Error fetching groups for user %d: %v", user.UserID, err)
			errorCount++
			continue
		}

		var allBalances []model.Balance

		for _, group := range groups {
			otherUsers, err := getGroupUserIDs(int(group.GroupID), user.UserID)
			if err != nil {
				log.Printf("Error fetching group users: %v", err)
				continue
			}

			if len(otherUsers) == 0 {
				continue
			}

			settlements, err := calculateSettlements(int(group.GroupID), int(user.UserID), otherUsers)
			if err != nil {
				log.Printf("Error calculating settlements: %v", err)
				continue
			}

			for _, settlement := range settlements {
				var otherUserName string
				err := db.QueryRow("SELECT name FROM users WHERE user_id = $1", settlement.UserID).Scan(&otherUserName)
				if err != nil {
					log.Printf("Error fetching user name: %v", err)
					continue
				}

				balance := model.Balance{
					OtherUserID:   settlement.UserID,
					OtherUserName: otherUserName,
					Amount:        settlement.ShareAmount,
					GroupName:     group.GroupName,
				}

				allBalances = append(allBalances, balance)
			}
		}

		if len(allBalances) == 0 {
			continue
		}

		err = s.emailService.SendMonthlyBalanceReminder(user.Email, user.Name, allBalances)
		if err != nil {
			log.Printf("Error sending reminder to %s: %v", user.Email, err)
			errorCount++
		} else {
			// Log that reminder was sent
			err = logReminderSent(user.UserID, "monthly_balance")
			if err != nil {
				log.Printf("Error logging reminder: %v", err)
			}
			successCount++
		}

		// Add a small delay to avoid SES rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	duration := time.Since(startTime)
	log.Printf("Completed monthly balance reminder job. Success: %d, Errors: %d, Duration: %v",
		successCount, errorCount, duration)

	updateReminderJob(jobID, "completed", "", successCount, errorCount)
}

func getGroupUserIDs(groupID int, excludeUserID int64) ([]int64, error) {
	rows, err := db.Query(`
		SELECT user_id 
		FROM group_users 
		WHERE group_id = $1 AND user_id != $2`,
		groupID, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

func calculateSettlements(groupID, userID int, otherUserIDs []int64) ([]model.UserShare, error) {
	var settlements []model.UserShare

	for _, otherUserID := range otherUserIDs {
		var totalAmount int64 = 0

		rows1, err := db.Query("SELECT item_id FROM items WHERE group_id = $1 AND paid_by = $2",
			groupID, otherUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch items: %w", err)
		}
		defer rows1.Close()

		for rows1.Next() {
			var itemID int64
			if err := rows1.Scan(&itemID); err != nil {
				return nil, fmt.Errorf("failed to scan items: %w", err)
			}

			var shareAmount int64
			err = db.QueryRow("SELECT share FROM item_splits WHERE item_id = $1 AND user_id = $2",
				itemID, int64(userID)).Scan(&shareAmount)

			if err == sql.ErrNoRows {
				continue
			} else if err != nil {
				return nil, fmt.Errorf("failed to get share: %w", err)
			}

			totalAmount += shareAmount
		}

		rows2, err := db.Query("SELECT amount FROM transactions WHERE group_id = $1 AND user_id = $2 AND payer_id = $3",
			groupID, userID, otherUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch transactions: %w", err)
		}
		defer rows2.Close()

		for rows2.Next() {
			var amount int64
			if err := rows2.Scan(&amount); err != nil {
				return nil, fmt.Errorf("failed to scan transactions: %w", err)
			}
			totalAmount -= amount
		}

		rows3, err := db.Query("SELECT item_id FROM items WHERE group_id = $1 AND paid_by = $2",
			groupID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch items: %w", err)
		}
		defer rows3.Close()

		for rows3.Next() {
			var itemID int64
			if err := rows3.Scan(&itemID); err != nil {
				return nil, fmt.Errorf("failed to scan items: %w", err)
			}

			var shareAmount int64
			err = db.QueryRow("SELECT share FROM item_splits WHERE item_id = $1 AND user_id = $2",
				itemID, otherUserID).Scan(&shareAmount)

			if err == sql.ErrNoRows {
				continue
			} else if err != nil {
				return nil, fmt.Errorf("failed to get share: %w", err)
			}

			totalAmount -= shareAmount
		}

		rows4, err := db.Query("SELECT amount FROM transactions WHERE group_id = $1 AND user_id = $2 AND payer_id = $3",
			groupID, otherUserID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch transactions: %w", err)
		}
		defer rows4.Close()

		for rows4.Next() {
			var amount int64
			if err := rows4.Scan(&amount); err != nil {
				return nil, fmt.Errorf("failed to scan transactions: %w", err)
			}
			totalAmount += amount
		}

		if totalAmount != 0 {
			settlements = append(settlements, model.UserShare{
				UserID:      otherUserID,
				ShareAmount: totalAmount,
			})
		}
	}

	return settlements, nil
}

// logReminderJob records the start of a reminder job
func logReminderJob(jobType, status string) (int64, error) {
	query := `
		INSERT INTO reminder_jobs (job_type, status, started_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var jobID int64
	err := db.QueryRow(query, jobType, status, time.Now()).Scan(&jobID)
	return jobID, err
}

// updateReminderJob updates the status of a reminder job
func updateReminderJob(jobID int64, status, errorMsg string, success, failed int) error {
	query := `
		UPDATE reminder_jobs
		SET status = $2, error = $3, success_count = $4, error_count = $5, completed_at = $6
		WHERE id = $1
	`

	_, err := db.Exec(query, jobID, status, errorMsg, success, failed, time.Now())
	return err
}

// logReminderSent records that a reminder was sent to a user
func logReminderSent(userID int64, reminderType string) error {
	query := `
		INSERT INTO reminder_logs (user_id, reminder_type, sent_at)
		VALUES ($1, $2, $3)
	`

	_, err := db.Exec(query, userID, reminderType, time.Now())
	return err
}

func TriggerMonthlyReminders(w http.ResponseWriter, r *http.Request) {

	authToken := r.Header.Get("X-Auth-Token")
	expectedToken := os.Getenv("AUTH_TOKEN")

	if expectedToken == "" {
		log.Println("Warning: AUTH_TOKEN environment variable not set")
	} else if authToken != expectedToken {
		log.Printf("Unauthorized access attempt with token: %s", authToken)
		jsonError(w, "Unauthorized access. Please check your credentials and try again.", http.StatusUnauthorized)
		return
	}

	emailService, err := email.NewEmailService(
		r.Context(),
		os.Getenv("AWS_REGION"),
		os.Getenv("EMAIL_SENDER"),
	)
	if err != nil {
		log.Printf("Error initializing email service: %v", err)
		jsonError(w, "Failed to initialize email service", http.StatusInternalServerError)
		return
	}

	reminderService := newReminderService(emailService)

	go func() {
		log.Println("Starting monthly balance reminder job triggered by AWS Lambda")
		reminderService.SendMonthlyBalanceReminders()
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Monthly balance reminder job started"}`)
}
