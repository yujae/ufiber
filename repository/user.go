package repository

import (
	"database/sql"
	"ufiber/model"
)

type UserI interface {
	Create(*model.User) error
	Retrieve(string) (model.User, error)
	UpdatePw(*model.User) error
	UpdateNick(*model.User) error
	UpdateActive(string, string) error
	UpdateLogin(string) error
	Delete(string) error
}

type userR struct {
	*sql.DB
}

func NewUserR(db *sql.DB) *userR {
	return &userR{db}
}

func (db *userR) Create(u *model.User) error {
	_, err := db.Exec("insert into public.user (id, pw, nick, joined, activekey) "+
		"values ($1, $2, $3, $4, current_timestamp)", u.ID, u.PW, u.NICK, u.ACTIVEKEY)
	if err != nil {
		return err
	}
	return nil
}
func (db *userR) Retrieve(id string) (model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, pw, nick, active FROM public.user WHERE ID = $1", id).Scan(&user.ID, &user.PW, &user.NICK, &user.ACTIVE)
	if err != nil {
		return user, err
	}
	return user, nil
}
func (db *userR) UpdatePw(u *model.User) error {
	_, err := db.Exec("update public.user set pw=$1 where id=$2", u.PW, u.ID)
	if err != nil {
		return err
	}
	return nil
}
func (db *userR) UpdateNick(u *model.User) error {
	_, err := db.Exec("update public.user set nick=$1 where id=$2", u.NICK, u.ID)
	if err != nil {
		return err
	}
	return nil
}
func (db *userR) UpdateActive(id string, activekey string) error {
	_, err := db.Exec("update public.user set active=true, activated=current_timestamp where id=$1 and activekey=$2", id, activekey)
	if err != nil {
		return err
	}
	return nil
}
func (db *userR) UpdateLogin(id string) error {
	_, err := db.Exec("update public.user set login=current_timestamp where id=$1", id)
	if err != nil {
		return err
	}

	return nil
}
func (db *userR) Delete(id string) error {
	_, err := db.Exec("delete from public.user where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}
