package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/yourorg/Go/models"
	"github.com/yourorg/Go/repos"
	"github.com/yourorg/Go/scd"
)

// FocusedSCDImpactTest measures the direct performance impact of SCD abstraction
func BenchmarkFocusedSCDImpact(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)

	fmt.Printf("\n=== SCD Abstraction Impact Analysis ===\n")
	fmt.Printf("Testing with 10,000 records\n\n")

	// Test 1: Simple Job Lookup
	b.Run("Job_Lookup_By_Company", func(b *testing.B) {
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
	b.Run("Job_Lookup_By_Contractor", func(b *testing.B) {
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

	// Test 3: Timelog Lookup
	b.Run("Timelog_Lookup", func(b *testing.B) {
		timelogRepo := repos.TimelogRepo{DB: db}
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now().Add(24 * time.Hour)

		// Test SCD abstraction
		b.Run("SCD_Abstraction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				timelogRepo.FindTimelogsByContractorAndPeriod("cont1", from, to)
			}
		})

		// Test equivalent raw SQL with version handling
		b.Run("Raw_SQL_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var timelogs []models.Timelog
				db.Raw(`
					SELECT t.* FROM timelogs t
					JOIN jobs j ON t.job_uid = j.uid
					JOIN (
						SELECT id, MAX(version) as max_version 
						FROM timelogs 
						GROUP BY id
					) latest ON t.id = latest.id AND t.version = latest.max_version
					WHERE j.contractor_id = ? AND t.time_start >= ? AND t.time_end <= ?
				`, "cont1", from, to).Scan(&timelogs)
			}
		})
	})

	// Test 4: Payment Line Items Lookup
	b.Run("Payment_Line_Items_Lookup", func(b *testing.B) {
		pliRepo := repos.PaymentLineItemRepo{DB: db}
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now().Add(24 * time.Hour)

		// Test SCD abstraction
		b.Run("SCD_Abstraction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pliRepo.FindLineItemsByContractorAndPeriod("cont1", from, to)
			}
		})

		// Test equivalent raw SQL with version handling
		b.Run("Raw_SQL_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var lineItems []models.PaymentLineItem
				db.Raw(`
					SELECT pli.* FROM payment_line_items pli
					JOIN timelogs t ON pli.timelog_uid = t.uid
					JOIN jobs j ON pli.job_uid = j.uid
					JOIN (
						SELECT id, MAX(version) as max_version 
						FROM payment_line_items 
						GROUP BY id
					) latest ON pli.id = latest.id AND pli.version = latest.max_version
					WHERE j.contractor_id = ? AND t.time_start >= ? AND t.time_end <= ?
				`, "cont1", from, to).Scan(&lineItems)
			}
		})
	})
}

// BenchmarkSCDCoreOperations measures the core SCD operations
func BenchmarkSCDCoreOperations(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)

	fmt.Printf("\n=== SCD Core Operations Performance ===\n\n")

	// Test 1: Latest Subquery Performance
	b.Run("Latest_Subquery_Performance", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scd.LatestSubquery(db, models.Job{})
		}
	})

	// Test 2: Create New Version Performance
	b.Run("Create_New_Version_Performance", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobID := fmt.Sprintf("impact-test-job-%d", i)
			jobUID := fmt.Sprintf("impact-test-job-uid-%d", i)

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

			// Create new version
			scd.CreateNewSCDVersion(db, jobID, func(j *models.Job) {
				j.Status = "completed"
				j.Rate = 150
			})
		}
	})
}

// BenchmarkQueryComplexityImpact measures the complexity overhead
func BenchmarkQueryComplexityImpact(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)

	fmt.Printf("\n=== Query Complexity Impact Analysis ===\n\n")

	// Test 1: Simple vs Complex SCD Query
	b.Run("Simple_vs_Complex_SCD", func(b *testing.B) {
		// Simple SCD query (just the abstraction)
		b.Run("Simple_SCD_Query", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				scd.LatestSubquery(db, models.Job{})
			}
		})

		// Complex SCD query (with actual data retrieval)
		b.Run("Complex_SCD_Query", func(b *testing.B) {
			jobRepo := repos.JobRepo{DB: db}
			for i := 0; i < b.N; i++ {
				jobRepo.FindActiveJobsByCompany("comp1")
			}
		})
	})

	// Test 2: SCD Join Complexity
	b.Run("SCD_Join_Complexity", func(b *testing.B) {
		// Test the actual SCD JOIN query
		b.Run("SCD_JOIN_Query", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var jobs []models.Job
				subq := scd.LatestSubquery(db, models.Job{})
				db.Model(&models.Job{}).
					Joins("JOIN (?) AS latest ON jobs.id = latest.id AND jobs.version = latest.max_version", subq).
					Where("jobs.status = ? AND jobs.company_id = ?", "active", "comp1").
					Find(&jobs)
			}
		})

		// Test equivalent without SCD
		b.Run("Direct_Query_Equivalent", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var jobs []models.Job
				db.Raw(`SELECT * FROM jobs WHERE status = ? AND company_id = ?`, "active", "comp1").Scan(&jobs)
			}
		})
	})
}

// BenchmarkMemoryImpact measures memory usage differences
func BenchmarkMemoryImpact(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)

	fmt.Printf("\n=== Memory Impact Analysis ===\n\n")

	// Test memory usage for SCD vs Raw queries
	b.Run("Memory_Usage_Comparison", func(b *testing.B) {
		jobRepo := repos.JobRepo{DB: db}

		// Measure SCD memory usage
		b.Run("SCD_Memory_Usage", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jobRepo.FindActiveJobsByCompany("comp1")
			}
		})

		// Measure Raw SQL memory usage
		b.Run("Raw_SQL_Memory_Usage", func(b *testing.B) {
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
}

// BenchmarkSCDOverhead calculates the actual overhead percentage
func BenchmarkSCDOverhead(b *testing.B) {
	db := setupDB(b)
	seedMillion(db)

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

			// Calculate overhead as percentage
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
