package handler

import (
	"context"
	"net/http"
	"strconv"

	pb "book-management-system/pb"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/metadata"
)

type BorrowHandler struct {
	client pb.BorrowServiceClient
}

func NewBorrowHandler(client pb.BorrowServiceClient) *BorrowHandler {
	return &BorrowHandler{client: client}
}

type BorrowBookRequest struct {
	BookID string `json:"book_id" validate:"required"`
}

func (h *BorrowHandler) BorrowBook(c echo.Context) error {
	var req BorrowBookRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	resp, err := h.client.BorrowBook(ctx, &pb.BorrowBookRequest{
		BookId: req.BookID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BorrowHandler) ReturnBook(c echo.Context) error {
	borrowID := c.Param("id")
	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	resp, err := h.client.ReturnBook(ctx, &pb.ReturnBookRequest{
		BorrowId: borrowID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BorrowHandler) ListBorrowedBooks(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage <= 0 {
		perPage = 10
	}
	includeReturned, _ := strconv.ParseBool(c.QueryParam("include_returned"))

	token := c.Get("token").(string)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	resp, err := h.client.ListBorrowedBooks(ctx, &pb.ListBorrowedBooksRequest{
		Page:            int32(page),
		PerPage:         int32(perPage),
		IncludeReturned: includeReturned,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}
