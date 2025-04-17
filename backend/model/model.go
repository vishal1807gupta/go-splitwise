package model

type UserRequest struct {
	UserID   int64  `json:"-"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
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
}
