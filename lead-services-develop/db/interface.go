package db

import (
	"database/sql"

	"gorm.io/gorm"
)
type Database interface {
	Where(query interface{}, args ...interface{}) Database
	First(dest interface{}, conditions ...interface{}) error
	Create(value interface{}) error
	Model(value interface{}) Database
	Count(count *int64) Database
	Error() error
	Save(value interface{}) error
	Preload(column string, conditions ...interface{}) Database
	Delete(value interface{}, conditions ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	AutoMigrate(dst ...interface{}) error
	Offset(offset int) Database
	Limit(limit int) Database
	Scan(dest interface{}) error
	Updates(values ...interface{}) error
	Update(column string, value interface{}) error
	Association(column string) *gorm.Association
	Select(query interface{}) Database
	Joins(query string, args ...interface{}) Database
	Group(query string) Database
	Having(query interface{}, args ...interface{}) Database
	Order(value interface{}) Database
	Scopes(funcs ...func(Database) Database) Database
	Take(dest interface{}, conditions ...interface{}) error
	Last(dest interface{}, conditions ...interface{}) error
	UpdateColumns(values interface{}) error
	CreateInBatches(value interface{}, batchSize int) error
	FindInBatches(dest interface{}, batchSize int, fc func(tx Database, batch int) error) error
	UpdateInBatches(value interface{}, batchSize int) error
	Transaction(fc func(tx Database) error, opts ...*sql.TxOptions) error
	Begin(opts ...*sql.TxOptions) Database
	Commit() error
	Rollback() error
	Raw(sql string, values ...interface{}) Database
	Exec(sql string, values ...interface{}) error
	Pluck(column string, dest interface{}) error
	Migrator() gorm.Migrator
	Distinct(args ...interface{}) Database
	Debug() Database
	Session(config *gorm.Session) Database
	GetUnderlying() *gorm.DB
	Unscoped() Database
}