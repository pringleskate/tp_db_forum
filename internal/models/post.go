package models

type Post struct {
	ThreadSlagOrID
	ID       int       `json:"id,omitempty"`       // Идентификатор данного сообщения.
	Parent   int       `json:"parent,omitempty"`   // Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Author   string    `json:"author,omitempty"`   // Автор, написавший данное сообщение.
	Message  string    `json:"message,omitempty"`  // Собственно сообщение форума.
	IsEdited bool      `json:"isEdited,omitempty"` // Истина, если данное сообщение было изменено.
	Forum    string    `json:"forum,omitempty"`    // Идентификатор форума (slug) данного сообещния.
//	Created  time.Time `json:"created,omitempty"`  // Дата создания сообщения на форуме.
	Created  string `json:"created,omitempty"`
}

type PostCreate struct {
	Parent   int  `json:"parent,omitempty"`
	Author   string `json:"author,omitempty"`
	Message  string `json:"message,omitempty"`
}

type PostUpdate struct {
	ID      int    `json:"-"`
	Message string `json:"message"` // Собственно сообщение форума.
}

type PostFull struct {
	Post   *Post   `json:"post,omitempty"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
}
