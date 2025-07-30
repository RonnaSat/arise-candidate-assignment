package repository

import (
	"database/sql"
	"gorepositorytest/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestPostgresOrderRepository_GetAll(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresOrderRepository(db)

	t.Run("successful get all orders", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "transaction_id", "total_amount", "status", "created_at", "updated_at"}).
			AddRow(1, "TXN001", 99.99, "pending", time.Now(), time.Now()).
			AddRow(2, "TXN002", 149.99, "confirmed", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders"`)).
			WillReturnRows(rows)

		orders, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(orders) != 2 {
			t.Errorf("Expected 2 orders, got %d", len(orders))
		}

		if orders[0].TransactionID != "TXN001" {
			t.Errorf("Expected first order transaction ID to be 'TXN001', got %s", orders[0].TransactionID)
		}

		if orders[1].Status != "confirmed" {
			t.Errorf("Expected second order status to be 'confirmed', got %s", orders[1].Status)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders"`)).
			WillReturnError(sql.ErrConnDone)

		orders, err := repo.GetAll()

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if len(orders) != 0 {
			t.Errorf("Expected empty orders slice, got %d orders", len(orders))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresOrderRepository_GetByTransactionID(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresOrderRepository(db)

	t.Run("successful get by transaction id", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "transaction_id", "total_amount", "status", "created_at", "updated_at"}).
			AddRow(1, "TXN001", 99.99, "pending", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE transaction_id = $1 ORDER BY "orders"."id" LIMIT $2`)).
			WithArgs("TXN001", 1).
			WillReturnRows(row)

		order, err := repo.GetByTransactionID("TXN001")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if order == nil {
			t.Error("Expected order, got nil")
			return
		}

		if order.TransactionID != "TXN001" {
			t.Errorf("Expected order transaction ID to be 'TXN001', got %s", order.TransactionID)
		}

		if order.TotalAmount != 99.99 {
			t.Errorf("Expected order total amount to be 99.99, got %f", order.TotalAmount)
		}

		if order.Status != "pending" {
			t.Errorf("Expected order status to be 'pending', got %s", order.Status)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("order not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE transaction_id = $1 ORDER BY "orders"."id" LIMIT $2`)).
			WithArgs("NONEXISTENT", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		order, err := repo.GetByTransactionID("NONEXISTENT")

		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
		}

		if order != nil {
			t.Error("Expected nil order, got non-nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresOrderRepository_Create(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresOrderRepository(db)

	t.Run("successful create", func(t *testing.T) {
		order := &models.Order{
			TransactionID: "TXN003",
			TotalAmount:   199.99,
			Status:        "pending",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders" ("transaction_id","total_amount","status","created_at","updated_at") VALUES ($1,$2,$3,$4,$5) RETURNING "id"`)).
			WithArgs("TXN003", 199.99, "pending", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(order)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if order.ID != 1 {
			t.Errorf("Expected order ID to be set to 1, got %d", order.ID)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("create error", func(t *testing.T) {
		order := &models.Order{
			TransactionID: "TXN004",
			TotalAmount:   299.99,
			Status:        "pending",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders" ("transaction_id","total_amount","status","created_at","updated_at") VALUES ($1,$2,$3,$4,$5) RETURNING "id"`)).
			WithArgs("TXN004", 299.99, "pending", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Create(order)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresOrderRepository_Update(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresOrderRepository(db)

	t.Run("successful update", func(t *testing.T) {
		order := &models.Order{
			ID:            1,
			TransactionID: "TXN001",
			TotalAmount:   149.99,
			Status:        "confirmed",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "transaction_id"=$1,"total_amount"=$2,"status"=$3,"created_at"=$4,"updated_at"=$5 WHERE "id" = $6`)).
			WithArgs("TXN001", 149.99, "confirmed", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(order)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("update error", func(t *testing.T) {
		order := &models.Order{
			ID:            1,
			TransactionID: "TXN001",
			TotalAmount:   149.99,
			Status:        "confirmed",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "transaction_id"=$1,"total_amount"=$2,"status"=$3,"created_at"=$4,"updated_at"=$5 WHERE "id" = $6`)).
			WithArgs("TXN001", 149.99, "confirmed", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Update(order)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresOrderRepository_UpdateStatus(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresOrderRepository(db)

	t.Run("successful status update", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"updated_at"=$2 WHERE id = $3`)).
			WithArgs("shipped", sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(1, "shipped")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("status update error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"updated_at"=$2 WHERE id = $3`)).
			WithArgs("delivered", sqlmock.AnyArg(), 2).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.UpdateStatus(2, "delivered")

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("update status for non-existent order", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders" SET "status"=$1,"updated_at"=$2 WHERE id = $3`)).
			WithArgs("cancelled", sqlmock.AnyArg(), 999).
			WillReturnResult(sqlmock.NewResult(1, 0)) // 0 rows affected
		mock.ExpectCommit()

		err := repo.UpdateStatus(999, "cancelled")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}
