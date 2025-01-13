package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"book-management-system/client/handler"
	pb "book-management-system/pb"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Connect to gRPC server
	serverUrl := os.Getenv("SERVER_URL")
	serverPort := os.Getenv("SERVER_PORT")
	conn, err := grpc.Dial(serverUrl+":"+serverPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Initialize gRPC clients
	authClient := pb.NewAuthServiceClient(conn)
	bookClient := pb.NewBookServiceClient(conn)
	borrowClient := pb.NewBorrowServiceClient(conn)

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authClient)
	bookHandler := handler.NewBookHandler(bookClient)
	borrowHandler := handler.NewBorrowHandler(borrowClient)

	// Serve Swagger UI
	e.GET("/swagger/*", echoSwagger.EchoWrapHandler(echoSwagger.URL("/swagger/doc.json")))

	// Serve Swagger JSON
	e.GET("/swagger/doc.json", func(c echo.Context) error {
		swagger, err := os.ReadFile("docs/swagger/openapi.json")
		if err != nil {
			return err
		}
		return c.JSONBlob(http.StatusOK, swagger)
	})

	// API Routes
	api := e.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Book routes
	books := api.Group("/books")
	books.Use(handler.AuthMiddleware)
	{
		books.POST("", bookHandler.CreateBook)
		books.GET("", bookHandler.ListBooks)
		books.GET("/:id", bookHandler.GetBook)
		books.PUT("/:id", bookHandler.UpdateBook)
		books.DELETE("/:id", bookHandler.DeleteBook)
	}

	// Borrow routes
	borrow := api.Group("/borrow")
	borrow.Use(handler.AuthMiddleware)
	{
		borrow.POST("", borrowHandler.BorrowBook)
		borrow.PUT("/:id/return", borrowHandler.ReturnBook)
		borrow.GET("/history", borrowHandler.ListBorrowedBooks)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("CLIENT_PORT")
	}
	e.Logger.Fatal(e.Start(":" + port))
}
