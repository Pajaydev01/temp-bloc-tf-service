package exception

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

func HandleDBError(err error) error {
	if err == nil {
		return nil
	}

	// 1️⃣ Check for "Record Not Found"
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("record not found")
	}

	// 2️⃣ Check for MySQL-specific errors
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case 1062:
			return errors.New("duplicate entry: this record already exists") // Unique constraint
		case 1452:
			return errors.New("foreign key constraint failed") // FK violation
		case 1406:
			return errors.New("data too long for column") // Exceeding max length
		}
	}

	// 3️⃣ Check for PostgreSQL-specific errors
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505":
			return errors.New("duplicate entry: this record already exists") // Unique constraint
		case "23503":
			return errors.New("foreign key constraint failed") // FK violation
		case "23502":
			return errors.New("null value in column violates NOT NULL constraint") // Not Null
		case "22001":
			return errors.New("data too long for column") // Exceeding max length
		}
	}

	// 4️⃣ Default: Return the original error
	return fmt.Errorf("database error: %v", err)
}
