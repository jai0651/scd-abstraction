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

func setupSimpleDB(b *testing.B) *gorm.DB {
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

func seedSimpleData(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE payment_line_items, timelogs, jobs RESTART IDENTITY CASCADE")

	// Create a smaller dataset for faster benchmarking
	for i := 0; i < 1000; i++ {
		job := models.Job{
			Versioned:    models.Versioned{ID: fmt.Sprintf("job%d", i), Version: 1, UID: fmt.Sprintf("job-uid-%d", i)},
			Status:       "active",
			Rate:         100,
			Title:        "Engineer",
			CompanyID:    "comp1",
			ContractorID: "cont1",
		}
		db.Create(&job)
	}
}

// BenchmarkSimpleSCDImpact measures the basic SCD abstraction overhead
func BenchmarkSimpleSCDImpact(b *testing.B) {
	db := setupSimpleDB(b)
	seedSimpleData(db)

	fmt.Printf("\n=== Simple SCD Impact Analysis ===\n")
	fmt.Printf("Testing with 1,000 records\n\n")

	// Test 1: Job Lookup by Company
	b.Run("Job_By_Company", func(b *testing.B) {
		jobRepo := repos.JobRepo{DB: db}

		// Test SCD abstraction
		b.Run("SCD_Abstraction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jobRepo.FindActiveJobsByCompany("comp1")
			}
		})

		// Test equivalent raw SQL with version handling
		b.Run("Raw_SQL_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var jobs []models.Job
				db.Raw(`
					SELECT j.* FROM jobs j
					JOIN (
						SELECT id, MAX(version) as max_version 
						FROM jobs 
						GROUP BY id
					) latest ON j.id = latest.id AND j.version = latest.max_version
					WHERE j.status = ? AND j.company_id = ?
				`, "active", "comp1").Scan(&jobs)
			}
		})
	})

	// Test 2: Job Lookup by Contractor
	b.Run("Job_By_Contractor", func(b *testing.B) {
		jobRepo := repos.JobRepo{DB: db}

		// Test SCD abstraction
		b.Run("SCD_Abstraction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jobRepo.FindActiveJobsByContractor("cont1")
			}
		})

		// Test equivalent raw SQL with version handling
		b.Run("Raw_SQL_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var jobs []models.Job
				db.Raw(`
					SELECT j.* FROM jobs j
					JOIN (
						SELECT id, MAX(version) as max_version 
						FROM jobs 
						GROUP BY id
					) latest ON j.id = latest.id AND j.version = latest.max_version
					WHERE j.status = ? AND j.contractor_id = ?
				`, "active", "cont1").Scan(&jobs)
			}
		})
	})
}

// BenchmarkSimpleSCDCoreOperations measures the core SCD operations
func BenchmarkSimpleSCDCoreOperations(b *testing.B) {
	db := setupSimpleDB(b)
	seedSimpleData(db)

	fmt.Printf("\n=== SCD Core Operations Performance ===\n\n")

	// Test 1: Latest Subquery Performance (SCD vs Raw SQL)
	b.Run("Latest_Subquery", func(b *testing.B) {
		// Test SCD abstraction
		b.Run("SCD_Abstraction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scd.LatestSubquery(db, models.Job{})
			}
		})

		// Test equivalent raw SQL
		b.Run("Raw_SQL_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				db.Raw(`SELECT id, MAX(version) as max_version FROM jobs GROUP BY id`)
			}
		})
	})

	// Test 2: Create New Version Performance (SCD vs Raw SQL)
	b.Run("Create_New_Version", func(b *testing.B) {
		// Test SCD abstraction
		b.Run("SCD_Abstraction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jobID := fmt.Sprintf("simple-test-job-%d", i)
				jobUID := fmt.Sprintf("simple-test-job-uid-%d", i)

				// Create initial version
				job := models.Job{
					Versioned:    models.Versioned{ID: jobID, Version: 1, UID: jobUID},
					Status:       "active",
					Rate:         100,
					Title:        "Engineer",
					CompanyID:    "comp1",
					ContractorID: "cont1",
				}
				db.Create(&job)

				// Create new version using SCD abstraction
				scd.CreateNewSCDVersion(db, jobID, func(j *models.Job) {
					j.Status = "completed"
					j.Rate = 150
				})
			}
		})

		// Test equivalent raw SQL
		b.Run("Raw_SQL_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jobID := fmt.Sprintf("simple-test-job-raw-%d", i)
				jobUID := fmt.Sprintf("simple-test-job-uid-raw-%d", i)

				// Create initial version using raw SQL
				db.Exec(`
					INSERT INTO jobs (id, version, uid, status, rate, title, company_id, contractor_id) 
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				`, jobID, 1, jobUID, "active", 100, "Engineer", "comp1", "cont1")

				// Create new version using raw SQL (equivalent to SCD abstraction)
				// Generate a unique UID for the new version to avoid conflicts
				newUID := fmt.Sprintf("simple-test-job-uid-raw-new-%d", i)
				db.Exec(`
					INSERT INTO jobs (id, version, uid, status, rate, title, company_id, contractor_id)
					SELECT id, MAX(version) + 1, ?, 'completed', 150, title, company_id, contractor_id
					FROM jobs 
					WHERE id = ?
					GROUP BY id, title, company_id, contractor_id
				`, newUID, jobID)
			}
		})
	})
}

// BenchmarkSimpleSCDOverhead calculates the actual overhead percentage
func BenchmarkSimpleSCDOverhead(b *testing.B) {
	db := setupSimpleDB(b)
	seedSimpleData(db)

	fmt.Printf("\n=== SCD Overhead Calculation ===\n\n")

	// Test different query types and calculate overhead
	testCases := []struct {
		name     string
		scdQuery func()
		rawQuery func()
	}{
		{
			name: "Job_By_Company",
			scdQuery: func() {
				jobRepo := repos.JobRepo{DB: db}
				jobRepo.FindActiveJobsByCompany("comp1")
			},
			rawQuery: func() {
				var jobs []models.Job
				db.Raw(`
					SELECT j.* FROM jobs j
					JOIN (
						SELECT id, MAX(version) as max_version 
						FROM jobs 
						GROUP BY id
					) latest ON j.id = latest.id AND j.version = latest.max_version
					WHERE j.status = ? AND j.company_id = ?
				`, "active", "comp1").Scan(&jobs)
			},
		},
		{
			name: "Job_By_Contractor",
			scdQuery: func() {
				jobRepo := repos.JobRepo{DB: db}
				jobRepo.FindActiveJobsByContractor("cont1")
			},
			rawQuery: func() {
				var jobs []models.Job
				db.Raw(`
					SELECT j.* FROM jobs j
					JOIN (
						SELECT id, MAX(version) as max_version 
						FROM jobs 
						GROUP BY id
					) latest ON j.id = latest.id AND j.version = latest.max_version
					WHERE j.status = ? AND j.contractor_id = ?
				`, "active", "cont1").Scan(&jobs)
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Measure SCD performance
			scdStart := time.Now()
			for i := 0; i < b.N; i++ {
				tc.scdQuery()
			}
			scdDuration := time.Since(scdStart)

			// Measure Raw SQL performance
			rawStart := time.Now()
			for i := 0; i < b.N; i++ {
				tc.rawQuery()
			}
			rawDuration := time.Since(rawStart)

			// Calculate overhead
			overheadPercent := float64(scdDuration-rawDuration) / float64(rawDuration) * 100

			fmt.Printf("%s:\n", tc.name)
			fmt.Printf("  SCD Time: %v\n", scdDuration)
			fmt.Printf("  Raw Time: %v\n", rawDuration)
			fmt.Printf("  Overhead: %.2f%%\n", overheadPercent)

			if overheadPercent > 50 {
				fmt.Printf("  ⚠️  High overhead (>50%%)\n")
			} else if overheadPercent > 20 {
				fmt.Printf("  ⚠️  Moderate overhead (20-50%%)\n")
			} else {
				fmt.Printf("  ✅ Acceptable overhead (<20%%)\n")
			}
			fmt.Printf("\n")
		})
	}
}
