package mysql

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type UsersRepo struct {
	db *sqlx.DB
}

func NewUsersRepo(db *sqlx.DB) *UsersRepo {
	return &UsersRepo{
		db: db,
	}
}

func (r *UsersRepo) NewUser(ctx context.Context, user *entity.User) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO users (name, email, password) VALUES (:name, :email, :password)", user)
	if err != nil {
		return failure.NewInternalError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return failure.NewInternalError(err)
	}
	user.Id = int(id)
	return nil
}

func (r *UsersRepo) CheckExist(ctx context.Context, email string) (bool, error) {
	var exist bool

	if err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT * FROM users WHERE email=?)", email).Scan(&exist); err != nil {
		return exist, failure.NewInternalError(err)
	}

	return exist, nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	if err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email=?", email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewNotFoundError(errors.New("user not found"))
		}
		return nil, failure.NewInternalError(err)
	}
	return &user, nil
}
