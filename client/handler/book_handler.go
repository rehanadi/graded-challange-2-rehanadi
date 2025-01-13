package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	pb "book-management-system/pb"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BookHandler struct {
	client pb.BookServiceClient
}

func NewBookHandler(client pb.BookServiceClient) *BookHandler {
	return &BookHandler{client: client}
}

type CreateBookRequest struct {
	Title         string `json:"title" validate:"required"`
	Author        string `json:"author" validate:"required"`
	PublishedDate string `json:"published_date" validate:"required"`
}

func (h *BookHandler) CreateBook(c echo.Context) error {
	var req CreateBookRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get token from context
	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	publishedDate, err := time.Parse("2006-01-02", req.PublishedDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid published date format")
	}
	resp, err := h.client.CreateBook(ctx, &pb.CreateBookRequest{
		Title:         req.Title,
		Author:        req.Author,
		PublishedDate: timestamppb.New(publishedDate),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *BookHandler) GetBook(c echo.Context) error {
	id := c.Param("id")
	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	resp, err := h.client.GetBook(ctx, &pb.GetBookRequest{Id: id})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BookHandler) ListBooks(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage <= 0 {
		perPage = 10
	}

	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	resp, err := h.client.ListBooks(ctx, &pb.ListBooksRequest{
		Page:    int32(page),
		PerPage: int32(perPage),
		Status:  c.QueryParam("status"),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BookHandler) UpdateBook(c echo.Context) error {
	var req CreateBookRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	id := c.Param("id")
	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	publishedDate, err := time.Parse("2006-01-02", req.PublishedDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid published date format")
	}
	resp, err := h.client.UpdateBook(ctx, &pb.UpdateBookRequest{
		Id:            id,
		Title:         req.Title,
		Author:        req.Author,
		PublishedDate: timestamppb.New(publishedDate),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BookHandler) DeleteBook(c echo.Context) error {
	id := c.Param("id")
	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	_, err := h.client.DeleteBook(ctx, &pb.DeleteBookRequest{Id: id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
