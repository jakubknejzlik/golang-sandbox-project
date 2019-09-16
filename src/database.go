package src

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// NewDBFromEnvVars Create database client using DATABASE_URL environment variable
func NewDBFromEnvVars() *gorm.DB {
	urlString := os.Getenv("DATABASE_URL")
	if urlString == "" {
		panic(fmt.Errorf("missing DATABASE_URL environment variable"))
	}
	return NewDBWithString(urlString)
}

// NewDBWithString Create database instance with database URL string
func NewDBWithString(urlString string) *gorm.DB {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	urlString = getConnectionString(u)

	db, err := gorm.Open(u.Scheme, urlString)
	if err != nil {
		panic(err)
	}
	if urlString == "sqlite3://:memory:" {
		db.DB().SetMaxIdleConns(1)
		db.DB().SetConnMaxLifetime(time.Second * 300)
		db.DB().SetMaxOpenConns(1)
	} else {
		db.DB().SetMaxIdleConns(5)
		db.DB().SetConnMaxLifetime(time.Second * 60)
		db.DB().SetMaxOpenConns(10)
	}
	db.LogMode(os.Getenv("DEBUG") == "true")

	return db
}

// Database URL string unifier for postgres/mysql/sqlite
func getConnectionString(u *url.URL) string {
	if u.Scheme == "postgres" {
		password, _ := u.User.Password()
		host := strings.Split(u.Host, ":")[0]
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, u.Port(), u.User.Username(), password, strings.TrimPrefix(u.Path, "/"))
	}
	if u.Scheme != "sqlite3" {
		u.Host = "tcp(" + u.Host + ")"
	}
	if u.Scheme == "mysql" {
		q := u.Query()
		q.Set("parseTime", "true")
		u.RawQuery = q.Encode()
	}
	return strings.Replace(u.String(), u.Scheme+"://", "", 1)
}
