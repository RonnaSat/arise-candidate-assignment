package repository

import (
	"gorepositorytest/internal/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetAll() ([]models.Product, error)
	GetByID(id uint) (*models.Product, error)
	Create(product *models.Product) error
	Update(product *models.Product) error
	Delete(id uint) error
}

type postgresProductRepository struct {
	db *gorm.DB
}

func NewPostgresProductRepository(db *gorm.DB) ProductRepository {
	return &postgresProductRepository{db: db}
}

func (r *postgresProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product
	err := r.db.Find(&products).Error
	return products, err
}

func (r *postgresProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *postgresProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *postgresProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *postgresProductRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}
