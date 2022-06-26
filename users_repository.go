package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUniqueConstraint = errors.New("unique constraint")
)

type UsersRepositorier interface {
	CreateUser(user UserWithPassword) (UserBase, error)
	UpdateUser(user UserBase) (UserBase, error)
	DeleteUser(id string) error
	GetUser(id string) (UserBase, error)
	ListUsers() ([]UserBase, error)
}

type UsersRepository struct {
	DB *sqlx.DB
}

func NewUsersRepository(db *sqlx.DB) *UsersRepository {
	return &UsersRepository{DB: db}
}

func (u *UsersRepository) CreateUser(user UserWithPassword) (UserBase, error) {
	// TODO: Handle validation
	_, err := u.DB.NamedExec(`
    insert into users (username,email,password_hash)
    values (:username,:email,:password)`, user)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok {
			if pgerr.Code == "23505" { // unique constraint
				log.Println(err.Error())
				return UserBase{}, fmt.Errorf("invalid database constraint: %s %w",
					pgerr.Detail, ErrUniqueConstraint)
			}
		}

		return UserBase{}, err
	}

	return UserBase{Email: user.Email, Username: user.Username, Id: user.Id}, nil
}

func (u *UsersRepository) UpdateUser(user UserBase) (UserBase, error) {
	res, err := u.DB.NamedExec(
		`update users
    set username=:username
        , email=:email
        , last_modification_date=now()
    where id=:id`, user)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok {
			if pgerr.Code == "23505" { // unique constraint
				log.Println(err.Error())
				return UserBase{}, fmt.Errorf("invalid database constraint: %s %w",
					pgerr.Detail, ErrUniqueConstraint)
			}
		}

		return UserBase{}, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return UserBase{}, err
	}

	if rows == 0 {
		return UserBase{}, fmt.Errorf("error updating user: %w", ErrUserNotFound)
	}

	return user, nil
}

func (u *UsersRepository) DeleteUser(id string) error {
	res, err := u.DB.Exec("delete from users where id=$1", id)
	if err != nil {
		return fmt.Errorf("couldn't delete user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("error deleting user: %w", ErrUserNotFound)
	}

	return nil
}

func (u *UsersRepository) GetUser(id string) (UserBase, error) {
	user := UserBase{}
	err := u.DB.Get(&user, `
    select username
    ,email
    ,id
    from users where id=$1`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserBase{}, fmt.Errorf("could not fetch user: %w", ErrUserNotFound)
		}
		return UserBase{}, fmt.Errorf("could not fetch user: %w", err)
	}

	return user, nil
}

func (u *UsersRepository) ListUsers() ([]UserBase, error) {
	users := []UserBase{}
	err := u.DB.Select(&users, "select username, email, id from users")
	if err != nil {
		return []UserBase{}, fmt.Errorf("couldn't list users: %w", err)
	}
	return users, nil
}
