package repository

import (
	"database/sql"
	"ufiber/model"
)

type UserHistoryI interface {
	Create(*model.UserHistory) error
}

type UserHistoryR struct {
	*sql.DB
}

func NewUserHistoryR(db *sql.DB) *UserHistoryR {
	return &UserHistoryR{db}
}

func (db *UserHistoryR) Create(u *model.UserHistory) error {
	_, err := db.Exec("insert into public.user_history (user_id, msg, accessed) "+
		"values ($1, $2, current_timestamp)", u.USER_ID, u.MSG)
	if err != nil {
		return err
	}
	return nil
}
