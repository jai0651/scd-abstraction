package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourorg/Go/models"
	"github.com/yourorg/Go/repos"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=scd port=5432 sslmode=disable"
	}
	// Connect to DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// AutoMigrate
	db.AutoMigrate(&models.Job{}, &models.Timelog{}, &models.PaymentLineItem{})

	// Seed sample data
	seedData(db)

	// Repos
	jobRepo := repos.JobRepo{DB: db}
	timelogRepo := repos.TimelogRepo{DB: db}
	pliRepo := repos.PaymentLineItemRepo{DB: db}

	// Demo queries
	fmt.Println("Active jobs for company comp1:")
	jobs, _ := jobRepo.FindActiveJobsByCompany("comp1")
	for _, j := range jobs {
		fmt.Printf("%+v\n", j)
	}

	fmt.Println("Active jobs for contractor cont1:")
	jobs, _ = jobRepo.FindActiveJobsByContractor("cont1")
	for _, j := range jobs {
		fmt.Printf("%+v\n", j)
	}

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)

	fmt.Println("Timelogs for contractor cont1 in period:")
	timelogs, _ := timelogRepo.FindTimelogsByContractorAndPeriod("cont1", from, to)
	for _, t := range timelogs {
		fmt.Printf("%+v\n", t)
	}

	fmt.Println("Payment line items for contractor cont1 in period:")
	items, _ := pliRepo.FindLineItemsByContractorAndPeriod("cont1", from, to)
	for _, i := range items {
		fmt.Printf("%+v\n", i)
	}
}

func seedData(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE payment_line_items, timelogs, jobs RESTART IDENTITY CASCADE")
	// Seed jobs
	job := models.Job{
		Versioned:    models.Versioned{ID: "job1", Version: 1, UID: "job-uid-1"},
		Status:       "active",
		Rate:         100,
		Title:        "Engineer",
		CompanyID:    "comp1",
		ContractorID: "cont1",
	}
	db.Create(&job)
	// Seed timelog
	timelog := models.Timelog{
		Versioned: models.Versioned{ID: "tl1", Version: 1, UID: "tl-uid-1"},
		Duration:  8,
		TimeStart: time.Now().Add(-2 * time.Hour),
		TimeEnd:   time.Now().Add(-1 * time.Hour),
		Type:      "work",
		JobUID:    "job-uid-1",
	}
	db.Create(&timelog)
	// Seed payment line item
	pli := models.PaymentLineItem{
		Versioned:  models.Versioned{ID: "pli1", Version: 1, UID: "pli-uid-1"},
		JobUID:     "job-uid-1",
		TimelogUID: "tl-uid-1",
		Amount:     800,
		Status:     "pending",
	}
	db.Create(&pli)
}
