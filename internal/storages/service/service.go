package service

import (
	"github.com/jackc/pgx"
	"github.com/pringleskate/tp_db_forum/internal/models"
)

type Service struct {
	Db *pgx.ConnPool
}

func (s *Service) Clear() (err error) {
	query := `TRUNCATE users, forums, threads, posts, forum_users, votes CASCADE`
	_, err = s.Db.Exec(query)
	if err != nil {
		return models.ServError{Code: 500}
	}
	return
}

func (s *Service) Status() (status models.Status, err error) {
	query := `SELECT (SELECT COUNT(*) FROM forums), (SELECT COUNT(*) FROM threads), (SELECT COUNT(*) FROM posts), (SELECT COUNT(*) FROM users)`
	err = s.Db.QueryRow(query).Scan(&status.Forum, &status.Thread, &status.Post, &status.User)
	if err != nil && err != pgx.ErrNoRows {
		return status, models.ServError{Code: 500}
	}

	return
}
