package repositories

import "log"

type AdminRepository struct {
    dsn string
}

func NewAdminRepository(dsn string) *AdminRepository {
    log.Printf("AdminRepository using DSN: %s", dsn)
    return &AdminRepository{dsn: dsn}
}

// TODO: implement persistence for admin views, stats, approvals.
