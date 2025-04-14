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
