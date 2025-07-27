package benchmark

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yourorg/Go/models"
	"github.com/yourorg/Go/repos"
	"github.com/yourorg/Go/scd"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	for i := 0; i < 10000; i++ {
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
			Versioned: models.Versioned{ID: fmt.Sprintf("tl%d", i), Version: 1, UID: fmt.Sprintf("tl-uid-%d", i)},
			Duration:  8,
			TimeStart: time.Now().Add(-2 * time.Hour),
			TimeEnd:   time.Now().Add(-1 * time.Hour),
			Type:      "work",
			JobUID:    fmt.Sprintf("job-uid-%d", i),
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

// BenchmarkSCDvsRaw compares SCD abstraction vs raw SQL queries
func BenchmarkSCDvsRaw(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)
	jobRepo := repos.JobRepo{DB: db}
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)

	// Test 1: Find Active Jobs by Company - SCD vs Raw
	b.Run("FindActiveJobsByCompany_SCD", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobRepo.FindActiveJobsByCompany("comp1")
		}
	})

	b.Run("FindActiveJobsByCompany_Raw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var jobs []models.Job
			// Raw SQL without SCD abstraction
			db.Raw(`
				SELECT * FROM jobs 
				WHERE status = ? AND company_id = ?
			`, "active", "comp1").Scan(&jobs)
		}
	})

	// Test 2: Find Active Jobs by Contractor - SCD vs Raw
	b.Run("FindActiveJobsByContractor_SCD", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobRepo.FindActiveJobsByContractor("cont1")
		}
	})

	b.Run("FindActiveJobsByContractor_Raw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var jobs []models.Job
			// Raw SQL without SCD abstraction
			db.Raw(`
				SELECT * FROM jobs 
				WHERE status = ? AND contractor_id = ?
			`, "active", "cont1").Scan(&jobs)
		}
	})

	// Test 3: Find Timelogs by Contractor and Period - SCD vs Raw
	b.Run("FindTimelogsByContractorAndPeriod_SCD", func(b *testing.B) {
		timelogRepo := repos.TimelogRepo{DB: db}
		for i := 0; i < b.N; i++ {
			timelogRepo.FindTimelogsByContractorAndPeriod("cont1", from, to)
		}
	})

	b.Run("FindTimelogsByContractorAndPeriod_Raw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var timelogs []models.Timelog
			// Raw SQL without SCD abstraction
			db.Raw(`
				SELECT * FROM timelogs 
				WHERE time_start >= ? AND time_end <= ?
			`, from, to).Scan(&timelogs)
		}
	})

	// Test 4: Find Line Items by Contractor and Period - SCD vs Raw
	b.Run("FindLineItemsByContractorAndPeriod_SCD", func(b *testing.B) {
		pliRepo := repos.PaymentLineItemRepo{DB: db}
		for i := 0; i < b.N; i++ {
			pliRepo.FindLineItemsByContractorAndPeriod("cont1", from, to)
		}
	})

	b.Run("FindLineItemsByContractorAndPeriod_Raw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var lineItems []models.PaymentLineItem
			// Raw SQL without SCD abstraction
			db.Raw(`
				SELECT * FROM payment_line_items 
				WHERE status = ?
			`, "pending").Scan(&lineItems)
		}
	})
}

// BenchmarkSCDAbstraction tests the SCD helper functions directly
func BenchmarkSCDAbstraction(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)

	b.Run("LatestSubquery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scd.LatestSubquery(db, models.Job{})
		}
	})

	b.Run("CreateNewSCDVersion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Create a test job with unique ID for each iteration
			jobID := fmt.Sprintf("test-job-%d", i)
			jobUID := fmt.Sprintf("test-job-uid-%d", i)
			job := models.Job{
				Versioned:    models.Versioned{ID: jobID, Version: 1, UID: jobUID},
				Status:       "active",
				Rate:         100,
				Title:        "Engineer",
				CompanyID:    "comp1",
				ContractorID: "cont1",
			}
			db.Create(&job)

			// Create new version
			scd.CreateNewSCDVersion(db, jobID, func(j *models.Job) {
				j.Status = "completed"
			})
		}
	})
}
