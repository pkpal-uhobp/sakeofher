package domain

type UserListInput struct {
	Query  string     `json:"query"`
	Status UserStatus `json:"status"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

type UserListResponse struct {
	Items []User `json:"items"`
	Total int64  `json:"total"`
	Limit int    `json:"limit"`
	Offset int   `json:"offset"`
}

type UpdateUserInput struct {
	TelegramUsername  *string     `json:"telegram_username,omitempty"`
	TelegramFirstName *string     `json:"telegram_first_name,omitempty"`
	TelegramLastName  *string     `json:"telegram_last_name,omitempty"`
	LanguageCode      *string     `json:"language_code,omitempty"`
	Alias             *string     `json:"alias,omitempty"`
	Status            *UserStatus `json:"status,omitempty"`
}
