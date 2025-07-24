package benchmark

import (
	"testing"
	"time"
	"github.com/yourorg/scd-abstraction/models"
	"github.com/yourorg/scd-abstraction/repos"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"fmt"
)

func setupDB(b *testing.B) *gorm.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=scd port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.Job{}, &models.Timelog{}, &models.PaymentLineItem{})
	return db
}

func seedMillion(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE payment_line_items, timelogs, jobs RESTART IDENTITY CASCADE")
	for i := 0; i < 100000; i++ {
		job := models.Job{
			Versioned:    models.Versioned{ID: fmt.Sprintf("job%d", i), Version: 1, UID: fmt.Sprintf("job-uid-%d", i)},
			Status:       "active",
			Rate:         100,
			Title:        "Engineer",
			CompanyID:    "comp1",
			ContractorID: "cont1",
		}
		db.Create(&job)
		timelog := models.Timelog{
			Versioned:  models.Versioned{ID: fmt.Sprintf("tl%d", i), Version: 1, UID: fmt.Sprintf("tl-uid-%d", i)},
			Duration:   8,
			TimeStart:  time.Now().Add(-2 * time.Hour),
			TimeEnd:    time.Now().Add(-1 * time.Hour),
			Type:       "work",
			JobUID:     fmt.Sprintf("job-uid-%d", i),
		}
		db.Create(&timelog)
		pli := models.PaymentLineItem{
			Versioned:  models.Versioned{ID: fmt.Sprintf("pli%d", i), Version: 1, UID: fmt.Sprintf("pli-uid-%d", i)},
			JobUID:     fmt.Sprintf("job-uid-%d", i),
			TimelogUID: fmt.Sprintf("tl-uid-%d", i),
			Amount:     800,
			Status:     "pending",
		}
		db.Create(&pli)
	}
}

func BenchmarkRepoQueries(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)
	jobRepo := repos.JobRepo{DB: db}
	timelogRepo := repos.TimelogRepo{DB: db}
	pliRepo := repos.PaymentLineItemRepo{DB: db}
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)

	b.Run("FindActiveJobsByCompany", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobRepo.FindActiveJobsByCompany("comp1")
		}
	})
	b.Run("FindActiveJobsByContractor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobRepo.FindActiveJobsByContractor("cont1")
		}
	})
	b.Run("FindTimelogsByContractorAndPeriod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			timelogRepo.FindTimelogsByContractorAndPeriod("cont1", from, to)
		}
	})
	b.Run("FindLineItemsByContractorAndPeriod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pliRepo.FindLineItemsByContractorAndPeriod("cont1", from, to)
		}
	})
} 