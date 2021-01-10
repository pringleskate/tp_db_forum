package handlers

import (
	"github.com/labstack/echo"
	forumHandler "github.com/pringleskate/tp_db_forum/internal/handlers/forum"
	profileHandler "github.com/pringleskate/tp_db_forum/internal/handlers/profile"
	serviceHandler "github.com/pringleskate/tp_db_forum/internal/handlers/service"
)

func Router(e *echo.Echo, forum forumHandler.Handler, profile profileHandler.Handler, service serviceHandler.Handler) {

	e.POST("/api/forum/create", forum.ForumCreate)
	e.POST("/api/forum/:slug/create", forum.ThreadCreate)
	e.GET("/api/forum/:slug/details", forum.ForumGet)
	e.GET("/api/forum/:slug/threads", forum.ForumThreadsGet)
	e.GET("/api/forum/:slug/users", forum.ForumUsersGet)

	e.GET("/api/post/:id/details", forum.PostGet)
	e.POST("/api/post/:id/details", forum.PostUpdate)

	e.POST("/api/service/clear", service.ServiceClear)
	e.GET("/api/service/status", service.ServiceStatus)

	e.POST("/api/thread/:slug_or_id/create", forum.PostCreate)
	e.GET("/api/thread/:slug_or_id/details", forum.ThreadGet)
	e.POST("/api/thread/:slug_or_id/details", forum.ThreadUpdate)
	e.GET("/api/thread/:slug_or_id/posts", forum.ThreadPostsGet)
	e.POST("/api/thread/:slug_or_id/vote", forum.ThreadVote)

	e.POST("/api/user/:nickname/create", profile.ProfileCreate)
	e.GET("/api/user/:nickname/profile", profile.ProfileGet)
	e.POST("/api/user/:nickname/profile", profile.ProfileUpdate)
}