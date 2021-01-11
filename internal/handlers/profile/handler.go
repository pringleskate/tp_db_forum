package profileHandler

import (
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/models"
	"github.com/pringleskate/tp_db_forum/internal/storages/profile"
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

	user, err := h.userStorage.UserCreate(*userInput)

	if err == nil {
		return c.JSON(http.StatusCreated, user)
	}

	users := make([]models.User, 0)
	if err.(models.ServError).Code == 409 {
		userNick, err := h.userStorage.SelectFullUser(userInput.Nickname)
		if err != nil && err.(models.ServError).Code != 404 {
			return c.JSON(models.InternalServerError, models.Error{})
		}
		if err == nil {
			users = append(users, userNick)
		}

		if strings.ToLower(userNick.Email) == strings.ToLower(userInput.Email){
			return c.JSON(http.StatusConflict, users)
		}

		userEmail, err := h.userStorage.SelectUserByEmail(userInput.Email)
		if err != nil && err.(models.ServError).Code != 404 {
			return c.JSON(models.InternalServerError, models.Error{})
		}
		if err == nil {
			users = append(users, userEmail)
		}

		return c.JSON(http.StatusConflict, users)
	}

	return c.JSON(err.(models.ServError).Code, models.Error{})
}

func (h *handler) ProfileGet(c echo.Context) error {

	nickname := c.Param("nickname")

	user, err := h.userStorage.SelectFullUser(nickname)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, models.Error{})
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
		user, err :=  h.userStorage.SelectFullUser(userInput.Nickname)
		if err != nil {
			return c.JSON(err.(models.ServError).Code, models.Error{})
		}
		return c.JSON(http.StatusOK, user)
	}

	user, err := h.userStorage.EditUser(*userInput)
	if err != nil {
		return c.JSON(err.(models.ServError).Code, models.Error{})
	}

	return c.JSON(http.StatusOK, user)
}
