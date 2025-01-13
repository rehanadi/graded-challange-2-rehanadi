package service

import (
	"context"
	"time"

	"book-management-system/internal/auth"
	"book-management-system/internal/models"
	pb "book-management-system/pb"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	db         *gorm.DB
	jwtManager *auth.JWTManager
}

func NewAuthService(db *gorm.DB, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		db:         db,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.RegisterResponse{
		Id:       user.ID,
		Username: user.Username,
		Token:    token,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{
		Token: token,
		User: &pb.UserInfo{
			Id:       user.ID,
			Username: user.Username,
		},
	}, nil
}

type BookService struct {
	pb.UnimplementedBookServiceServer
	db *gorm.DB
}

func NewBookService(db *gorm.DB) *BookService {
	return &BookService{db: db}
}

func (s *BookService) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.BookResponse, error) {

	book := &models.Book{
		Title:         req.Title,
		Author:        req.Author,
		PublishedDate: req.PublishedDate.AsTime(),
		Status:        "Available",
		UserID:        nil,
	}

	if err := s.db.Create(book).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create book")
	}

	return &pb.BookResponse{
		Book: &pb.Book{
			Id:            book.ID,
			Title:         book.Title,
			Author:        book.Author,
			PublishedDate: timestamppb.New(book.PublishedDate),
			Status:        book.Status,
			UserId:        "",
		},
	}, nil
}

func (s *BookService) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.BookResponse, error) {
	var book models.Book
	if err := s.db.First(&book, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "book not found")
	}

	var userID string
	if book.UserID != nil {
		userID = *book.UserID
	}

	return &pb.BookResponse{
		Book: &pb.Book{
			Id:            book.ID,
			Title:         book.Title,
			Author:        book.Author,
			PublishedDate: timestamppb.New(book.PublishedDate),
			Status:        book.Status,
			UserId:        userID,
		},
	}, nil
}

func (s *BookService) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	var books []models.Book
	var total int64

	query := s.db.Model(&models.Book{})

	// Apply status filter if provided
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count books: %v", err)
	}

	// Apply pagination
	offset := (req.Page - 1) * req.PerPage
	if err := query.Offset(int(offset)).Limit(int(req.PerPage)).Find(&books).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list books: %v", err)
	}

	// Convert to proto message
	pbBooks := make([]*pb.Book, len(books))
	for i, book := range books {
		var userID string
		if book.UserID != nil {
			userID = *book.UserID
		}

		pbBooks[i] = &pb.Book{
			Id:            book.ID,
			Title:         book.Title,
			Author:        book.Author,
			PublishedDate: timestamppb.New(book.PublishedDate),
			Status:        book.Status,
			UserId:        userID,
		}
	}

	return &pb.ListBooksResponse{
		Books:   pbBooks,
		Total:   int32(total),
		Page:    req.Page,
		PerPage: req.PerPage,
	}, nil
}

func (s *BookService) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.BookResponse, error) {
	var book models.Book
	if err := s.db.First(&book, "id = ?", req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get book: %v", err)
	}

	book.Title = req.Title
	book.Author = req.Author
	book.PublishedDate = req.PublishedDate.AsTime()

	if err := s.db.Save(&book).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update book: %v", err)
	}

	var userID string
	if book.UserID != nil {
		userID = *book.UserID
	}

	return &pb.BookResponse{
		Book: &pb.Book{
			Id:            book.ID,
			Title:         book.Title,
			Author:        book.Author,
			PublishedDate: timestamppb.New(book.PublishedDate),
			Status:        book.Status,
			UserId:        userID,
		},
	}, nil
}

func (s *BookService) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*emptypb.Empty, error) {
	result := s.db.Delete(&models.Book{}, "id = ?", req.Id)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete book: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "book not found")
	}

	return &emptypb.Empty{}, nil
}

type BorrowService struct {
	pb.UnimplementedBorrowServiceServer
	db *gorm.DB
}

func NewBorrowService(db *gorm.DB) *BorrowService {
	return &BorrowService{db: db}
}

func (s *BorrowService) BorrowBook(ctx context.Context, req *pb.BorrowBookRequest) (*pb.BorrowResponse, error) {
	userID, err := auth.ExtractUserID(ctx)
	if err != nil {
		return nil, err
	}

	var book models.Book
	if err := s.db.First(&book, "id = ? AND status = ?", req.BookId, "Available").Error; err != nil {
		return nil, status.Error(codes.NotFound, "book not found or not available")
	}

	borrowedBook := &models.BorrowedBook{
		BookID:       book.ID,
		UserID:       userID,
		BorrowedDate: time.Now(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(borrowedBook).Error; err != nil {
			return err
		}

		if err := tx.Model(&book).Updates(map[string]interface{}{
			"status":  "Borrowed",
			"user_id": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to borrow book")
	}

	return &pb.BorrowResponse{
		BorrowedBook: &pb.BorrowedBook{
			Id:           borrowedBook.ID,
			BookId:       borrowedBook.BookID,
			UserId:       borrowedBook.UserID,
			BorrowedDate: timestamppb.New(borrowedBook.BorrowedDate),
			Book: &pb.Book{
				Id:            book.ID,
				Title:         book.Title,
				Author:        book.Author,
				PublishedDate: timestamppb.New(book.PublishedDate),
				Status:        "Borrowed",
				UserId:        userID,
			},
		},
	}, nil
}

func (s *BorrowService) ReturnBook(ctx context.Context, req *pb.ReturnBookRequest) (*pb.BorrowResponse, error) {
	var borrowedBook models.BorrowedBook
	if err := s.db.Preload("Book").First(&borrowedBook, "id = ?", req.BorrowId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "borrow record not found")
	}

	returnTime := time.Now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		borrowedBook.ReturnDate = &returnTime
		if err := tx.Save(&borrowedBook).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Book{}).Where("id = ?", borrowedBook.BookID).
			Updates(map[string]interface{}{
				"status":  "Available",
				"user_id": nil,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to return book")
	}

	return &pb.BorrowResponse{
		BorrowedBook: &pb.BorrowedBook{
			Id:           borrowedBook.ID,
			BookId:       borrowedBook.BookID,
			UserId:       borrowedBook.UserID,
			BorrowedDate: timestamppb.New(borrowedBook.BorrowedDate),
			ReturnDate:   timestamppb.New(*borrowedBook.ReturnDate),
			Book: &pb.Book{
				Id:            borrowedBook.Book.ID,
				Title:         borrowedBook.Book.Title,
				Author:        borrowedBook.Book.Author,
				PublishedDate: timestamppb.New(borrowedBook.Book.PublishedDate),
				Status:        "Available",
			},
		},
	}, nil
}

func (s *BorrowService) ListBorrowedBooks(ctx context.Context, req *pb.ListBorrowedBooksRequest) (*pb.ListBorrowedBooksResponse, error) {
	userID, err := auth.ExtractUserID(ctx)
	if err != nil {
		return nil, err
	}

	var borrowedBooks []models.BorrowedBook
	var total int64

	query := s.db.Model(&models.BorrowedBook{}).Where("user_id = ?", userID)

	if !req.IncludeReturned {
		query = query.Where("return_date IS NULL")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count borrowed books: %v", err)
	}

	offset := (req.Page - 1) * req.PerPage
	if err := query.Preload("Book").Offset(int(offset)).Limit(int(req.PerPage)).Find(&borrowedBooks).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list borrowed books: %v", err)
	}

	pbBorrowedBooks := make([]*pb.BorrowedBook, len(borrowedBooks))
	for i, bb := range borrowedBooks {
		var userID string
		if bb.Book.UserID != nil {
			userID = *bb.Book.UserID
		}

		pbBorrowedBooks[i] = &pb.BorrowedBook{
			Id:           bb.ID,
			BookId:       bb.BookID,
			UserId:       bb.UserID,
			BorrowedDate: timestamppb.New(bb.BorrowedDate),
			Book: &pb.Book{
				Id:            bb.Book.ID,
				Title:         bb.Book.Title,
				Author:        bb.Book.Author,
				PublishedDate: timestamppb.New(bb.Book.PublishedDate),
				Status:        bb.Book.Status,
				UserId:        userID,
			},
		}
		if bb.ReturnDate != nil {
			pbBorrowedBooks[i].ReturnDate = timestamppb.New(*bb.ReturnDate)
		}
	}

	return &pb.ListBorrowedBooksResponse{
		BorrowedBooks: pbBorrowedBooks,
		Total:         int32(total),
		Page:          req.Page,
		PerPage:       req.PerPage,
	}, nil
}
