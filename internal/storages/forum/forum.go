package forum

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/pringleskate/tp_db_forum/internal/models"
)

type Storage interface {
	CreateForum(forumSlug models.ForumCreate) (forum models.Forum, err error)
	GetForumDetails(forumSlug models.ForumInput) (forum models.Forum, err error)
	UpdateThreadsCount(input models.ForumInput) (err error)
	UpdatePostsCount(input models.ForumInput, posts int) (err error)
	AddUserToForum(userID int, forumID int) (err error)
	CheckIfForumExists(input models.ForumInput) (err error)
	GetForumID(input models.ForumInput) (ID int, err error)
	GetForumForPost(forumSlug string, forum *models.Forum) (err error)

	CreateThread(input models.Thread) (thread models.Thread, err error)
	GetThreadDetails(input models.ThreadInput) (thread models.Thread, err error)
	ThreadEdit(input models.ThreadUpdate) (thread models.Thread, err error)
	GetThreadsByForum(input models.ForumGetThreads) (threads []models.Thread, err error)
	CheckThreadIfExists(input models.ThreadInput) (thread models.ThreadInput, err error)
	GetThreadForPost(input models.ThreadInput, post *models.Thread) (err error)
	GetForumByThread(input *models.ThreadInput) (forum string, err error)

	CreateVote(vote models.Vote, update bool) (thread models.Thread, err error)
	CheckDoubleVote(vote models.Vote) (thread models.Thread, err error)
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

func (s *storage) CreateForum(forumSlug models.ForumCreate) (forum models.Forum, err error) {
	err = s.db.QueryRow("INSERT INTO forums (slug, title, user_nick) VALUES ($1, $2,(SELECT u.nickname FROM users u WHERE u.nickname = $3)) RETURNING slug, title, user_nick",
		forumSlug.Slug, forumSlug.Title, forumSlug.User).Scan(&forum.Slug, &forum.Title, &forum.User)

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return forum, models.Error{Code: "409"}
		case pgerrcode.NotNullViolation, pgerrcode.ForeignKeyViolation:
			return forum, models.Error{Code: "404"}
		default:
			fmt.Println(err)
			return forum, models.Error{Code: "500"}
		}
	}

	return forum, nil
}

func (s *storage) GetForumDetails(forumSlug models.ForumInput) (forum models.Forum, err error) {
	err = s.db.QueryRow("SELECT slug, title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug.Slug).
		Scan(&forum.Slug, &forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return forum, models.Error{Code: "404"}

		}
		return forum, models.Error{Code: "500"}
	}

	return forum, nil
}

func (s *storage) UpdateThreadsCount(input models.ForumInput) (err error) {
	_, err = s.db.Exec("UPDATE forums SET threads = threads + 1 WHERE slug = $1", input.Slug)
	if err != nil {
		fmt.Println(err)
		return models.Error{Code: "500"}
	}
	return
}

func (s *storage) UpdatePostsCount(input models.ForumInput, posts int) (err error) {
	_, err = s.db.Exec("UPDATE forums SET posts = posts + $2 WHERE slug = $1", input.Slug, posts)
	if err != nil {
		fmt.Println(err)
		return models.Error{Code: "500"}
	}
	return
}

func (s *storage) AddUserToForum(userID int, forumID int) (err error) {
	_, err = s.db.Exec("INSERT INTO forum_users (forumID, userID) VALUES ($1, $2)", forumID, userID)
	if err != nil {
		if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.UniqueViolation:
				return models.Error{Code: "409"}
			}
		}
		return models.Error{Code: "500"}
	}

	return
}

func (s *storage) CheckIfForumExists(input models.ForumInput) (err error) {
	var ID int
	err = s.db.QueryRow("SELECT ID from forums WHERE slug = $1", input.Slug).Scan(&ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Error{Code: "404"}
		}
		return models.Error{Code: "500"}
	}

	return
}

func (s storage) GetForumID(input models.ForumInput) (ID int, err error) {
	err = s.db.QueryRow("SELECT ID from forums WHERE slug = $1", input.Slug).Scan(&ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ID, models.Error{Code: "404"}
		}
		return ID, models.Error{Code: "500"}
	}

	return
}

func (s *storage) GetForumForPost(forumSlug string, forum *models.Forum) (err error) {
	forum.Slug = forumSlug
	err = s.db.QueryRow("SELECT title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug).
		Scan(&forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	if err != nil {
		return models.Error{Code: "500"}
	}

	return
}

var (
	insertWithSlug = "INSERT INTO threads (author, created, forum, message, slug, title, votes) VALUES ((SELECT u.nickname FROM users u WHERE u.nickname = $1), $2, (SELECT f.slug FROM forums f WHERE f.slug = $3), $4, $5, $6, $7) RETURNING ID, author, created, forum, message, slug, title, votes"
	insertWithoutSlug = "INSERT INTO threads (author, created, forum, message, title, votes) VALUES ((SELECT u.nickname FROM users u WHERE u.nickname = $1), $2, (SELECT f.slug FROM forums f WHERE f.slug = $3), $4, $5, $6) RETURNING ID, author, created, forum, message, title, votes"

	selectBySlug = "SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE slug = $1"
	selectByID = "SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE ID = $1"

	selectThreads = "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 ORDER BY created LIMIT $2"
	selectThreadsSince = "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3"
	selectThreadsDesc = "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2"
	selectThreadsSinceDesc =  "SELECT id, slug, author, created, forum, title, message, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3"
)

func (s *storage) CreateThread(input models.Thread) (thread models.Thread, err error) {
	if input.Slug == "" {
		err = s.db.QueryRow(insertWithoutSlug, input.Author, input.Created, input.Forum, input.Message, input.Title, input.Votes).
			Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Title, &thread.Votes)
	} else {
		err = s.db.QueryRow(insertWithSlug, input.Author, input.Created, input.Forum, input.Message, input.Slug, input.Title, input.Votes).
			Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	}

	if pqErr, ok := err.(pgx.PgError); ok {
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return thread, models.Error{Code: "409"}
		case pgerrcode.NotNullViolation, pgerrcode.ForeignKeyViolation:
			return thread, models.Error{Code: "404"}
		default:
			return thread, models.Error{Code: "500"}
		}
	}

	return
}

func (s *storage) GetThreadDetails(input models.ThreadInput) (thread models.Thread, err error) {
	slug := sql.NullString{}
	if input.Slug == "" {
		err = s.db.QueryRow(selectByID, input.ThreadID).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)
	} else {
		err = s.db.QueryRow(selectBySlug, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, models.Error{Code: "404"}

		}
		return thread, models.Error{Code: "500"}
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	return
}


//TODO sqlbuilder (fmt.Sprintf())
func (s *storage) ThreadEdit(input models.ThreadUpdate) (thread models.Thread, err error) {
	if input.Title != "" && input.Message != "" {
		err = s.db.QueryRow("UPDATE threads SET message = $1, title = $2 WHERE ID = $3 OR slug = $4 " +
			"RETURNING author, created, forum, ID, message, slug, title, votes",
			input.Message, input.Title, input.ThreadID, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	} else if input.Title != "" && input.Message == "" {
		err = s.db.QueryRow("UPDATE threads SET title = $1 WHERE ID = $2 OR slug = $3 " +
			"RETURNING author, created, forum, ID, message, slug, title, votes",
			input.Title, input.ThreadID, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	} else if input.Title == "" && input.Message != "" {
		err = s.db.QueryRow("UPDATE threads SET message = $1 WHERE ID = $2 OR slug = $3 " +
			"RETURNING author, created, forum, ID, message, slug, title, votes",
			input.Message, input.ThreadID, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)


	} else if input.Title == "" && input.Message == "" {
		err = s.db.QueryRow("SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE ID = $1 OR slug = $2", input.ThreadID, input.Slug).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	}

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return thread, models.Error{Code: "404"}

		}
		return thread, models.Error{Code: "500"}
	}

	return
}

func (s *storage) GetThreadsByForum(input models.ForumGetThreads) (threads []models.Thread, err error) {
	var rows *pgx.Rows
	if input.Since == "" && !input.Desc {
		rows, err = s.db.Query(selectThreads, input.Slug, input.Limit)
	} else if input.Since == "" && input.Desc {
		rows, err = s.db.Query(selectThreadsDesc,  input.Slug, input.Limit)
	}  else if input.Since != "" && !input.Desc {
		rows, err = s.db.Query(selectThreadsSince,  input.Slug, input.Since, input.Limit)
	} else if input.Since != "" && input.Desc {
		rows, err = s.db.Query(selectThreadsSinceDesc, input.Slug, input.Since, input.Limit)
	}

	if err != nil {
		return threads, models.Error{Code: "500"}
	}
	defer rows.Close()

	threads = make([]models.Thread, 0)
	for rows.Next() {
		thread := models.Thread{}
		slug := sql.NullString{}

		err = rows.Scan(&thread.ID, &slug, &thread.Author, &thread.Created, &thread.Forum, &thread.Title, &thread.Message, &thread.Votes)
		if err != nil {
			return threads, models.Error{Code: "500"}
		}

		if slug.Valid {
			thread.Slug = slug.String
		}

		threads = append(threads, thread)
	}

	return
}

func (s storage) CheckThreadIfExists(input models.ThreadInput) (thread models.ThreadInput, err error) {
	if input.Slug == "" {
		err = s.db.QueryRow("SELECT ID from threads WHERE ID = $1", input.ThreadID).Scan(&thread.ThreadID)
	} else {
		err = s.db.QueryRow("SELECT ID from threads WHERE slug = $1", input.Slug).Scan(&thread.ThreadID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, models.Error{Code: "404"}
		}
		return thread, models.Error{Code: "500"}
	}

	return
}

func (s *storage) GetThreadForPost(input models.ThreadInput, thread *models.Thread) (err error) {
	slug := sql.NullString{}
	err = s.db.QueryRow(selectByID, input.ThreadID).
		Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)

	if err != nil {
		return models.Error{Code: "500"}
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	return
}

func (s *storage) GetForumByThread(input *models.ThreadInput) (forum string, err error) {
	if input.Slug == "" {
		err = s.db.QueryRow("SELECT forum FROM threads WHERE ID = $1", input.ThreadID).Scan(&forum)
	} else {
		err = s.db.QueryRow("SELECT forum, ID FROM threads WHERE slug = $1", input.Slug).Scan(&forum, &input.ThreadID)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return forum, models.Error{Code: "404"}
		}

		return forum, models.Error{Code: "500"}
	}

	return
}

var (
	insertVote            = "INSERT INTO votes (user_nick, voice, thread) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT uniq_votes DO UPDATE SET voice = EXCLUDED.voice;"
	createThreadVotesUp   = "UPDATE threads SET votes = votes + 1 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"
	createThreadVotesDown = "UPDATE threads SET votes = votes - 1 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"

	updateThreadVotesUp   = "UPDATE threads SET votes = votes + 2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"
	updateThreadVotesDown = "UPDATE threads SET votes = votes - 2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes"
)

func (s *storage) CreateVote(vote models.Vote, update bool) (thread models.Thread, err error) {
	boolVoice := getBoolVoice(vote)
	tx, err := s.db.Begin()
	if err != nil {
		fmt.Println("txerr", err)
		return thread, models.Error{Code: "500"}
	}

	_, err = tx.Exec("SET LOCAL synchronous_commit TO OFF")
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			fmt.Println(txErr)
			return thread, models.Error{Code: "500"}
		}
		return thread, models.Error{Code: "500"}
	}

	_, err = tx.Exec(insertVote, vote.User, boolVoice, vote.Thread.ThreadID)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			fmt.Println(txErr)
			return thread, models.Error{Code: "500"}
		}
		if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return thread, models.Error{Code: "404"}
			default:
				return thread, models.Error{Code: "500"}
			}
		}
		return thread, models.Error{Code: "500"}
	}

	slug := sql.NullString{}

	if update {
		if boolVoice {
			err = tx.QueryRow(updateThreadVotesUp, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		} else {
			err = tx.QueryRow(updateThreadVotesDown, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		}
	} else {
		if boolVoice {
			err = tx.QueryRow(createThreadVotesUp, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		} else {
			err = tx.QueryRow(createThreadVotesDown, vote.Thread.ThreadID).
				Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
		}
	}

	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			fmt.Println(txErr)
			return thread, models.Error{Code: "500"}
		}

		return thread, models.Error{Code: "500"}
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	if commitErr := tx.Commit(); commitErr != nil {
		fmt.Println(commitErr)
		return thread, models.Error{Code: "500"}
	}

	return
}

func getBoolVoice(vote models.Vote) bool {
	if vote.Voice == 1 {
		return true
	}
	return false
}

func (s *storage) CheckDoubleVote(vote models.Vote) (thread models.Thread, err error) {
	boolVoice := getBoolVoice(vote)
	var oldVoice bool
	err = s.db.QueryRow("SELECT voice FROM votes WHERE user_nick = $1 AND thread = $2", vote.User, vote.Thread.ThreadID).
		Scan(&oldVoice)
	if err != nil {
		if err == pgx.ErrNoRows {
			return thread, nil
		}
		return thread, models.Error{Code: "500"}
	}

	if oldVoice != boolVoice {
		return thread, models.Error{Code: "101"}
	}

	slug := sql.NullString{}
	err = s.db.QueryRow("SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1", vote.Thread.ThreadID).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)

	if slug.Valid {
		thread.Slug = slug.String
	}

	if err != nil {
		return thread, models.Error{Code: "500"}
	}

	return thread, models.Error{Code: "409"}
}

