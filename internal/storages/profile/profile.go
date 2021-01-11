package profile

import (
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/pringleskate/tp_db_forum/internal/models"
)

type Storage struct {
	Db *pgx.ConnPool
}

func (s *Storage) SelectUserID(input string) (userID int, err error) {
	query := `SELECT ID FROM users WHERE nickname = $1`
	err = s.Db.QueryRow(query, input).Scan(&userID)
	if err != nil {
		return userID, models.ServError{Code: 500}
	}

	return
}

func (s *Storage) SelectUserByEmail(email string) (user models.User, err error) {
	query := `SELECT fullname, nickname, about, email FROM users WHERE email = $1`
	err = s.Db.QueryRow(query, email).
		Scan(&user.FullName, &user.Nickname, &user.About, &user.Email)

	if err != nil {
		if err == pgx.ErrNoRows {
			return user, models.ServError{Code: 404}

		}
		return user, models.ServError{Code: 500}
	}

	return
}

func (s *Storage) SelectFullUser(input string) (user models.User, err error) {
	query := `SELECT fullname, email, about, nickname FROM users WHERE nickname = $1`
	err = s.Db.QueryRow(query, input).Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)

	if err != nil {
		if err == pgx.ErrNoRows {
			return user, models.ServError{Code: 404}

		}
		return user, models.ServError{Code: 500}
	}

	return
}

func (s *Storage) UserCreate(input models.User) (user models.User, err error) {
	query := `INSERT INTO users (nickname, email, fullname, about) VALUES ($1, $2, $3, $4)`
	_, err = s.Db.Exec(query, input.Nickname, input.Email, input.FullName, input.About)

	if pqErr, ok := err.(pgx.PgError); ok {
		fmt.Println(err)
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			fmt.Println(pqErr.Code)
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

func (s *Storage) EditUser(input models.User) (user models.User, err error) {
	var query string
	if input.About != "" && input.Email != "" && input.FullName != "" {
		query = "UPDATE users SET nickname = $1, fullname = $2, email = $3, about = $4 WHERE nickname = $5 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.FullName, input.Email, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.About != "" && input.Email != "" {
		query = "UPDATE users SET nickname = $1, email = $2, about = $3 WHERE nickname = $4 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.Email, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.Email != "" && input.FullName != "" {
		query = "UPDATE users SET nickname = $1, fullname = $2, email = $3 WHERE nickname = $4 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.FullName, input.Email, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.About != "" && input.FullName != "" {
		query = "UPDATE users SET nickname = $1, fullname = $2, about = $3 WHERE nickname = $4 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.FullName, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.About != "" {
		query = "UPDATE users SET nickname = $1, about = $2 WHERE nickname = $3 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.About, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.FullName != "" {
		query = "UPDATE users SET nickname = $1, fullname = $2 WHERE nickname = $3 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.FullName, input.Nickname).
			Scan(&user.FullName, &user.Email, &user.About, &user.Nickname)
	} else if input.Email != "" {
		query = "UPDATE users SET nickname = $1, email = $2 WHERE nickname = $3 RETURNING fullname, email, about, nickname"
		err = s.Db.QueryRow(query, input.Nickname, input.Email, input.Nickname).
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

func (s *Storage) SelectAllUsers(input models.ForumQueryParams, forumID int) (users []models.User, err error) {
	var rows *pgx.Rows
	users = make([]models.User, 0)

	var query string
	if input.Since == "" && !input.Desc {
		query = "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 ORDER BY u.nickname LIMIT $2"
		rows, err = s.Db.Query(query, forumID, input.Limit)
	} else if input.Since == "" && input.Desc {
		query = "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 ORDER BY u.nickname DESC LIMIT $2"
		rows, err = s.Db.Query(query, forumID, input.Limit)
	}  else if input.Since != "" && !input.Desc {
		query = "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 AND u.nickname > $2 ORDER BY u.nickname LIMIT $3"
		rows, err = s.Db.Query(query, forumID, input.Since, input.Limit)
	} else if input.Since != "" && input.Desc {
		query =  "SELECT u.nickname, u.fullname, u.about, u.email FROM forum_users fu JOIN users u ON fu.userID = u.ID WHERE fu.forumID = $1 AND u.nickname < $2 ORDER BY u.nickname DESC LIMIT $3"
		rows, err = s.Db.Query(query, forumID, input.Since, input.Limit)
	}

	if err != nil {
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

func (s *Storage) UserPostSelect(input string, user *models.User) (err error) {
	user.Nickname = input
	query := "SELECT fullname, email, about FROM users WHERE nickname = $1"
	err = s.Db.QueryRow(query, input).
		Scan(&user.FullName, &user.Email, &user.About)

	if err != nil {
		return models.ServError{Code: 500}
	}

	return
}
