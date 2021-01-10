package models

type Forum struct {
	Title   string `json:"title,omitempty"`   // Название форума.
	User    string `json:"user,omitempty"`    // Nickname пользователя, который отвечает за форум.
	Slug    string `json:"slug,omitempty"`    // Человекопонятный URL
	Posts   int  `json:"posts,omitempty"`   // Общее кол-во сообщений в данном форуме.
	Threads int  `json:"threads,omitempty"` // Общее кол-во ветвей обсуждения в данном форуме.
}

type ForumQueryParams struct {
	Slug string
	Limit int    // Максимальное кол-во возвращаемых записей.
	Since string // Идентификатор пользователя, с которого будут выводиться пользоватли (пользователь с данным идентификатором в результат не попадает).
	// Дата создания ветви обсуждения, с которой будут выводиться записи (ветвь обсуждения с указанной датой попадает в результат выборки)
	Desc bool // Флаг сортировки по убыванию
}
