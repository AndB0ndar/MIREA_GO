package repo

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"app/internal/core"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already in use")

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&core.User{})
}

func (r *UserRepo) Create(ctx context.Context, u *core.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		// Проверка на дубликат email
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"idx_users_email\" (SQLSTATE 23505)" {
			return ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (core.User, error) {
	var u core.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.User{}, ErrUserNotFound
	}
	return u, err
}
