package handler

import (
	"github.com/Duvewo/testexercise/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	*pgxpool.Pool
	Users storage.UsersRepository
}
