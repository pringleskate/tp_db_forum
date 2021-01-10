package models

type Error struct {
	Message string `json:"message"`
}

type ServError struct {
	Code    int
	Message string
}

func (e ServError) Error() string {
	return e.Message
}

var (
	InternalServerError = 500
	NotFound = 404
	ConflictData = 409
)