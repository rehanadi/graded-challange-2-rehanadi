package tests

import (
	"context"
	"testing"
	"time"

	"book-management-system/internal/auth"
	"book-management-system/internal/models"
	"book-management-system/internal/service"
	pb "book-management-system/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Book{}, &models.BorrowedBook{})
	require.NoError(t, err)

	return db
}

func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(db, jwtManager)

	tests := []struct {
		name    string
		request *pb.RegisterRequest
		wantErr bool
	}{
		{
			name: "valid registration",
			request: &pb.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "duplicate username",
			request: &pb.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := authService.Register(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Id)
			assert.Equal(t, tt.request.Username, resp.Username)
			assert.NotEmpty(t, resp.Token)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour)
	authService := service.NewAuthService(db, jwtManager)

	// First register a user
	registerResp, err := authService.Register(context.Background(), &pb.RegisterRequest{
		Username: "testuser",
		Password: "password123",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		request *pb.LoginRequest
		wantErr bool
	}{
		{
			name: "valid login",
			request: &pb.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "wrong password",
			request: &pb.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			wantErr: true,
		},
		{
			name: "non-existent user",
			request: &pb.LoginRequest{
				Username: "nonexistentuser",
				Password: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := authService.Login(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Token)
			assert.NotNil(t, resp.User)
			assert.Equal(t, registerResp.Id, resp.User.Id)
			assert.Equal(t, registerResp.Username, resp.User.Username)
		})
	}
}
