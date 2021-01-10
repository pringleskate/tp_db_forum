package forumHandler

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/models"
	"github.com/pringleskate/tp_db_forum/internal/storages/forum"
	"github.com/pringleskate/tp_db_forum/internal/storages/profile"
	"math"
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

	forum, err := h.forumStorage.CreateForum(*forumInput)
	if err != nil && err.Error() == "409" {
		oldForum, err := h.forumStorage.GetForumDetails(forumInput.Slug)
		if err != nil {
			return c.JSON(err.(models.ServError).Code, "")
		}

		return c.JSON(models.ConflictData, oldForum)
	}

	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusCreated, forum)
}

func (h *handler) ForumGet(c echo.Context) error {
	slug := c.Param("slug")

	forumRequest, err := h.forumStorage.GetForumDetails(slug)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}
	
	return c.JSON(http.StatusOK, forumRequest)
}

func (h *handler) ForumThreadsGet(c echo.Context) error {
	slug := c.Param("slug")
	params, err := getForumQueryParams(c.QueryParams())
	if err != nil {
		return err
	}

	err = h.forumStorage.CheckIfForumExists(slug)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}
	if params.Limit == 0 {
		params.Limit = 10000
	}
	threads, err :=  h.forumStorage.GetThreadsByForum(params)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusOK, threads)
}

func (h *handler) ForumUsersGet(c echo.Context) error {
	slug := c.Param("slug")
	params, err := getForumQueryParams(c.QueryParams())
	if err != nil {
		return err
	}

	forumID, err := h.forumStorage.GetForumID(slug)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	if params.Limit == 0 {
		params.Limit = 10000
	}
	users, err :=  h.userStorage.GetUsers(params, forumID)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusOK, users)
}


func (h *handler) ThreadCreate(c echo.Context) error {
	threadInput := new(models.Thread)
	if err := c.Bind(threadInput); err != nil {
		return err
	}

	threadInput.Forum = c.Param("slug")

	thread, err := h.forumStorage.CreateThread(*threadInput)
	if err == nil {
		err = h.forumStorage.UpdateThreadsCount(threadInput.Forum)
		if err != nil {
			return c.JSON(err.(models.ServError).Code, "")
		}
		userID, err := h.userStorage.GetUserIDByNickname(threadInput.Author)
		if err != nil {
			return c.JSON(err.(models.ServError).Code, "")
		}

		forumID, err := h.forumStorage.GetForumID(threadInput.Forum)
		if err != nil {
			return c.JSON(err.(models.ServError).Code, "")
		}

		err = h.forumStorage.AddUserToForum(userID, forumID)
		if err != nil && err.Error() != "409" {
			return c.JSON(err.(models.ServError).Code, "")
		}

		return c.JSON(http.StatusCreated, thread)
	}

	if err.Error() == "409"  {
		oldThread, err := h.forumStorage.GetThreadDetails(models.ThreadInput{Slug: input.Slug})
		if err == nil {
			return c.JSON(models.ConflictData, oldThread)
		}
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(err.(models.ServError).Code, "")
}

func (h *handler) ThreadGet(c echo.Context) error {
	slugOrID := isItSlugOrID(c.Param("slug_or_id"))

	thread, err := h.forumStorage.GetThreadDetails(input)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusOK, thread)
}

func (h *handler) ThreadUpdate(c echo.Context) error {
	threadInput := new(models.ThreadUpdate)
	if err := c.Bind(threadInput); err != nil {
		return err
	}

	threadInput.ThreadSlagOrID = isItSlugOrID(c.Param("slug_or_id"))

	thread, err := h.forumStorage.ThreadEdit(*threadInput)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
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

	thread, err := h.forumStorage.CheckThreadIfExists(params.ThreadSlagOrID)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	params.ThreadSlagOrID = thread

	if params.Limit == 0 {
		params.Limit = math.MaxInt32
	}

	posts, err := h.forumStorage.GetPostsByThread(params)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusOK, posts)
}


func (h *handler) ThreadVote(c echo.Context) error {
	voteInput := new(models.Vote)
	if err := c.Bind(voteInput); err != nil {
		return err
	}

	voteInput.ThreadSlagOrID = isItSlugOrID(c.Param("slug_or_id"))

	thread, err := h.forumStorage.CheckThreadIfExists(voteInput.ThreadSlagOrID)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}
	voteInput.ThreadSlagOrID = thread

	var updateFlag bool

	_, checkThread, err := h.forumStorage.CheckDoubleVote(*voteInput)
	if err != nil {
		if err.Error() == "409" {
			return c.JSON(models.ConflictData, checkThread)
		}
		if err.Error() == "500" {
			return c.JSON(err.(models.ServError).Code, "")
		}
		if err.Error() == "101" {
			updateFlag = true
		}
	}

	output, err := h.forumStorage.CreateVote(*voteInput, updateFlag)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusOK, output)
}


func (h *handler) PostCreate(c echo.Context) error {
	postInput := make([]models.Post, 0)

	err := c.Bind(&postInput)
	if err != nil {
		return err
	}

	slagOrID := isItSlugOrID(c.Param("slug_or_id"))
	thread := models.Thread{}

	posts := make([]models.Post, 0)

	forum, err := h.forumStorage.GetForumByThread(&slagOrID)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	if len(postInput) == 0 {
		return c.JSON(http.StatusCreated, posts)
	}

	created := time.Now().Format(time.RFC3339Nano)
	posts, err = h.forumStorage.CreatePosts(thread, forum, created, posts)
	if err != nil {
		if err.Error() == "404" {
			fmt.Println(err)
			return c.JSON(err.(models.ServError).Code, "")
		}
		fmt.Println(err)
		return c.JSON(err.(models.ServError).Code, "")
	}

	err = h.forumStorage.UpdatePostsCount(forum, len(posts))
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusCreated, posts)
}

func (h *handler) PostGet(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	related := relatedParse(c.QueryParam("related"))

	post, err := h.forumStorage.GetPost(id, related)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
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

	post, err := h.forumStorage.UpdatePost(*postInput)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
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