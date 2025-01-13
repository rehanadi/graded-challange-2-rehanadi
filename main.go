package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"book-management-system/internal/auth"
	"book-management-system/internal/models"
	"book-management-system/internal/scheduler"
	"book-management-system/internal/service"
	pb "book-management-system/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=12345 dbname=bookmanagement port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto migrate the schemas
	err = db.AutoMigrate(&models.User{}, &models.Book{}, &models.BorrowedBook{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create JWT manager
	jwtManager := auth.NewJWTManager("secret", 24*time.Hour)

	// Initialize services
	authService := service.NewAuthService(db, jwtManager) // Changed from NewUserService to NewAuthService
	bookService := service.NewBookService(db)
	borrowService := service.NewBorrowService(db)

	// Initialize scheduler
	bookScheduler := scheduler.NewBookScheduler(db)
	if err := bookScheduler.Start(); err != nil {
		log.Printf("Failed to start scheduler: %v", err)
	}
	defer bookScheduler.Stop()

	// Initialize gRPC server with auth interceptor
	server := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryAuthInterceptor(jwtManager)),
	)

	// Register services
	pb.RegisterAuthServiceServer(server, authService)
	pb.RegisterBookServiceServer(server, bookService)
	pb.RegisterBorrowServiceServer(server, borrowService)

	// Register reflection service for grpcurl
	reflection.Register(server)

	// Start listening
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
