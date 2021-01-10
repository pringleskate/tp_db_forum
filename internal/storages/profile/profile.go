package profile

import (
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/pringleskate/tp_db_forum/internal/models"
)

type Storage interface {
	CreateUser(input models.User) (user models.User, err error)
	GetProfile(input string) (user models.User, err error)
	UpdateProfile(input models.User) (user models.User, err error)
	GetUsers(input models.ForumQueryParams, forumID int) (users []models.User, err error)
	GetUserForPost(input string,  user *models.User) (err error)
	GetUserIDByNickname(input string) (userID int, err error)
	GetEmailConflictUser(email string) (user models.User, err error)
}

type storage struct {
	db *pgx.ConnPool
}

/* constructor */
func NewStorage(db *pgx.ConnPool) Storage {
	return &storage{
		db: db,
	}
}

var (
	selectEmpty = "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 ORDER BY u.nickname LIMIT $2"
	selectWithSince = "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 AND u.nickname > $2 ORDER BY u.nickname LIMIT $3"
	selectWithDesc = "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 ORDER BY u.nickname DESC LIMIT $2"
	selectWithSinceDesc =  "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 AND u.nickname < $2 ORDER BY u.nickname DESC LIMIT $3"

	updateFull = "UPDATE users SET nickname = $1, fullname = $2, email = $3, about = $4 WHERE nickname = $5 RETURNING fullname, email, about, nickname"
	updateEmail = "UPDATE users SET nickname = $1, email = $2 WHERE nickname = $3 RETURNING fullname, email, about, nickname"
	updateFullname = "UPDATE users SET nickname = $1, fullname = $2 WHERE nickname = $3 RETURNING fullname, email, about, nickname"
	updateAbout = "UPDATE users SET nickname = $1, about = $2 WHERE nickname = $3 RETURNING fullname, email, about, nickname"
	updateEmailFullname = "UPDATE users SET nickname = $1, fullname = $2, email = $3 WHERE nickname = $4 RETURNING fullname, email, about, nickname"
	updateEmailAbout = "UPDATE users SET nickname = $1, email = $2, about = $3 WHERE nickname = $4 RETURNING fullname, email, about, nickname"
	updateFullnameAbout = "UPDATE users SET nickname = $1, fullname = $2, about = $3 WHERE nickname = $4 RETURNING fullname, email, about, nickname"
)

func (s *storage) CreateUser(input models.User) (user models.User, err error) {
	_, err = s.db.Exec("INSERT INTO users (nickname, email, fullname, about) VALUES ($1, $2, $3, $4)",
		input.Nickname, input.Email, input.FullName, input.About)

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return user, models.ServError{Code: 409, Message: "conflict user"}
		default:
			return user, models.ServError{Code: 500}
		}
	}

	user.Nickname = input.Nickname
	user.FullName = input.FullName
	user.Email = input.Email
	user.About = input.About

	return
}

func (s *storage) GetProfile(input string) (user models.User, err error) {
	err = s.db.QueryRow("SELECT fullname, email, about, nickname FROM users WHERE nickname = $1", input).
		Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return user, models.ServError{Code: 404}

		}
		return user, models.ServError{Code: 500}
	}

	return
}

func (s *storage) UpdateProfile(input models.User) (user models.User, err error) {
	if input.About != "" && input.Email != "" && input.FullName != "" {
		err = s.db.QueryRow(updateFull, input.Nickname, input.FullName, input.Email, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.About != "" && input.Email != "" {
		err = s.db.QueryRow(updateEmailAbout, input.Nickname, input.Email, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.Email != "" && input.FullName != "" {
		err = s.db.QueryRow(updateEmailFullname, input.Nickname, input.FullName, input.Email, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.About != "" && input.FullName != "" {
		err = s.db.QueryRow(updateFullnameAbout, input.Nickname, input.FullName, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.About != "" {
		err = s.db.QueryRow(updateAbout, input.Nickname, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.FullName != "" {
		err = s.db.QueryRow(updateFullname, input.Nickname, input.FullName, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.Email != "" {
		err = s.db.QueryRow(updateEmail, input.Nickname, input.Email, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	}

	if err == pgx.ErrNoRows {
		return user, models.ServError{Code: 404}
	}

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return user, models.ServError{Code: 409}
		default:
			return user, models.ServError{Code: 500}
		}
	}

	return
}

func (s *storage) GetUsers(input models.ForumQueryParams, forumID int) (users []models.User, err error) {
	var rows *pgx.Rows
	users = make([]models.User, 0)
	if input.Since == "" && !input.Desc {
		rows, err = s.db.Query(selectEmpty, forumID, input.Limit)
	} else if input.Since == "" && input.Desc {
		rows, err = s.db.Query(selectWithDesc, forumID, input.Limit)
	}  else if input.Since != "" && !input.Desc {
		rows, err = s.db.Query(selectWithSince, forumID, input.Since, input.Limit)
	} else if input.Since != "" && input.Desc {
		rows, err = s.db.Query(selectWithSinceDesc, forumID, input.Since, input.Limit)
	}

	if err != nil {
		fmt.Println(err)
		return users, models.ServError{Code: 500}
	}

	defer rows.Close()

	for rows.Next() {
		user := models.User{}

		err = rows.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
		if err != nil {
			return users, models.ServError{Code: 500}
		}

		users = append(users, user)
	}

	return
}

func (s *storage) GetUserForPost(input string, user *models.User) (err error) {
	user.Nickname = input
	err = s.db.QueryRow("SELECT fullname, email, about FROM users WHERE nickname = $1", input).
		Scan(&user.FullName, &user.Email, &user.About)

	if err != nil {
		return models.ServError{Code: 500}
	}

	return
}

func (s *storage) GetUserIDByNickname(input string) (userID int, err error) {
	err = s.db.QueryRow("SELECT ID FROM users WHERE nickname = $1", input).Scan(&userID)
	if err != nil {
		return userID, models.ServError{Code: 500}
	}

	return
}

func (s *storage) GetEmailConflictUser(email string) (user models.User, err error) {
	err = s.db.QueryRow("SELECT fullname, nickname, about, email FROM users WHERE email = $1", email).
		Scan(&user.FullName, &user.Nickname, &user.About, &user.Email)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return user, models.ServError{Code: 404}

		}
		return user, models.ServError{Code: 500}
	}

	return
}