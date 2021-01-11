package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/pringleskate/tp_db_forum/internal/models"
	"strconv"
	"strings"
)

type Storage struct {
	Db *pgx.ConnPool
}

func (s *Storage) ForumCreate(forumSlug models.Forum) (forum models.Forum, err error) {
	err = s.Db.QueryRow("INSERT INTO forums (slug, title, user_nick) VALUES ($1, $2,(SELECT u.nickname FROM users u WHERE u.nickname = $3)) RETURNING slug, title, user_nick",
		forumSlug.Slug, forumSlug.Title, forumSlug.User).Scan(&forum.Slug, &forum.Title, &forum.User)

	return
}

func (s *Storage) ForumSelect(forumSlug string) (forum models.Forum, err error) {
	err = s.Db.QueryRow("SELECT slug, title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug).
		Scan(&forum.Slug, &forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	return
}

func (s *Storage) ThreadsCountUpdate(input string) (err error) {
	_, err = s.Db.Exec("UPDATE forums SET threads = threads + 1 WHERE slug = $1", input)
	return
}

func (s *Storage) PostsCountUpdate(input string, posts int) (err error) {
	_, err = s.Db.Exec("UPDATE forums SET posts = posts + $2 WHERE slug = $1", input, posts)
	return
}

func (s *Storage) NewForumUser(userID int, forumID int) (err error) {
	_, err = s.Db.Exec("INSERT INTO forum_users (forumID, userID) VALUES ($1, $2)", forumID, userID)
	return
}

func (s *Storage) IfForumExists(input string) (err error) {
	var ID int
	err = s.Db.QueryRow("SELECT ID from forums WHERE slug = $1", input).Scan(&ID)
	return
}

func (s Storage) ForumIDSelect(input string) (ID int, err error) {
	err = s.Db.QueryRow("SELECT ID from forums WHERE slug = $1", input).Scan(&ID)
	return
}

func (s *Storage) ForumPostSelect(forumSlug string, forum *models.Forum) (err error) {
	forum.Slug = forumSlug
	err = s.Db.QueryRow("SELECT title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug).
		Scan(&forum.Title, &forum.Threads, &forum.Posts, &forum.User)
	return
}

func (s *Storage) ThreadCreate(input models.Thread) (thread models.Thread, err error) {
	query := `INSERT INTO threads (author, created, forum, message, slug, title, votes) VALUES 
				((SELECT u.nickname FROM users u WHERE u.nickname = $1), 
				$2, 
				(SELECT f.slug FROM forums f WHERE f.slug = $3), 
				$4, $5, $6, $7) 
				RETURNING ID, author, created, forum, message, slug, title, votes`
	return scanThread(s.Db.QueryRow(query, input.Author, input.Created, input.Forum, input.Message, slugToNullable(input.Slag), input.Title, input.Votes))
}

func (s *Storage) SelectThreadBySlug(slug string) (thread models.Thread, err error) {
	query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE slug = $1`
	return scanThread(s.Db.QueryRow(query, slug))

}

func (s *Storage) SelectThreadByID(ID int) (thread models.Thread, err error) {
	query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1`
	return scanThread(s.Db.QueryRow(query, ID))
}

func (s *Storage) ThreadEdit(input models.ThreadUpdate) (thread models.Thread, err error) {
	query := "UPDATE threads SET "
	selectQuery := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1 OR slug = $2`
	if input.Title == "" && input.Message == "" {
		return scanThread(s.Db.QueryRow(selectQuery, input.ThreadID, input.ThreadSlug))
	}

	if input.Title != "" && input.Message != "" {
		query += fmt.Sprintf("title = '%s',", input.Title)
	} else if input.Title != "" {
		query += fmt.Sprintf("title = '%s'", input.Title)
	}
	if input.Message != "" {
		query += fmt.Sprintf("message = '%s'", input.Message)
	}

	query += fmt.Sprintf(" WHERE ID = $1 OR slug = $2 RETURNING ID, author, created, forum, message, slug, title, votes")

	return scanThread(s.Db.QueryRow(query, input.ThreadID, input.ThreadSlug))
}

func (s *Storage) SelectAllThreadsByForum(input models.ForumQueryParams) (threads []models.Thread, err error) {
	var rows *pgx.Rows

	if input.Since == "" && !input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 ORDER BY created LIMIT $2`
		rows, err = s.Db.Query(query, input.Slug, input.Limit)
	} else if input.Since == "" && input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2`
		rows, err = s.Db.Query(query, input.Slug, input.Limit)
	} else if input.Since != "" && !input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3`
		rows, err = s.Db.Query(query, input.Slug, input.Since, input.Limit)
	} else if input.Since != "" && input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3`
		rows, err = s.Db.Query(query, input.Slug, input.Since, input.Limit)
	}

	if err != nil {
		return threads, err
	}

	return scanThreadRows(rows)
}

func (s Storage) IfThreadExists(input models.ThreadSlagOrID) (thread models.ThreadSlagOrID, err error) {
	query := `SELECT ID from threads WHERE ID = $1 OR slug = $2`
	err = s.Db.QueryRow(query, input.ThreadID, input.ThreadSlug).Scan(&thread.ThreadID)
	return
}

func (s *Storage) ThreadPostSelect(input models.ThreadSlagOrID, thread *models.Thread) (err error) {
	slug := sql.NullString{}
	query := `SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE ID = $1`
	err = s.Db.QueryRow(query, input.ThreadID).
		Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &slug, &thread.Title, &thread.Votes)

	if err != nil {
		return models.ServError{Code: 500}
	}

	if slug.Valid {
		thread.Slag = slug.String
	}

	return
}

func (s *Storage) SelectForumByThread(input *models.ThreadSlagOrID) (forum string, err error) {
	query := `SELECT forum, ID FROM threads WHERE ID = $1 OR slug = $2`
	err = s.Db.QueryRow(query, input.ThreadID, input.ThreadSlug).Scan(&forum, &input.ThreadID)
	return
}

func (s *Storage) DoVote(vote models.Vote, update bool) (thread models.Thread, err error) {
	var boolVoice bool
	if vote.Voice == 1 {
		boolVoice = true
	}

	tx, err := s.Db.Begin()
	if err != nil {
		return thread, err
	}

	_, err = tx.Exec("SET LOCAL synchronous_commit TO OFF")
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return thread, err
		}
		return thread, err
	}

	insertVote := "INSERT INTO votes (user_nick, voice, thread) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT uniq_votes DO UPDATE SET voice = EXCLUDED.voice;"

	_, err = tx.Exec(insertVote, vote.Nickname, boolVoice, vote.ThreadSlagOrID.ThreadID)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return thread, err
		}
		return thread, err
	}

	slug := sql.NullString{}

	queryUp := `UPDATE threads SET votes = votes + $2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes`
	queryDown := `UPDATE threads SET votes = votes - $2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes`

	delta := 1
	if update {
		delta = 2
	}

	if boolVoice {
		err = tx.QueryRow(queryUp, vote.ThreadSlagOrID.ThreadID, delta).
			Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
	} else {
		err = tx.QueryRow(queryDown, vote.ThreadSlagOrID.ThreadID, delta).
			Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
	}

	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return thread, err
		}
		return thread, err
	}

	if slug.Valid {
		thread.Slag = slug.String
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return thread, err
	}

	return
}

func (s *Storage) CheckIfUserVoted(vote models.Vote) (thread models.Thread, nonConflictVice bool, err error) {
	var boolVoice bool
	if vote.Voice == 1 {
		boolVoice = true
	}

	var oldVoice bool
	queryVoice := `SELECT voice FROM votes WHERE user_nick = $1 AND thread = $2`
	err = s.Db.QueryRow(queryVoice, vote.Nickname, vote.ThreadSlagOrID.ThreadID).Scan(&oldVoice)
	if err != nil {
		if err == pgx.ErrNoRows {
			return thread,false,  nil
		}
		return thread, false, models.ServError{Code: 500}
	}

	if oldVoice != boolVoice {
		return thread, true, models.ServError{Code: 101}
	}

	queryThread := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1`
	thread, err = scanThread(s.Db.QueryRow(queryThread, vote.ThreadSlagOrID.ThreadID))

	if err != nil {
		return thread, false, models.ServError{Code: 500}
	}

	return thread, false, models.ServError{Code: 409}
}

func (s Storage) PostsCreate(thread models.ThreadSlagOrID, forum string, created string, posts []models.PostCreate) (post []models.Post, err error) {
	sqlStr := "INSERT INTO posts(id, parent, thread, forum, author, created, message, path) VALUES "
	vals := []interface{}{}
	for _, post := range posts {
		var authorID int
		err = s.Db.QueryRow(`SELECT id FROM users WHERE nickname = $1`,
			post.Author,
		).Scan(&authorID)
		if err != nil {
			return nil, models.ServError{Code: 404, Message: "cannot find user"}
		}

		var forumID int
		err = s.Db.QueryRow(`SELECT id FROM forums WHERE slug = $1`,
			forum,
		).Scan(&forumID)
		if err != nil {
			return nil, models.ServError{Code: 404, Message: "cannot find thread"}
		}

		sqlQuery := `
		INSERT INTO forum_users (forumID, userID)
		VALUES ($1,$2)`
		_, err = s.Db.Exec(sqlQuery, forumID, authorID)
		if err != nil {
			if pqErr, ok := err.(pgx.PgError); ok {
				if pqErr.Code != pgerrcode.UniqueViolation {
					return nil, errors.New("500")
				}
			}
		}

		if post.Parent == 0 {
			sqlStr += "(nextval('post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"ARRAY[currval(pg_get_serial_sequence('posts', 'id'))::INTEGER]),"
			vals = append(vals, post.Parent, thread.ThreadID, forum, post.Author, created, post.Message)
		} else {
			var parentThreadId int32
			err = s.Db.QueryRow("SELECT thread FROM posts WHERE id = $1",
				post.Parent,
			).Scan(&parentThreadId)
			if err != nil {
				return nil, models.ServError{Code: 409, Message: "Parent post was created in another thread"}
			}
			if parentThreadId != int32(thread.ThreadID) {
				return nil, models.ServError{Code: 409, Message: "Parent post was created in another thread"}
			}

			sqlStr += " (nextval('post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"(SELECT posts.path FROM posts WHERE posts.id = ? AND posts.thread = ?) || " +
				"currval(pg_get_serial_sequence('posts', 'id'))::INTEGER),"

			vals = append(vals, post.Parent, thread.ThreadID, forum, post.Author, created, post.Message, post.Parent, thread.ThreadID)
		}

	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	sqlStr += " RETURNING id, parent, thread, forum, author, created, message, edited "

	sqlStr = ReplaceSQL(sqlStr, "?")
	if len(posts) > 0 {
		rows, err := s.Db.Query(sqlStr, vals...)
		if err != nil {
			return nil, err
		}
		scanPost := models.Post{}
		for rows.Next() {
			err := rows.Scan(
				&scanPost.ID,
				&scanPost.Parent,
				&scanPost.ThreadID,
				&scanPost.Forum,
				&scanPost.Author,
				&scanPost.Created,
				&scanPost.Message,
				&scanPost.IsEdited,
			)
			if err != nil {
				rows.Close()
				return nil, err
			}
			post = append(post, scanPost)
		}
		rows.Close()
	}
	return post, nil
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

func (s *Storage) SelectPost(input int, post *models.Post) (err error) {
	query := `SELECT author, created, forum, message, ID , edited, parent, thread FROM posts WHERE ID = $1`
	err = s.Db.QueryRow(query, input).
		Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadSlagOrID.ThreadID)
	return
}

func (s *Storage) PostEdit(input models.PostUpdate) (post models.Post, err error) {
	var oldMessage string
	err = s.Db.QueryRow("SELECT message FROM posts WHERE ID = $1", input.ID).
		Scan(&oldMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return post, models.ServError{Code: 404}
		}
		return post, models.ServError{Code: 500}
	}

	if input.Message != "" && input.Message != oldMessage {
		err = s.Db.QueryRow("UPDATE posts SET message = $1, edited = $2 WHERE ID = $3 RETURNING author, created, forum, message, ID , edited, parent, thread", input.Message, true, input.ID).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadSlagOrID.ThreadID)
	} else {
		err = s.Db.QueryRow("SELECT author, created, forum, message, ID , edited, parent, thread FROM posts WHERE ID = $1", input.ID).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadSlagOrID.ThreadID)
	}

	if err != nil {
		return post, models.ServError{Code: 500}
	}
	return
}

func (s *Storage) SelectAllPostsByThread(input models.ThreadQueryParams) (posts []models.Post, err error){
	var rows *pgx.Rows
	posts  = make([]models.Post, 0)
	var query string

	switch input.Sort {
	case "flat":
		if input.Since > 0 {
			if input.Desc {
				query = ` SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and p.id < $2
							ORDER BY p.created DESC, p.id DESC LIMIT $3`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and p.id > $2 
							ORDER BY p.created, p.id LIMIT $3`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc == true {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1
							ORDER BY p.created DESC, p.id DESC LIMIT $2`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 ORDER BY p.created, p.id LIMIT $2`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.Limit)
			}
		}
	case "tree":
		if input.Since > 0 {
			if input.Desc {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and (p.path < (SELECT p2.path from posts p2 where p2.id = $2))
							ORDER BY p.path DESC LIMIT $3`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and (p.path > (SELECT p2.path from posts p2 where p2.id = $2))
							ORDER BY p.path LIMIT $3`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 ORDER BY path DESC LIMIT $2`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 ORDER BY p.path LIMIT $2`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.Limit)
			}
		}
	case "parent_tree":
		if input.Since > 0 {
			if input.Desc {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 
							and p.path[1] IN (
							SELECT p2.path[1] FROM posts p2
							WHERE p2.thread = $2 AND p2.parent = 0 and p2.path[1] < (SELECT p3.path[1] from posts p3 where p3.id = $3)
							ORDER BY p2.path DESC LIMIT $4) ORDER BY p.path[1] DESC, p.path[2:]`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 
							and p.path[1] IN (
							SELECT p2.path[1]
							FROM posts p2 WHERE p2.thread = $2 AND p2.parent = 0 and p2.path[1] > (SELECT p3.path[1] from posts p3 where p3.id = $3)
							ORDER BY p2.path LIMIT $4) ORDER BY p.path`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and p.path[1] IN (
							SELECT p2.path[1] FROM posts p2 WHERE p2.parent = 0 and p2.thread = $2
							ORDER BY p2.path DESC LIMIT $3) ORDER BY p.path[1] DESC, p.path[2:]`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.ThreadSlagOrID.ThreadID,
					input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and p.path[1] IN (
							SELECT p2.path[1] FROM posts p2 WHERE p2.thread = $2 AND p2.parent = 0
							ORDER BY p2.path LIMIT $3) ORDER BY path`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.ThreadSlagOrID.ThreadID,
					input.Limit)
			}
		}
	default:
		if input.Since > 0 {
			if input.Desc {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and p.id < $2
							ORDER BY p.created DESC, p.id DESC LIMIT $3`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 and p.id > $2
							ORDER BY p.created, p.id LIMIT $3`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc == true {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 ORDER BY p.created DESC, p.id DESC LIMIT $2`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.Limit)
			} else {
				query = `SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
							FROM posts p WHERE p.thread = $1 ORDER BY p.created, p.id LIMIT $2`
				rows, err = s.Db.Query(query, input.ThreadSlagOrID.ThreadID, input.Limit)
			}
		}
	}

	if err != nil {
		return posts, models.ServError{Code: 500}
	}
	defer rows.Close()

	if rows == nil {
		return posts, models.ServError{Code: 500}
	}

	for rows.Next() {
		post := models.Post{}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.IsEdited, &post.Message, &post.Parent, &post.ThreadSlagOrID.ThreadID, &post.Forum)
		if err != nil {
			return posts, models.ServError{Code: 500}
		}

		posts = append(posts, post)
	}

	return
}

func (s Storage) SelectPostsThread(post int) (thread int, err error) {
	query := `SELECT thread FROM posts WHERE ID = $1`
	err = s.Db.QueryRow(query, post).Scan(&thread)
	return
}