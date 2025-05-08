package model

import (
	"time"
)

type UserRequest struct {
	UserID     int64  `json:"-"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe,omitempty"`
}

type UserResponse struct {
	UserID   int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Group struct {
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
}

type GroupUsers struct {
	UserIDs []int64 `json:"user_ids"`
}

type UserShare struct {
	UserID      int64 `json:"user_id"`
	ShareAmount int64 `json:"share_amount"`
}

type Expense struct {
	ExpenseID   int64       `json:"expense_id"`
	Amount      int64       `json:"amount"`
	PayerID     int64       `json:"payer_id"`
	Description string      `json:"description"`
	ExpenseType string      `json:"expense_type"`
	Shares      []UserShare `json:"user_shares"`
	Created_at  string      `json:"date"`
}

type UserIDsInput struct {
	Users []int64 `json:"users"`
}

type PasswordReset struct {
	ID        int64
	UserID    int64
	Email     string
	Code      string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

type ResetPasswordCompleteRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type Memory struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"groupId"`
	Filename  string    `json:"filename"`
	ImageURL  string    `json:"imageUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type MemoryResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message,omitempty"`
	Memory   *Memory  `json:"memory,omitempty"`
	Memories []Memory `json:"memories,omitempty"`
}

type Transactions struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	PayerID   int64     `json:"payer_id"`
	GroupID   int64     `json:"group_id"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type Balance struct {
	OtherUserID   int64
	OtherUserName string
	Amount        int64
	GroupName     string
}
