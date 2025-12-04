package repositories

import (
	"context"
	"database/sql"
	"fmt"
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

// ListMerchants retrieves all merchants with their stats
func (r *AdminRepository) ListMerchants(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			m.id,
			m.name,
			m.email,
			m.business_name,
			m.status,
			m.kyc_status,
			m.country,
			m.created_at,
			COALESCE(SUM(CASE WHEN t.status = 'successful' THEN t.amount ELSE 0 END), 0) as total_volume,
			COUNT(t.id) as transaction_count
		FROM merchants m
		LEFT JOIN transactions t ON m.id = t.merchant_id
		GROUP BY m.id, m.name, m.email, m.business_name, m.status, m.kyc_status, m.country, m.created_at
		ORDER BY m.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, name, email, businessName, status, kycStatus, country string
			createdAt                                                   time.Time
			totalVolume, transactionCount                              int64
		)

		err := rows.Scan(
			&id, &name, &email, &businessName, &status, &kycStatus, &country,
			&createdAt, &totalVolume, &transactionCount,
		)
		if err != nil {
			return nil, err
		}

		merchants = append(merchants, map[string]interface{}{
			"id":                 id,
			"name":               name,
			"email":              email,
			"business_name":      businessName,
			"status":             status,
			"kyc_status":         kycStatus,
			"country":            country,
			"created_at":         createdAt,
			"total_volume":       totalVolume,
			"transaction_count":  transactionCount,
		})
	}

	return merchants, rows.Err()
}

// ApproveMerchant approves a merchant's KYC
func (r *AdminRepository) ApproveMerchant(ctx context.Context, id string) error {
	query := `
		UPDATE merchants
		SET kyc_status = 'approved', status = 'active', updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// SuspendMerchant suspends a merchant
func (r *AdminRepository) SuspendMerchant(ctx context.Context, id string) error {
	query := `
		UPDATE merchants
		SET status = 'suspended', updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("merchant not found")
	}

	return nil
}

// GetTransactions retrieves recent transactions
func (r *AdminRepository) GetTransactions(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	query := `
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
			t.created_at
		FROM transactions t
		JOIN merchants m ON t.merchant_id = m.id
		ORDER BY t.created_at DESC
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
			id, reference, merchantID, merchantName, customerEmail, currency, status string
			customerName, paymentMethod                                              sql.NullString
			amount                                                                   int64
			createdAt                                                                time.Time
		)

		err := rows.Scan(
			&id, &reference, &merchantID, &merchantName, &customerEmail,
			&customerName, &amount, &currency, &status, &paymentMethod, &createdAt,
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
			"amount":         amount,
			"currency":       currency,
			"status":         status,
			"created_at":     createdAt,
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

// GetStats retrieves platform statistics
func (r *AdminRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(DISTINCT m.id) as total_merchants,
			COUNT(DISTINCT CASE WHEN m.status = 'active' THEN m.id END) as active_merchants,
			COUNT(DISTINCT CASE WHEN m.kyc_status = 'pending' THEN m.id END) as pending_kyc,
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
		"total_merchants":     totalMerchants,
		"active_merchants":    activeMerchants,
		"pending_kyc":         pendingKYC,
		"total_transactions":  totalTransactions,
		"total_volume":        totalVolume,
		"monthly_volume":      monthlyVolume,
		"success_rate":        0.0,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
	}

	if successRate.Valid {
		stats["success_rate"] = successRate.Float64
	}

	return stats, nil
}
