package forum

import (
	"database/sql"
	"github.com/jackc/pgx"
	"github.com/pringleskate/tp_db_forum/internal/models"
)

func slugToNullable(slug string) sql.NullString {
	nullable := sql.NullString{
		String: slug,
		Valid:  true,
	}
	if slug == "" {
		nullable.Valid = false
	}

	return nullable
}

func scanThread(r *pgx.Row) (t models.Thread, err error) {
	slug := sql.NullString{}

	// ID, author, created, forum, message, slug, title, votes
	err = r.Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &slug, &t.Title, &t.Votes)
	if err != nil {
		return t, err
	}
	if slug.Valid {
		t.Slug = slug.String
	}
	return t, err
}

func scanThreadRows(r *pgx.Rows) (threads []models.Thread, err error) {
	threads = make([]models.Thread, 0)
	for r.Next() {
		thread := models.Thread{}
		slug := sql.NullString{}

		//ID, author, created, forum, message, slug, title, votes
		err = r.Scan(&thread.ID,&thread.Author, &thread.Created, &thread.Forum,&thread.Message, &slug, &thread.Title, &thread.Votes)
		if err != nil {
			return threads, err
		}

		if slug.Valid {
			thread.Slug = slug.String
		}

		threads = append(threads, thread)
	}

	return
}