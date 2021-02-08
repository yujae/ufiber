package usecase

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("The email already exists.")
	ErrEmailNotFound      = errors.New("The email does not match any account.")
	ErrPasswordIncorrect  = errors.New("The password is incorrect.")
	ErrInvalidEmail       = errors.New("Invalid Email")

	ErrNickAlreadyExists = errors.New("The nickname already exists.")
	ErrNickTooShort      = errors.New("The nickname is too short.")

	ErrActiveKeyNotFound = errors.New("ActiveKey Not Found")
)

type Error struct {
	//Status string
	Err string
}

func NewError(err error) *Error {
	return &Error{
		//Status: "f",
		Err: err.Error(),
	}
}
