package model

type User struct {
	ID        string `form:"ID" db:"id" binding:"required" json:"id"`
	PW        string `form:"PW" db:"pw" binding:"required" json:"pw"`
	NICK      string `db:"nick"`
	SUPERUSER bool   `db:"superuser"`
	ACTIVE    bool   `db:"active"`
	ACTIVEKEY string `db:"activekey"`
}
