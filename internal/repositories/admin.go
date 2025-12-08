package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(dsn string) (*AdminRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &AdminRepository{db: db}, nil
}

func (r *AdminRepository) Close() error {
	return r.db.Close()
}

// ListMerchants retrieves merchants with basic fields for the admin portal
func (r *AdminRepository) ListMerchants(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT
			m.id,
			m.name,
			m.email,
			m.business_name,
			m.status,
			m.kyc_status,
			m.created_at,
			m.updated_at,
			COALESCE(mb.total_volume, 0) as total_volume,
			COALESCE(mb.currency, 'NGN') as currency
		FROM merchants m
		LEFT JOIN merchant_balances mb ON m.id = mb.merchant_id
		ORDER BY m.created_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merchants []map[string]interface{}
	for rows.Next() {
		var (
			id                                                         int // Changed from string to int
			name, email, businessName, status, kycStatus, currency string
			createdAt, updatedAt                                       time.Time
			totalVolume                                                int64
		)
		if err := rows.Scan(&id, &name, &email, &businessName, &status, &kycStatus, &createdAt, &updatedAt, &totalVolume, &currency); err != nil {
			return nil, err
		}
		merchants = append(merchants, map[string]interface{}{
			"id":            id, // Changed to int
			"name":          name,
			"email":         email,
			"business_name": businessName,
			"status":        status,
			"kyc_status":    kycStatus,
			"created_at":    createdAt,
			"updated_at":    updatedAt,
			"total_volume":  totalVolume,
			"currency":      currency,
		})
	}
	return merchants, rows.Err()
}

// UpdateMerchantStatus updates the merchant status
func (r *AdminRepository) UpdateMerchantStatus(ctx context.Context, id int, status string) error { // Changed id from string to int
	log.Printf("Attempting to update merchant status for ID: %d to status: %s", id, status) // Changed %s to %d
	query := `UPDATE merchants SET status = $2, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id, status) // id is int
	if err != nil {
		log.Printf("Error updating merchant status for ID %d: %v", id, err) // Changed %s to %d
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected after updating merchant status for ID %d: %v", id, err) // Changed %s to %d
		return err
	}
	if affected == 0 {
		log.Printf("No merchant found with ID: %d or status was already %s", id, status) // Changed %s to %d
		return fmt.Errorf("merchant not found or status already %s", status)
	}
	log.Printf("Successfully updated merchant status for ID: %d to status: %s. Rows affected: %d", id, status, affected) // Changed %s to %d
	return nil
}

// GetStats retrieves platform statistics
func (r *AdminRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(DISTINCT m.id) as total_merchants,
			COUNT(DISTINCT CASE WHEN m.status = 'active' THEN m.id END) as active_merchants,
			COUNT(DISTINCT CASE WHEN m.kyc_status IN ('pending', 'not_started') OR m.status = 'inactive' THEN m.id END) as pending_kyc,
			COUNT(t.id) as total_transactions,
			COALESCE(SUM(CASE WHEN t.status = 'successful' THEN t.amount ELSE 0 END), 0) as total_volume,
			COALESCE(SUM(CASE WHEN t.status = 'successful' AND t.created_at >= NOW() - INTERVAL '30 days' THEN t.amount ELSE 0 END), 0) as monthly_volume,
			COUNT(CASE WHEN t.status = 'successful' THEN 1 END)::float / NULLIF(COUNT(t.id), 0) * 100 as success_rate
		FROM merchants m
		LEFT JOIN transactions t ON m.id = t.merchant_id
	`

	var (
		totalMerchants, activeMerchants, pendingKYC, totalTransactions int
		totalVolume, monthlyVolume                                     int64
		successRate                                                    sql.NullFloat64
	)

	err := r.db.QueryRowContext(ctx, query).Scan(
		&totalMerchants, &activeMerchants, &pendingKYC,
		&totalTransactions, &totalVolume, &monthlyVolume, &successRate,
	)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_merchants":    totalMerchants,
		"active_merchants":   activeMerchants,
		"pending_kyc":        pendingKYC,
		"total_transactions": totalTransactions,
		"total_volume":       totalVolume,
		"monthly_volume":     monthlyVolume,
		"success_rate":       0.0,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
	}

	if successRate.Valid {
		stats["success_rate"] = successRate.Float64
	}

	return stats, nil
}

// GetTransactions retrieves recent transactions
func (r *AdminRepository) GetTransactions(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT * FROM (
			SELECT
				t.id,
				t.reference,
				t.merchant_id,
				m.business_name as merchant_name,
				t.customer_email,
				t.customer_name,
				t.amount,
				t.currency,
				t.status,
				t.payment_method,
				t.created_at,
				'payment' as type
			FROM transactions t
			JOIN merchants m ON t.merchant_id = m.id
			
			UNION ALL
			
			SELECT
				p.id,
				p.reference,
				p.merchant_id,
				m.business_name as merchant_name,
				'' as customer_email,
				'' as customer_name,
				p.amount,
				p.currency,
				p.status,
				'payout' as payment_method,
				p.created_at,
				'payout' as type
			FROM payouts p
			JOIN merchants m ON p.merchant_id = m.id
		) AS combined
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, merchantID                                                                    int // Changed from string to int
			reference, merchantName, customerEmail, currency, status, txnType string
			customerName, paymentMethod                                                       sql.NullString
			amount                                                                            int64
			createdAt                                                                         time.Time
		)

		err := rows.Scan(
			&id, &reference, &merchantID, &merchantName, &customerEmail,
			&customerName, &amount, &currency, &status, &paymentMethod, &createdAt, &txnType,
		)
		if err != nil {
			return nil, err
		}

		transaction := map[string]interface{}{
			"id":             id,
			"reference":      reference,
			"merchant_id":    merchantID,
			"merchant_name":  merchantName,
			"customer_email": customerEmail,
			"amount":         float64(amount) / 100, // return currency units
			"currency":       currency,
			"status":         status,
			"created_at":     createdAt,
			"type":           txnType,
		}

		if customerName.Valid {
			transaction["customer_name"] = customerName.String
		}
		if paymentMethod.Valid {
			transaction["payment_method"] = paymentMethod.String
		}

		transactions = append(transactions, transaction)
	}

	return transactions, rows.Err()
}
