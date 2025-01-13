package scheduler

import (
	"log"
	"time"

	"book-management-system/internal/models"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type BookScheduler struct {
	db   *gorm.DB
	cron *cron.Cron
}

func NewBookScheduler(db *gorm.DB) *BookScheduler {
	return &BookScheduler{
		db:   db,
		cron: cron.New(),
	}
}

func (s *BookScheduler) Start() error {
	// Check for overdue books every hour
	_, err := s.cron.AddFunc("@hourly", s.checkOverdueBooks)
	if err != nil {
		return err
	}

	s.cron.Start()
	return nil
}

func (s *BookScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *BookScheduler) checkOverdueBooks() {
	log.Println("Running overdue books check...")

	// Find all borrowed books that are overdue (more than 14 days)
	overdueDate := time.Now().AddDate(0, 0, -14)

	var overdueBooks []models.BorrowedBook

	result := s.db.Where("return_date IS NULL AND borrowed_date < ?", overdueDate).
		Preload("Book").
		Find(&overdueBooks)

	if result.Error != nil {
		log.Printf("Error finding overdue books: %v", result.Error)
		return
	}

	// Update status of overdue books
	for _, borrowed := range overdueBooks {
		err := s.db.Model(&models.Book{}).
			Where("id = ?", borrowed.BookID).
			Update("status", "Overdue").Error

		if err != nil {
			log.Printf("Error updating book status to overdue (ID: %s): %v", borrowed.BookID, err)
			continue
		}

		log.Printf("Updated book status to overdue: %s", borrowed.BookID)
	}

	log.Printf("Completed overdue check. Found %d overdue books", len(overdueBooks))
}

// Additional utility functions
func (s *BookScheduler) ManualCheckOverdue() error {
	s.checkOverdueBooks()
	return nil
}
