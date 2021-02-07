package model

type UserHistory struct {
	USER_ID string `db:"user_id"`
	MSG     string `db:"msg"`
}
