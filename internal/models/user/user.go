package user

type User struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
	HashPassword string `json:"-"`
}
