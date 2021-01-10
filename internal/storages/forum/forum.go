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
	//GetThreadDetails(input models.ThreadInput) (thread models.Thread, err error)
	ThreadEdit(input models.ThreadUpdate) (thread models.Thread, err error)
	GetThreadsByForum(input models.ForumGetThreads) (threads []models.Thread, err error)
	CheckThreadIfExists(input models.ThreadInput) (thread models.ThreadInput, err error)
	GetThreadForPost(input models.ThreadInput, post *models.Thread) (err error)
	GetForumByThread(input *models.ThreadInput) (forum string, err error)

	CreatePosts(thread models.ThreadInput, forum string, created string, posts []models.PostCreate) (post []models.Post, err error)
	CreatePost(input models.Post) (post models.Post, err error)
	GetPostDetails(input models.PostInput, post *models.Post) (err error)
	UpdatePost(input models.PostUpdate) (post models.Post, err error)
	GetPostsByThread(input models.ThreadGetPosts) (posts []models.Post, err error)
	CheckParentPostThread(post int) (thread int, err error)

	CreateVote(vote models.Vote, update bool) (thread models.Thread, err error)
	CheckDoubleVote(vote models.Vote) (thread models.Thread, nonConflictVoice bool, err error)
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

	return
/*	if pqErr, ok := err.(pgx.PgError); ok {
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

	return forum, nil*/
}

func (s *storage) GetForumDetails(forumSlug models.ForumInput) (forum models.Forum, err error) {
	err = s.db.QueryRow("SELECT slug, title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug.Slug).
		Scan(&forum.Slug, &forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	return
	/*if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return forum, models.Error{Code: "404"}

		}
		return forum, models.Error{Code: "500"}
	}

	return forum, nil*/
}

func (s *storage) UpdateThreadsCount(input models.ForumInput) (err error) {
	_, err = s.db.Exec("UPDATE forums SET threads = threads + 1 WHERE slug = $1", input.Slug)
	return
/*	if err != nil {
		fmt.Println(err)
		return models.Error{Code: "500"}
	}
	return*/
}

func (s *storage) UpdatePostsCount(input models.ForumInput, posts int) (err error) {
	_, err = s.db.Exec("UPDATE forums SET posts = posts + $2 WHERE slug = $1", input.Slug, posts)
	return
/*	if err != nil {
		fmt.Println(err)
		return models.Error{Code: "500"}
	}
	return*/
}

func (s *storage) AddUserToForum(userID int, forumID int) (err error) {
	_, err = s.db.Exec("INSERT INTO forum_users (forumID, userID) VALUES ($1, $2)", forumID, userID)
/*	if err != nil {
		if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.UniqueViolation:
				return models.Error{Code: "409"}
			}
		}
		return models.Error{Code: "500"}
	}
*/
	return
}

func (s *storage) CheckIfForumExists(input models.ForumInput) (err error) {
	var ID int
	err = s.db.QueryRow("SELECT ID from forums WHERE slug = $1", input.Slug).Scan(&ID)
	/*if err != nil {
		if err == pgx.ErrNoRows {
			return models.Error{Code: "404"}
		}
		return models.Error{Code: "500"}
	}
*/
	return
}

func (s storage) GetForumID(input models.ForumInput) (ID int, err error) {
	err = s.db.QueryRow("SELECT ID from forums WHERE slug = $1", input.Slug).Scan(&ID)
	/*if err != nil {
		if err == pgx.ErrNoRows {
			return ID, models.Error{Code: "404"}
		}
		return ID, models.Error{Code: "500"}
	}
*/
	return
}

func (s *storage) GetForumForPost(forumSlug string, forum *models.Forum) (err error) {
	forum.Slug = forumSlug
	err = s.db.QueryRow("SELECT title, threads, posts, user_nick FROM forums WHERE slug = $1", forumSlug).
		Scan(&forum.Title, &forum.Threads, &forum.Posts, &forum.User)

	/*if err != nil {
		return models.Error{Code: "500"}
	}
*/
	return
}

func (s *storage) CreateThread(input models.Thread) (thread models.Thread, err error) {
	query := `INSERT INTO threads (author, created, forum, message, slug, title, votes) VALUES 
				((SELECT u.nickname FROM users u WHERE u.nickname = $1), 
				$2, 
				(SELECT f.slug FROM forums f WHERE f.slug = $3), 
				$4, $5, $6, $7) 
				RETURNING ID, author, created, forum, message, slug, title, votes`
	return scanThread(s.db.QueryRow(query, input.Author, input.Created, input.Forum, input.Message, slugToNullable(input.Slug), input.Title, input.Votes))
}

//TODO 	ID, author, created, forum, message, slug, title, votes
func (s *storage) SelectThreadBySlug(slug string) (thread models.Thread, err error) {
	query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE slug = $1`
	return scanThread(s.db.QueryRow(query, slug))

}

func (s *storage) SelectThreadByID(ID int) (thread models.Thread, err error) {
	query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1`
	return scanThread(s.db.QueryRow(query, ID))
}

/*func (s *storage) GetThreadDetails(input models.ThreadInput) (thread models.Thread, err error) {
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
*/
func (s *storage) ThreadEdit(input models.ThreadUpdate) (thread models.Thread, err error) {
	query := "UPDATE threads SET "
	selectQuery := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1 OR slug = $2`
	if input.Title == "" && input.Message == "" {
		return scanThread(s.db.QueryRow(selectQuery, input))
	}

	if input.Title != "" && input.Message != "" {
		query += fmt.Sprintf("title = %s,", input.Title)
	} else if input.Title != "" {
		query += fmt.Sprintf("title = %s", input.Title)
	}
	if input.Message != "" {
		query += fmt.Sprintf("message = %s", input.Message)
	}

	query += fmt.Sprintf("WHERE ID = $1 OR slug = $2 RETURNING ID, author, created, forum, message, slug, title, votes")

	return scanThread(s.db.QueryRow(query, input.ThreadID, input.Slug))
}

func (s *storage) GetThreadsByForum(input models.ForumGetThreads) (threads []models.Thread, err error) {
	var rows *pgx.Rows
	if input.Since == "" && !input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 ORDER BY created LIMIT $2`
		rows, err = s.db.Query(query, input.Slug, input.Limit)
	} else if input.Since == "" && input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3`
		rows, err = s.db.Query(query, input.Slug, input.Limit)
	}  else if input.Since != "" && !input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2`
		rows, err = s.db.Query(query, input.Slug, input.Limit)
	} else if input.Since != "" && input.Desc {
		query := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3`
		rows, err = s.db.Query(query, input.Slug, input.Limit)
	}

	if err != nil {
		return threads, err
	}
	return scanThreadRows(rows)
}

func (s storage) CheckThreadIfExists(input models.ThreadInput) (thread models.ThreadInput, err error) {
	query := `SELECT ID from threads WHERE ID = $1 OR slug = $2`
	err = s.db.QueryRow(query, input.ThreadID, input.Slug).Scan(&thread.ThreadID)
	return
}

func (s *storage) GetThreadForPost(input models.ThreadInput, thread *models.Thread) (err error) {
	slug := sql.NullString{}
	query := `SELECT author, created, forum, ID, message, slug, title, votes FROM threads WHERE ID = $1`
	err = s.db.QueryRow(query, input.ThreadID).
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
	query := `SELECT forum, ID FROM threads WHERE ID = $1 OR slug = $2`
	err = s.db.QueryRow(query, input.ThreadID, input.Slug).Scan(&forum, &input.ThreadID)
	return
}

func (s *storage) CreateVote(vote models.Vote, update bool) (thread models.Thread, err error) {
	var boolVoice bool
	if vote.Voice == 1 {
		boolVoice = true
	}
	tx, err := s.db.Begin()
	if err != nil {
		fmt.Println("txerr", err)
		return thread, err
	}

	_, err = tx.Exec("SET LOCAL synchronous_commit TO OFF")
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			fmt.Println(txErr)
			return thread, err
		}
		return thread, err
	}

	insertVote := "INSERT INTO votes (user_nick, voice, thread) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT uniq_votes DO UPDATE SET voice = EXCLUDED.voice;"

	_, err = tx.Exec(insertVote, vote.User, boolVoice, vote.Thread.ThreadID)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			fmt.Println(txErr)
			return thread, err
		}
		return thread, err
		/*if pqErr, ok := err.(pgx.PgError); ok {
			switch pqErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return thread, models.Error{Code: "404"}
			default:
				return thread, models.Error{Code: "500"}
			}
		}
		return thread, models.Error{Code: "500"}*/
	}

	slug := sql.NullString{}

	queryUp := `UPDATE threads SET votes = votes + $2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes`
	queryDown := `UPDATE threads SET votes = votes - $2 WHERE ID = $1 RETURNING ID, author, created, forum, message, slug, title, votes`

	delta := 1
	if update {
		delta = 2
	}
	if boolVoice {
		err = tx.QueryRow(queryUp, vote.Thread.ThreadID, delta).
			Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
	} else {
		err = tx.QueryRow(queryDown, vote.Thread.ThreadID, delta).
			Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
	}

	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			fmt.Println(txErr)
			return thread, err
		}

		return thread, err
	}

	if slug.Valid {
		thread.Slug = slug.String
	}

	if commitErr := tx.Commit(); commitErr != nil {
		fmt.Println(commitErr)
		return thread, err
	}

	return
}

/*
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
*/

func (s *storage) CheckDoubleVote(vote models.Vote) (thread models.Thread, nonConflictVice bool, err error) {
	var boolVoice bool
	if vote.Voice == 1 {
		boolVoice = true
	}

	var oldVoice bool
	queryVoice := `SELECT voice FROM votes WHERE user_nick = $1 AND thread = $2`
	err = s.db.QueryRow(queryVoice, vote.User, vote.Thread.ThreadID).Scan(&oldVoice)
	if err != nil {
		if err == pgx.ErrNoRows {
			return thread,false,  nil
		}
		return thread, false, models.Error{Code: "500"}
		//return
	}

	if oldVoice != boolVoice {
		return thread, true, models.Error{Code: "101"}
	}

	//// ID, author, created, forum, message, slug, title, votes
	queryThread := `SELECT ID, author, created, forum, message, slug, title, votes FROM threads WHERE ID = $1`
	thread, err = scanThread(s.db.QueryRow(queryThread, vote.Thread.ThreadID))

	if err != nil {
		return thread, false, models.Error{Code: "500"}
	}

	return thread, false, models.Error{Code: "409"}
}

func (s storage) CreatePosts(thread models.ThreadInput, forum string, created string, posts []models.PostCreate) (post []models.Post, err error) {
	sqlStr := "INSERT INTO posts(id, parent, thread, forum, author, created, message, path) VALUES "
	vals := []interface{}{}
	for _, post := range posts {
		var authorID int
		err = s.db.QueryRow(`SELECT id FROM users WHERE nickname = $1`,
			post.Author,
		).Scan(&authorID)
		if err != nil {
			fmt.Println("cannot find user", post.Author, err)
			return nil, models.Error{Code: "404", Message: "cannot find user"}
		}

		var forumID int
		err = s.db.QueryRow(`SELECT id FROM forums WHERE slug = $1`,
			forum,
		).Scan(&forumID)
		if err != nil {
			fmt.Println("cannot find forumID", forum)
			return nil, models.Error{Code: "404", Message: "cannot find thread"}

			//	return nil, errors.New("404")
		}

		sqlQuery := `
		INSERT INTO forum_users (forumID, userID)
		VALUES ($1,$2)`
		_, err = s.db.Exec(sqlQuery, forumID, authorID)
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
			err = s.db.QueryRow("SELECT thread FROM posts WHERE id = $1",
				post.Parent,
			).Scan(&parentThreadId)
			if err != nil {
				return nil, models.Error{Code: "409", Message: "Parent post was created in another thread"}

				/*	fmt.Println("cannot find thread by post", post.Parent)
					return nil, err*/
			}
			if parentThreadId != int32(thread.ThreadID) {
				return nil, models.Error{Code: "409", Message: "Parent post was created in another thread"}
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
		rows, err := s.db.Query(sqlStr, vals...)
		if err != nil {
			fmt.Println(err)
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
				fmt.Println(err)
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

func (s *storage) CreatePost(input models.Post) (post models.Post, err error) {
	if input.Parent == 0 {
		err = s.db.QueryRow("INSERT INTO posts (author, created, forum, message, parent, thread, path) VALUES ($1,$2,$3,$4,$5,$6, array[(select currval('post_id_seq')::integer)]) RETURNING ID",
			input.Author, input.Created, input.Forum, input.Message, input.Parent, input.ThreadInput.ThreadID).Scan(&post.ID)
	} else {
		err = s.db.QueryRow("INSERT INTO posts (author, created, forum, message, parent, thread, path) VALUES ($1,$2,$3,$4,$5,$6, (SELECT path FROM posts WHERE id = $5) || (select currval('post_id_seq')::integer)) RETURNING ID",
			input.Author, input.Created, input.Forum, input.Message, input.Parent, input.ThreadInput.ThreadID).Scan(&post.ID)
	}

	if pqErr, ok := err.(pgx.PgError); ok {
		fmt.Println(err)
		switch pqErr.Code {
		case pgerrcode.UniqueViolation:
			return post, models.Error{Code: "409", Message: "conflict post"}
		case pgerrcode.NotNullViolation, pgerrcode.ForeignKeyViolation:
			return post, models.Error{Code: "404", Message: "conflict post"}
		default:
			return post, models.Error{Code: "500", Message: "conflict post"}
		}
	}

	post.Author = input.Author
	post.Created = input.Created
	post.Forum = input.Forum
	post.Message = input.Message
	post.Parent = input.Parent
	post.ThreadInput.ThreadID = input.ThreadInput.ThreadID
	post.IsEdited = false

	return
}

func (s *storage) GetPostDetails(input models.PostInput, post *models.Post) (err error) {
	err = s.db.QueryRow("SELECT author, created, forum, message, ID , edited, parent, thread FROM posts WHERE ID = $1", input.ID).
		Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadInput.ThreadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Error{Code: "404"}

		}
		return models.Error{Code: "500"}
	}
	return
}

func (s *storage) UpdatePost(input models.PostUpdate) (post models.Post, err error) {
	var oldMessage string
	err = s.db.QueryRow("SELECT message FROM posts WHERE ID = $1", input.ID).
		Scan(&oldMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return post, models.Error{Code: "404"}
		}
		return post, models.Error{Code: "500"}
	}

	if input.Message != "" && input.Message != oldMessage {
		err = s.db.QueryRow("UPDATE posts SET message = $1, edited = $2 WHERE ID = $3 RETURNING author, created, forum, message, ID , edited, parent, thread", input.Message, true, input.ID).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadInput.ThreadID)
	} else {
		err = s.db.QueryRow("SELECT author, created, forum, message, ID , edited, parent, thread FROM posts WHERE ID = $1", input.ID).
			Scan(&post.Author, &post.Created, &post.Forum, &post.Message, &post.ID, &post.IsEdited, &post.Parent, &post.ThreadInput.ThreadID)
	}

	if err != nil {
		return post, models.Error{Code: "500"}
	}
	return
}

const selectPostsFlatLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY p.created, p.id
	LIMIT $2
`

const selectPostsFlatLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY p.created DESC, p.id DESC
	LIMIT $2
`
const selectPostsFlatLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.id > $2
	ORDER BY p.created, p.id
	LIMIT $3
`
const selectPostsFlatLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.id < $2
	ORDER BY p.created DESC, p.id DESC
	LIMIT $3
`
const selectPostsTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY p.path
	LIMIT $2
`
const selectPostsTreeLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1
	ORDER BY path DESC
	LIMIT $2
`
const selectPostsTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and (p.path > (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path
	LIMIT $3
`
const selectPostsTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and (p.path < (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path DESC
	LIMIT $3
`
const selectPostsParentTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread = $2 AND p2.parent = 0
		ORDER BY p2.path
		LIMIT $3
	)
	ORDER BY path
`
const selectPostsParentTreeLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.parent = 0 and p2.thread = $2
		ORDER BY p2.path DESC
		LIMIT $3
	)
	ORDER BY p.path[1] DESC, p.path[2:]
`

const selectPostsParentTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread = $2 AND p2.parent = 0 and p2.path[1] > (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path
		LIMIT $4
	)
	ORDER BY p.path
`
const selectPostsParentTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent, p.thread, p.forum
	FROM posts p
	WHERE p.thread = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread = $2 AND p2.parent = 0 and p2.path[1] < (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path DESC
		LIMIT $4
	)
	ORDER BY p.path[1] DESC, p.path[2:]
`

func (s *storage) GetPostsByThread(input models.ThreadGetPosts) (posts []models.Post, err error){
	var rows *pgx.Rows
	posts  = make([]models.Post, 0)
	switch input.Sort {
	case "flat":
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsFlatLimitSinceDescByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitSinceByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc == true {
				rows, err = s.db.Query(selectPostsFlatLimitDescByID, input.ThreadInput.ThreadID, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitByID, input.ThreadInput.ThreadID, input.Limit)
			}
		}
	case "tree":
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsTreeLimitSinceDescByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsTreeLimitSinceByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc {
				rows, err = s.db.Query(selectPostsTreeLimitDescByID, input.ThreadInput.ThreadID, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsTreeLimitByID, input.ThreadInput.ThreadID, input.Limit)
			}
		}
	case "parent_tree":
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsParentTreeLimitSinceDescByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsParentTreeLimitSinceByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc {
				rows, err = s.db.Query(selectPostsParentTreeLimitDescByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsParentTreeLimitByID, input.ThreadInput.ThreadID, input.ThreadInput.ThreadID,
					input.Limit)
			}
		}
	default:
		if input.Since > 0 {
			if input.Desc {
				rows, err = s.db.Query(selectPostsFlatLimitSinceDescByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitSinceByID, input.ThreadInput.ThreadID,
					input.Since, input.Limit)
			}
		} else {
			if input.Desc == true {
				rows, err = s.db.Query(selectPostsFlatLimitDescByID, input.ThreadInput.ThreadID, input.Limit)
			} else {
				rows, err = s.db.Query(selectPostsFlatLimitByID, input.ThreadInput.ThreadID, input.Limit)
			}
		}
	}

	if err != nil {
		return posts, models.Error{Code: "500"}
	}
	defer rows.Close()

	if rows == nil {
		return posts, models.Error{Code: "500"}
	}

	for rows.Next() {
		post := models.Post{}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.IsEdited, &post.Message, &post.Parent, &post.ThreadInput.ThreadID, &post.Forum)
		if err != nil {
			return posts, models.Error{Code: "500"}
		}

		posts = append(posts, post)
	}

	return
}

func (s storage) CheckParentPostThread(post int) (thread int, err error) {
	err = s.db.QueryRow("SELECT thread FROM posts WHERE ID = $1", post).Scan(&thread)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, models.Error{Code: "409"}
		}
		return 0, models.Error{Code: "500"}
	}

	return
}