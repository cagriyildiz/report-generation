package store

import "database/sql"

type Store struct {
	Users             *UserStore
	RefreshTokenStore *RefreshTokenStore
	ReportsStore      *ReportsStore
}

func New(db *sql.DB) *Store {
	return &Store{
		Users:             NewUserStore(db),
		RefreshTokenStore: NewRefreshTokenStore(db),
		ReportsStore:      NewReportsStore(db),
	}
}
