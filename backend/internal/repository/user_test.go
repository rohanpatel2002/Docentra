package repository

import (
	"ai-document-assistant/internal/models"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: false,
	})
	require.NoError(t, err)
	return db, mock, mockDB
}
func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, mockDB := setupMockDB(t)
	defer mockDB.Close()
	repo := NewUserRepository(db)
	email := "test@example.com"
	hash := "hashedpassword"
	user := &models.User{
		Email:        email,
		PasswordHash: hash,
	}
	mock.ExpectBegin()
	// Insert query
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("email","password_hash","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5) RETURNING "id"`)).
		WithArgs(user.Email, user.PasswordHash, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	err := repo.CreateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), user.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}
func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, mock, mockDB := setupMockDB(t)
	defer mockDB.Close()
	repo := NewUserRepository(db)
	email := "test@example.com"
	now := time.Now().Truncate(time.Microsecond) // PostgreSQL truncates precision
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, email, "hashedpassword", now, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = 1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(email, 1).
		WillReturnRows(rows)
	resUser, err := repo.GetUserByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, resUser)
	assert.Equal(t, uint(1), resUser.ID)
	assert.Equal(t, email, resUser.Email)
	assert.Equal(t, now, resUser.CreatedAt, "CreatedAt timestamp should accurately map from database")
	assert.Equal(t, now, resUser.UpdatedAt, "UpdatedAt timestamp should accurately map from database")
	require.NoError(t, mock.ExpectationsWereMet())
}
func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, mockDB := setupMockDB(t)
	defer mockDB.Close()
	repo := NewUserRepository(db)
	id := uint(1)
	now := time.Now().Truncate(time.Microsecond)
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "deleted_at"}).
		AddRow(id, "test@example.com", "hashedpassword", now, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = 1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(id, 1). // Limit is usually bound to an argument in GORM Postgres
		WillReturnRows(rows)
	resUser, err := repo.GetUserByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, resUser)
	assert.Equal(t, id, resUser.ID)
	assert.Equal(t, "test@example.com", resUser.Email)
	assert.Equal(t, now, resUser.CreatedAt)
	assert.Equal(t, now, resUser.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}
