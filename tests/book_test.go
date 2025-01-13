package tests

import (
	"context"
	"testing"
	"time"

	"book-management-system/internal/service"
	pb "book-management-system/pb"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBookService_CreateBook(t *testing.T) {
	db := setupTestDB(t)
	bookService := service.NewBookService(db)

	publishedDate := time.Now()
	tests := []struct {
		name    string
		request *pb.CreateBookRequest
		wantErr bool
	}{
		{
			name: "valid book creation",
			request: &pb.CreateBookRequest{
				Title:         "Test Book",
				Author:        "Test Author",
				PublishedDate: timestamppb.New(publishedDate),
			},
			wantErr: false,
		},
		{
			name: "missing title",
			request: &pb.CreateBookRequest{
				Author:        "Test Author",
				PublishedDate: timestamppb.New(publishedDate),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := bookService.CreateBook(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Book.Id)
			assert.Equal(t, tt.request.Title, resp.Book.Title)
			assert.Equal(t, tt.request.Author, resp.Book.Author)
			assert.Equal(t, "Available", resp.Book.Status)
		})
	}
}

func TestBookService_GetBook(t *testing.T) {
	db := setupTestDB(t)
	bookService := service.NewBookService(db)

	// Create a test book first
	createResp, err := bookService.CreateBook(context.Background(), &pb.CreateBookRequest{
		Title:         "Test Book",
		Author:        "Test Author",
		PublishedDate: timestamppb.New(time.Now()),
	})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		bookID  string
		wantErr bool
	}{
		{
			name:    "existing book",
			bookID:  createResp.Book.Id,
			wantErr: false,
		},
		{
			name:    "non-existent book",
			bookID:  "non-existent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := bookService.GetBook(context.Background(), &pb.GetBookRequest{
				Id: tt.bookID,
			})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, createResp.Book.Id, resp.Book.Id)
			assert.Equal(t, createResp.Book.Title, resp.Book.Title)
			assert.Equal(t, createResp.Book.Author, resp.Book.Author)
		})
	}
}
