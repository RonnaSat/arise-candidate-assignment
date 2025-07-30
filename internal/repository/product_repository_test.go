package repository

import (
	"database/sql"
	"gorepositorytest/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func TestPostgresProductRepository_GetAll(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresProductRepository(db)

	t.Run("successful get all products", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Product 1", "Description 1", 10.99, 100, time.Now(), time.Now()).
			AddRow(2, "Product 2", "Description 2", 15.99, 50, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products"`)).
			WillReturnRows(rows)

		products, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(products) != 2 {
			t.Errorf("Expected 2 products, got %d", len(products))
		}

		if products[0].Name != "Product 1" {
			t.Errorf("Expected first product name to be 'Product 1', got %s", products[0].Name)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products"`)).
			WillReturnError(sql.ErrConnDone)

		products, err := repo.GetAll()

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if len(products) != 0 {
			t.Errorf("Expected empty products slice, got %d products", len(products))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresProductRepository_GetByID(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresProductRepository(db)

	t.Run("successful get by id", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Product 1", "Description 1", 10.99, 100, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnRows(row)

		product, err := repo.GetByID(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if product == nil {
			t.Error("Expected product, got nil")
			return
		}

		if product.ID != 1 {
			t.Errorf("Expected product ID to be 1, got %d", product.ID)
		}

		if product.Name != "Product 1" {
			t.Errorf("Expected product name to be 'Product 1', got %s", product.Name)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(999, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		product, err := repo.GetByID(999)

		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
		}

		if product != nil {
			t.Error("Expected nil product, got non-nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresProductRepository_Create(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresProductRepository(db)

	t.Run("successful create", func(t *testing.T) {
		product := &models.Product{
			Name:        "New Product",
			Description: "New Description",
			Price:       25.99,
			Stock:       75,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "products" ("name","description","price","stock","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).
			WithArgs("New Product", "New Description", 25.99, 75, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(product)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if product.ID != 1 {
			t.Errorf("Expected product ID to be set to 1, got %d", product.ID)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("create error", func(t *testing.T) {
		product := &models.Product{
			Name:        "New Product",
			Description: "New Description",
			Price:       25.99,
			Stock:       75,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "products" ("name","description","price","stock","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).
			WithArgs("New Product", "New Description", 25.99, 75, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Create(product)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresProductRepository_Update(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresProductRepository(db)

	t.Run("successful update", func(t *testing.T) {
		product := &models.Product{
			ID:          1,
			Name:        "Updated Product",
			Description: "Updated Description",
			Price:       35.99,
			Stock:       25,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "products" SET "name"=$1,"description"=$2,"price"=$3,"stock"=$4,"created_at"=$5,"updated_at"=$6 WHERE "id" = $7`)).
			WithArgs("Updated Product", "Updated Description", 35.99, 25, sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(product)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("update error", func(t *testing.T) {
		product := &models.Product{
			ID:          1,
			Name:        "Updated Product",
			Description: "Updated Description",
			Price:       35.99,
			Stock:       25,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "products" SET "name"=$1,"description"=$2,"price"=$3,"stock"=$4,"created_at"=$5,"updated_at"=$6 WHERE "id" = $7`)).
			WithArgs("Updated Product", "Updated Description", 35.99, 25, sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Update(product)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresProductRepository_Delete(t *testing.T) {
	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewPostgresProductRepository(db)

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Delete(1)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("delete non-existent product", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(999).
			WillReturnResult(sqlmock.NewResult(1, 0)) // 0 rows affected
		mock.ExpectCommit()

		err := repo.Delete(999)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}
