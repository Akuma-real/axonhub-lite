package objects

type UserInfo struct {
	ID             GUID    `json:"id"`
	Email          string  `json:"email"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	IsOwner        bool    `json:"isOwner"`
	PreferLanguage string  `json:"preferLanguage"`
	Avatar         *string `json:"avatar,omitempty"`
	HasPassword    bool    `json:"hasPassword"`
}
