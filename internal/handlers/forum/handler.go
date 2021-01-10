package forumHandler

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/models"
	"github.com/pringleskate/tp_db_forum/internal/storages/forum"
	"github.com/pringleskate/tp_db_forum/internal/storages/profile"
	"golang.org/x/mod/sumdb/storage"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Handler interface {
	ForumCreate(c echo.Context) error
	ForumGet(c echo.Context) error
	ForumThreadsGet(c echo.Context) error
	ForumUsersGet(c echo.Context) error

	ThreadCreate(c echo.Context) error
	ThreadGet(c echo.Context) error
	ThreadUpdate(c echo.Context) error
	ThreadPostsGet(c echo.Context) error

	ThreadVote(c echo.Context) error

	PostCreate(c echo.Context) error
	PostGet(c echo.Context) error
	PostUpdate(c echo.Context) error
}

type handler struct {
	forumStorage forum.Storage
	userStorage profile.Storage
}

func NewHandler(forumStorage forum.Storage, userStorage profile.Storage) *handler {
	return &handler{
		forumStorage: forumStorage,
		userStorage: userStorage,
	}
}

func (h *handler) ForumCreate(c echo.Context) error {
	forumInput := new(models.Forum)
	if err := c.Bind(forumInput); err != nil {
		return err
	}

	existing, err := h.forumStorage.GetFullForum(forumInput.Slug)
	if err == nil {
		return c.JSON(http.StatusConflict, existing)
	}
	if err != pgx.ErrNoRows {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	nickname, err := h.userStorage.GetUserNickname((*forumInput).User)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	forumInput.User = nickname

	err = h.forumStorage.InsertForum(*forumInput)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusCreated, forumInput)
}

func (h *handler) ForumGet(c echo.Context) error {
	slug := c.Param("slug")

	forumRequest, err := h.forumStorage.GetFullForum(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})	
	}
	
	return c.JSON(http.StatusOK, forumRequest)
}

func (h *handler) ForumThreadsGet(c echo.Context) error {
	slug := c.Param("slug")
	params, err := getForumQueryParams(c.QueryParams())
	if err != nil {
		return err
	}

	threads := make([]models.Thread, 0)

	forumForSlug, err := h.forumStorage.GetFullForum(slug)
	if err != nil {
		log.Print(err)
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	params.Slug = forumForSlug.Slug
	if params.Limit == 0 {
		params.Limit = 10000
	}

	threads, err = h.threadStorage.GetAllThreadsByForum(params)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusOK, threads)
}

func (h *handler) ForumUsersGet(c echo.Context) error {
	slug := c.Param("slug")
	params, err := getForumQueryParams(c.QueryParams())
	if err != nil {
		return err
	}

	users := make([]models.User, 0)

	forumForSlug, err := h.forumStorage.GetFullForum(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	params.Slug = forumForSlug.Slug
	if params.Limit == 0 {
		params.Limit = 10000
	}

	users, err = h.userStorage.GetAllUsersByForum(params)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusOK, users)
}


func (h *handler) ThreadCreate(c echo.Context) error {
	threadInput := new(models.Thread)
	if err := c.Bind(threadInput); err != nil {
		return err
	}

	threadInput.Forum = c.Param("slug")

	forumForSlug, err := h.forumStorage.GetFullForum(threadInput.Forum)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: "No such forum"})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	threadInput.Forum = forumForSlug.Slug

	if threadInput.Slag != "" {
		existing, err := h.threadStorage.GetFullThreadBySlug(threadInput.Slag)
		if err == nil {
			return c.JSON(http.StatusConflict, existing)
		}
		if err != pgx.ErrNoRows {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}
	}

	nickname, err := h.userStorage.GetUserNickname(threadInput.Author)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: "User not found"})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	threadInput.Author = nickname
	ID, err := h.threadStorage.InsertThread(*threadInput)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	err = h.forumStorage.InsertForumUser(threadInput.Forum, threadInput.Author)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	err = h.forumStorage.UpdateThreadsCount(threadInput.Forum)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	threadInput.ID = ID

	return c.JSON(http.StatusCreated, threadInput)
}

func (h *handler) ThreadGet(c echo.Context) error {
	slugOrID := isItSlugOrID(c.Param("slug_or_id"))

	var thread models.Thread
	var err error

	if slugOrID.ThreadSlug != "" {
		thread, err = h.threadStorage.GetFullThreadBySlug(slugOrID.ThreadSlug)
	} else {
		thread, err = h.threadStorage.GetFullThreadByID(slugOrID.ThreadID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusOK, thread)
}

func (h *handler) ThreadUpdate(c echo.Context) error {
	threadInput := new(models.ThreadUpdate)
	if err := c.Bind(threadInput); err != nil {
		return err
	}

	threadInput.ThreadSlagOrID = isItSlugOrID(c.Param("slug_or_id"))

	var thread models.Thread
	var err error

	if threadInput.ThreadSlug != "" {
		thread, err = h.threadStorage.GetFullThreadBySlug(threadInput.ThreadSlug)
	} else {
		thread, err = h.threadStorage.GetFullThreadByID(threadInput.ThreadID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	if threadInput.Title != "" {
		thread.Title = threadInput.Title
	}
	if threadInput.Message != "" {
		thread.Message = threadInput.Message
	}
	if threadInput.Title == "" && threadInput.Message == "" {
		return c.JSON(http.StatusOK, thread)
	}

	err = h.threadStorage.UpdateThread(thread)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusOK, thread)
}

func (h *handler) ThreadPostsGet(c echo.Context) error {
	params, err := getThreadQueryParams(c.QueryParams())
	if err != nil {
		fmt.Println(err)
		return err
	}

	params.ThreadSlagOrID = isItSlugOrID(c.Param("slug_or_id"))

	posts := make([]models.Post, 0)

	thread := models.Thread{}
	if params.ThreadSlug != "" {
		thread, err = h.threadStorage.GetFullThreadBySlug(params.ThreadSlug)
	} else {
		thread, err = h.threadStorage.GetFullThreadByID(params.ThreadID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	if params.Limit == 0 {
		params.Limit = 10000
	}
	params.ThreadID = thread.ID

	posts, err = h.postStorage.GetAllPostsByThread(params)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusOK, posts)
}


func (h *handler) ThreadVote(c echo.Context) error {
	voteInput := new(models.Vote)
	if err := c.Bind(voteInput); err != nil {
		return err
	}

	voteInput.ThreadSlagOrID = isItSlugOrID(c.Param("slug_or_id"))

	var thread models.Thread
	var err error

	if voteInput.ThreadSlug != "" {
		thread, err = h.threadStorage.GetFullThreadBySlug(voteInput.ThreadSlug)
	} else {
		thread, err = h.threadStorage.GetFullThreadByID(voteInput.ThreadID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}
	voteInput.ThreadID = thread.ID

	user, err := h.userStorage.GetUserNickname(voteInput.Nickname)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}
	voteInput.Nickname = user

	vote, err := h.threadStorage.SelectVote(*voteInput)
	if err != nil && err != pgx.ErrNoRows {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	var newVoice int
	if err == nil {
		if voteInput.Voice == vote.Voice {
			return c.JSON(http.StatusOK, thread)
		}
		err = h.threadStorage.UpdateVote(*voteInput)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}

		if voteInput.Voice == -1 {
			newVoice = -2
		} else {
			newVoice = 2
		}

		err = h.threadStorage.UpdateVotesCount(voteInput.ThreadID, newVoice)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}

	} else {

		err = h.threadStorage.InsertVote(*voteInput)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}

		newVoice = int(voteInput.Voice)
		err = h.threadStorage.UpdateVotesCount(voteInput.ThreadID, newVoice)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}
	}

	thread.Votes += newVoice

	return c.JSON(http.StatusOK, thread)
}


func (h *handler) PostCreate(c echo.Context) error {
	postInput := make([]models.Post, 0)

	err := c.Bind(&postInput)
	if err != nil {
		return err
	}

	slagOrID := isItSlugOrID(c.Param("slug_or_id"))
	thread := models.Thread{}

	if slagOrID.ThreadSlug != "" {
		thread, err = h.threadStorage.GetFullThreadBySlug(slagOrID.ThreadSlug)
	} else {
		thread, err = h.threadStorage.GetFullThreadByID(slagOrID.ThreadID)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ServError{Message: "Can't find thread"})
		}
		log.Print(err)
		return c.JSON(http.StatusInternalServerError, models.ServError{ Code: models.InternalServerError })
	}

	if len(postInput) == 0 {
		return c.JSON(http.StatusCreated, postInput)
	}
	createdTime := time.Now().Format(time.RFC3339Nano)
	forumOfThread := thread.Forum

	posts, err :=  h.postStorage.CreatePosts(thread, forumOfThread, createdTime, postInput)
	if err != nil {
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, models.ServError{Message: "Can't find post author by nickname:"})
			}
		return c.JSON(http.StatusConflict, models.ServError{Message: "Parent post was created in another thread"})
	}

	err = h.forumStorage.UpdatePostsCount(forumOfThread, len(posts))
	if err != nil {
		log.Print(err)
		return c.JSON(http.StatusInternalServerError, models.ServError{ Code: models.InternalServerError })
	}

	return c.JSON(http.StatusCreated, posts)
}

func (h *handler) PostGet(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	related := relatedParse(c.QueryParam("related"))

	var post models.PostFull

	onePost, err := h.postStorage.GetFullPost(int(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	post.Post = &onePost
	args := fmt.Sprint(related)
	if strings.Contains(args, "user") {
		user, err := h.userStorage.GetFullUserByNickname(post.Post.Author)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}
		post.Author = &user
	}
	if strings.Contains(args, "forum") {
		forumOfPost, err := h.forumStorage.GetFullForum(post.Post.Forum)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}
		post.Forum = &forumOfPost
	}
	if strings.Contains(args, "thread") {
		thread, err := h.threadStorage.GetFullThreadByID(post.Post.ThreadID)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}
		post.Thread = &thread
	}

	return c.JSON(http.StatusOK, post)
}

func (h *handler) PostUpdate(c echo.Context) (err error) {
	postInput := new(models.PostUpdate)
	if err := c.Bind(postInput); err != nil {
		return err
	}

	postInput.ID , err = strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	post, err := h.postStorage.GetFullPost(postInput.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	if postInput.Message != "" && postInput.Message != post.Message {
		err = h.postStorage.UpdatePost(*postInput)
		if err != nil {
			log.Print(err)
			return c.JSON(models.InternalServerError, models.Error{Message: ""})
		}

		post.Message = postInput.Message
		post.IsEdited = true
	}

	return c.JSON(http.StatusOK, post)
}

func getForumQueryParams(params url.Values) (values models.ForumQueryParams, err error) {
	limit := params.Get("limit")
	if limit != "" {
		values.Limit, err = strconv.Atoi(limit)
	}

	values.Since = params.Get("since")

	desc := params.Get("desc")
	if desc != "" {
		values.Desc, err = strconv.ParseBool(desc)
		if err != nil {
			return models.ForumQueryParams{}, err
		}
	}

	return values, nil
}

func getThreadQueryParams(params url.Values) (values models.ThreadQueryParams, err error) {
	limit := params.Get("limit")
	if limit != "" {
		values.Limit, err = strconv.Atoi(limit)
	}

	since := params.Get("since")
	if since != "" {
		values.Since, err = strconv.Atoi(since)
	}

	values.Sort = params.Get("sort")
	desc := params.Get("desc")
	if desc != "" {
		values.Desc, err = strconv.ParseBool(desc)
		if err != nil {
			return models.ThreadQueryParams{}, err
		}
	}

	if err != nil {
		return models.ThreadQueryParams{}, err
	}
	return values, nil
}

func relatedParse(related string) []string {
	related = strings.ReplaceAll(related, "[", "")
	related = strings.ReplaceAll(related, "]", "")
	return strings.Split(related, ",")
}

func isItSlugOrID(slagOrID string) (output models.ThreadSlagOrID) {
	id, err := strconv.Atoi(slagOrID)
	if err != nil {
		output.ThreadSlug = slagOrID
		return output
	}
	output.ThreadID = id
	return output
}