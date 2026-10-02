package db

import (
	"database/sql"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormDB struct {
	DB *gorm.DB
}

func InitDB() (Database, error) {
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		sslMode,
	)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		return nil, err
	}

	return &GormDB{DB: database}, nil

}
func (g *GormDB) Where(query interface{}, args ...interface{}) Database {
	return &GormDB{DB: g.DB.Where(query, args...)}
}

func (g *GormDB) First(dest interface{}, conditions ...interface{}) error {
	return g.DB.First(dest, conditions...).Error
}

func (g *GormDB) Create(value interface{}) error {
	return g.DB.Create(value).Error
}

func (g *GormDB) Model(value interface{}) Database {
	return &GormDB{DB: g.DB.Model(value)}
}

func (g *GormDB) Count(count *int64) Database {
	return &GormDB{DB: g.DB.Count(count)}
}

func (g *GormDB) Save(value interface{}) error {
	return g.DB.Save(value).Error
}

func (g *GormDB) Preload(column string, conditions ...interface{}) Database {
	return &GormDB{DB: g.DB.Preload(column, conditions...)}
}

func (g *GormDB) Delete(value interface{}, conditions ...interface{}) error {
	return g.DB.Delete(value, conditions...).Error
}

func (g *GormDB) Find(dest interface{}, conds ...interface{}) error {
	return g.DB.Find(dest, conds...).Error
}

func (g *GormDB) AutoMigrate(values ...interface{}) error {
	return g.DB.AutoMigrate(values...)
}

func (g *GormDB) Offset(offset int) Database {
	return &GormDB{DB: g.DB.Offset(offset)}
}

func (g *GormDB) Limit(limit int) Database {
	return &GormDB{DB: g.DB.Limit(limit)}
}

func (g *GormDB) Scan(dest interface{}) error {
	return g.DB.Scan(dest).Error
}

func (g *GormDB) Select(query interface{}) Database {
	return &GormDB{DB: g.DB.Select(query)}
}
func (g *GormDB) Joins(query string, args ...interface{}) Database {
	return &GormDB{DB: g.DB.Joins(query)}
}
func (g *GormDB) Error() error {
	return g.DB.Error
}
func (g *GormDB) Group(query string) Database {
	return &GormDB{DB: g.DB.Group(query)}
}

func (g *GormDB) Updates(values ...interface{}) error {
	return g.DB.Updates(values).Error
}

func (g *GormDB) Update(column string, value interface{}) error {
	return g.DB.Update(column, value).Error
}

func (g *GormDB) Association(column string) *gorm.Association {
	return g.DB.Association(column)
}

func (g *GormDB) Having(query interface{}, args ...interface{}) Database {
	return &GormDB{DB: g.DB.Having(query, args...)}
}

func (g *GormDB) Order(value interface{}) Database {
	return &GormDB{DB: g.DB.Order(value)}
}

func (g *GormDB) Scopes(funcs ...func(Database) Database) Database {
	gormFuncs := make([]func(*gorm.DB) *gorm.DB, len(funcs))
	for i, f := range funcs {
		gormFuncs[i] = func(db *gorm.DB) *gorm.DB {
			return f(&GormDB{DB: db}).(*GormDB).DB
		}
	}
	return &GormDB{DB: g.DB.Scopes(gormFuncs...)}
}
func (g *GormDB) Take(dest interface{}, conditions ...interface{}) error {
	return g.DB.Take(dest, conditions...).Error
}

func (g *GormDB) Last(dest interface{}, conditions ...interface{}) error {
	return g.DB.Last(dest, conditions...).Error
}

func (g *GormDB) UpdateColumns(values interface{}) error {
	return g.DB.UpdateColumns(values).Error
}
func (g *GormDB) CreateInBatches(value interface{}, batchSize int) error {
	return g.DB.CreateInBatches(value, batchSize).Error
}

func (g *GormDB) FindInBatches(dest interface{}, batchSize int, fc func(tx Database, batch int) error) error {
	return g.DB.FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
		return fc(&GormDB{DB: tx}, batch)
	}).Error
}

func (g *GormDB) UpdateInBatches(value interface{}, batchSize int) error {
	return g.DB.Model(value).Updates(value).Error
}
func (g *GormDB) Transaction(fc func(tx Database) error, opts ...*sql.TxOptions) error {
	return g.DB.Transaction(func(tx *gorm.DB) error {
		return fc(&GormDB{DB: tx})
	}, opts...)
}

func (g *GormDB) Begin(opts ...*sql.TxOptions) Database {
	return &GormDB{DB: g.DB.Begin(opts...)}
}
func (g *GormDB) IsNotFound() bool {
	if g.DB == nil || g.DB.Error == nil {
		return false
	}
	return g.DB.Error == gorm.ErrRecordNotFound
}
func (g *GormDB) Commit() error {
	return g.DB.Commit().Error
}

func (g *GormDB) Rollback() error {
	return g.DB.Rollback().Error
}
func (g *GormDB) Raw(sql string, values ...interface{}) Database {
	return &GormDB{DB: g.DB.Raw(sql, values...)}
}

func (g *GormDB) Exec(sql string, values ...interface{}) error {
	return g.DB.Exec(sql, values...).Error
}

func (g *GormDB) Pluck(column string, dest interface{}) error {
	return g.DB.Pluck(column, dest).Error
}
func (g *GormDB) Migrator() gorm.Migrator {
	return g.DB.Migrator()
}
func (g *GormDB) Distinct(args ...interface{}) Database {
	return &GormDB{DB: g.DB.Distinct(args...)}
}
func (g *GormDB) Debug() Database {
	return &GormDB{DB: g.DB.Debug()}
}
func (g *GormDB) Session(config *gorm.Session) Database {
	return &GormDB{DB: g.DB.Session(config)}
}
func (g *GormDB) GetUnderlying() *gorm.DB {
	return g.DB
}
func (g *GormDB) Unscoped() Database {
	return &GormDB{DB: g.DB.Unscoped()}
}
