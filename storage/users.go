package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	UsersRepository interface {
		Create(context.Context, User) error
		SetHealth(context.Context, User) error
		ByUsername(context.Context, User) (User, error)
		ExistsByUsernameAndPassword(context.Context, User) (bool, error)
	}

	Users struct {
		*pgxpool.Pool
	}

	User struct {
		ID           int    `db:"id"`
		Username     string `db:"username"`
		Password     string `db:"password"`
		HealthPoints int    `db:"health_points"`
	}
)

func (db *Users) Create(ctx context.Context, u User) error {
	const q = "INSERT INTO users (username, password, health_points) VALUES ($1, $2, 100)"
	_, err := db.Exec(ctx, q, u.Username, u.Password)
	return err
}

// TODO: change to Update
func (db *Users) SetHealth(ctx context.Context, u User) error {
	const q = "UPDATE users SET health_points=$1 WHERE username = $2"
	_, err := db.Exec(ctx, q, u.HealthPoints, u.Username)
	return err
}

func (db *Users) ByUsername(ctx context.Context, u User) (usr User, err error) {
	const q = "SELECT * FROM users WHERE username = $1"
	return usr, db.QueryRow(ctx, q, u.Username).Scan(&usr.ID, &usr.Username, &usr.Password, &usr.HealthPoints)
}

func (db *Users) ExistsByUsernameAndPassword(ctx context.Context, u User) (bool, error) {
	return false, nil
}
