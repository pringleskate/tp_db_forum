package profileHandler

import (
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/models"
	"github.com/pringleskate/tp_db_forum/internal/storages/profile"
	"log"
	"net/http"
	"strings"
)

type Handler interface {
	ProfileCreate(c echo.Context) error
	ProfileGet(c echo.Context) error
	ProfileUpdate(c echo.Context) error
}

type handler struct {
	userStorage profile.Storage
}

func NewHandler(userStorage profile.Storage) *handler {
	return &handler{
		userStorage: userStorage,
	}
}

func (h *handler) ProfileCreate(c echo.Context) error {
	userInput := new(models.User)
	if err := c.Bind(userInput); err != nil {
		return err
	}

	userInput.Nickname = c.Param("nickname")

	user, err := h.userStorage.CreateUser(*userInput)

	if err == nil {
		return c.JSON(http.StatusCreated, user)
	}

	users := make([]models.User, 0)
	if err.Error() == "409" {
		userNick, err := h.userStorage.GetProfile(userInput.Nickname)
		if err != nil && err.Error() != "404"{
			return c.JSON(models.InternalServerError, "")
		}
		if err == nil {
			users = append(users, userNick)
		}

		if strings.ToLower(userNick.Email) == strings.ToLower(userInput.Email){
			return c.JSON(http.StatusConflict, users)
		}

		userEmail, err := h.userStorage.GetEmailConflictUser(userInput.Email)
		if err != nil && err.Error() != "404"{
			return c.JSON(models.InternalServerError, "")
		}
		if err == nil {
			users = append(users, userEmail)
		}

		return c.JSON(http.StatusConflict, users)
	}

	return c.JSON(models.InternalServerError, "")
}

func (h *handler) ProfileGet(c echo.Context) error {

	nickname := c.Param("nickname")

	user, err := h.userStorage.GetProfile(nickname)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *handler) ProfileUpdate(c echo.Context) error {
	userInput := new(models.User)
	if err := c.Bind(userInput); err != nil {
		return err
	}

	userInput.Nickname = c.Param("nickname")

	if userInput.Email == "" && userInput.FullName == "" && userInput.About == "" {
		user, err :=  h.userStorage.GetProfile(userInput.Nickname)
		if err != nil {
			return c.JSON(err.(models.ServError).Code, "")
		}
		return c.JSON(http.StatusOK, user)
	}

	user, err := h.userStorage.UpdateProfile(*userInput)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, "")
	}

	return c.JSON(http.StatusOK, user)
}
