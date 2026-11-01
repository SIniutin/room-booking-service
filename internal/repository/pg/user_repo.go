package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	d "github.com/SIniutin/room-booking-service/internal/domain"
	"github.com/SIniutin/room-booking-service/internal/repository"
)

var _ repository.UserRepo = (*userRepoImpl)(nil)

type userRepoImpl struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) repository.UserRepo {
	return &userRepoImpl{db: db}
}

func (r *userRepoImpl) Create(ctx context.Context, user d.User, creds d.Credentials) (d.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return d.User{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO users (id, email, role) VALUES ($1, $2, $3)`,
		uuid.UUID(user.ID), string(user.Email), string(user.Role),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return d.User{}, d.ErrUserAlreadyExists
		}
		return d.User{}, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO credentials (user_id, password_hash) VALUES ($1, $2)`,
		uuid.UUID(creds.UserID), creds.PasswordHash,
	)
	if err != nil {
		return d.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return d.User{}, err
	}

	return user, nil
}

func (r *userRepoImpl) FindByEmail(ctx context.Context, email d.Email) (d.User, error) {
	var (
		id   uuid.UUID
		role string
	)

	err := r.db.QueryRow(ctx,
		`SELECT id, role FROM users WHERE email = $1`,
		string(email),
	).Scan(&id, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.User{}, d.ErrUserNotFound
		}
		return d.User{}, err
	}

	return d.User{
		ID:    d.UserID(id),
		Email: email,
		Role:  d.UserRole(role),
	}, nil
}

func (r *userRepoImpl) FindByID(ctx context.Context, id d.UserID) (d.User, error) {
	var (
		email string
		role  string
	)

	err := r.db.QueryRow(ctx,
		`SELECT email, role FROM users WHERE id = $1`,
		uuid.UUID(id),
	).Scan(&email, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.User{}, d.ErrUserNotFound
		}
		return d.User{}, err
	}

	parsedEmail, err := d.NewEmail(email)
	if err != nil {
		return d.User{}, err
	}

	return d.User{
		ID:    id,
		Email: *parsedEmail,
		Role:  d.UserRole(role),
	}, nil
}

func (r *userRepoImpl) GetCredentials(ctx context.Context, id d.UserID) (d.Credentials, error) {
	var passwordHash string

	err := r.db.QueryRow(ctx,
		`SELECT password_hash FROM credentials WHERE user_id = $1`,
		uuid.UUID(id),
	).Scan(&passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Credentials{}, d.ErrUserNotFound
		}
		return d.Credentials{}, err
	}

	return d.Credentials{
		UserID:       id,
		PasswordHash: passwordHash,
	}, nil
}
