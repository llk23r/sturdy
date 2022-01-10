package organization

import (
	"time"
)

type Organization struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type Member struct {
	ID             string     `db:"id"`
	UserID         string     `db:"user_id"`
	OrganizationID string     `db:"organization_id"`
	CreatedAt      time.Time  `db:"created_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
}
