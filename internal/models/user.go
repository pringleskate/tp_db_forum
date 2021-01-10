package models

type User struct {
	Nickname string `json:"nickname,omitempty"` // Имя пользователя (уникальное поле). Данное поле допускает только латиницу, цифры и знак подчеркивания.
	FullName string `json:"fullname,omitempty"`
	About    string `json:"about,omitempty"` // Описание пользователя.
	Email    string `json:"email,omitempty"` // Почтовый адрес пользователя (уникальное поле).
}
