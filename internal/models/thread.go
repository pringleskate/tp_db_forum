package models

import "time"

type ThreadSlagOrID struct {
	ThreadID int    `json:"thread"`
	ThreadSlug     string `json:"-"`
}

type Thread struct {
	ForumSlug string    `json:"-"`
	ID        int       `json:"id"`                // Идентификатор ветки обсуждения.
	Title     string    `json:"title,omitempty"`   // Заголовок ветки обсуждения.
	Author    string    `json:"author,omitempty"`  // Пользователь, создавший данную тему.
	Forum     string    `json:"forum,omitempty"`   // Форум, в котором расположена данная ветка обсуждения.
	Message   string    `json:"message,omitempty"` // Описание ветки обсуждения.
	Votes     int       `json:"votes,omitempty"`   // Кол-во голосов непосредственно за данное сообщение форума.
	Slag      string    `json:"slug,omitempty"`    // Человекопонятный URL
	Created   time.Time `json:"created,omitempty"` // Дата создания ветки на форуме.
}

type ThreadUpdate struct {
	ThreadSlagOrID
	Title    string `json:"title"`   // Заголовок ветки обсуждения.
	Message  string `json:"message"` // Описание ветки обсуждения.
}

type ThreadQueryParams struct {
	ThreadSlagOrID
	Limit int    // Максимальное кол-во возвращаемых записей.
	Since int // Идентификатор поста, после которого будут выводиться записи (пост с данным идентификатором в результат не попадает)
	Sort  string
	Desc  bool   // Флаг сортировки по убыванию
}

type Vote struct {
	ThreadSlagOrID
	Nickname string `json:"nickname"` // Идентификатор пользователя.
	Voice    int32  `json:"voice"`    // Отданный голос.
}
