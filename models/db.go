package models

import "time"

type DBConfig struct {
	URL             string
	MaxOpenConns    int           // maximum number of open connections to the database
	MaxIdleConns    int           // maximum number of connections in the idle connection pool
	ConnMaxLifetime time.Duration // maximum amount of time a connection may be reused
	ConnMaxIdleTime time.Duration // maximum amount of time a connection may be idle
}
