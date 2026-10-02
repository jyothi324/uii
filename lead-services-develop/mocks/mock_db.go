package mocks

import (
	"database/sql"

	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/db"
	"gorm.io/gorm"
)

// MockDatabase implements db.Database for tests.
type MockDatabase struct {
	mock.Mock
}

// helper to call m.Called with variadic expansion
func call(m *mock.Mock, args ...interface{}) mock.Arguments {
	return m.Called(args...)
}

// Utility to expand a fixed arg + variadic slice into one []interface{}
func buildArgs(first interface{}, rest []interface{}) []interface{} {
	res := make([]interface{}, 0, 1+len(rest))
	res = append(res, first)
	res = append(res, rest...)
	return res
}

// Utility to expand only variadic slice (when no fixed first arg)
func buildArgsFromSlice(rest []interface{}) []interface{} {
	res := make([]interface{}, 0, len(rest))
	res = append(res, rest...)
	return res
}

// ---- Methods ----
// Note: methods that accept variadic args must call m.Called with those args expanded.

func (m *MockDatabase) Where(query interface{}, args ...interface{}) db.Database {
	callArgs := buildArgs(query, args)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	// ensure the stored value implements db.Database
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) First(dest interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(dest, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Create(value interface{}) error {
	ret := m.Called(value)
	return ret.Error(0)
}

func (m *MockDatabase) Model(value interface{}) db.Database {
	ret := m.Called(value)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Count(count *int64) db.Database {
	ret := m.Called(count)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Error() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDatabase) Save(value interface{}) error {
	ret := m.Called(value)
	return ret.Error(0)
}

func (m *MockDatabase) Preload(column string, conditions ...interface{}) db.Database {
	callArgs := buildArgsFromSlice(append([]interface{}{column}, conditions...))
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Delete(value interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(value, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Find(dest interface{}, conds ...interface{}) error {
	callArgs := buildArgs(dest, conds)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) AutoMigrate(dst ...interface{}) error {
	ret := m.Called(dst)
	return ret.Error(0)
}

func (m *MockDatabase) Offset(offset int) db.Database {
	ret := m.Called(offset)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Limit(limit int) db.Database {
	ret := m.Called(limit)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Scan(dest interface{}) error {
	ret := m.Called(dest)
	return ret.Error(0)
}

func (m *MockDatabase) Updates(values ...interface{}) error {
	ret := m.Called(values)
	return ret.Error(0)
}

func (m *MockDatabase) Update(column string, value interface{}) error {
	ret := m.Called(column, value)
	return ret.Error(0)
}

func (m *MockDatabase) Association(column string) *gorm.Association {
	ret := m.Called(column)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*gorm.Association)
}

func (m *MockDatabase) Select(query interface{}) db.Database {
	ret := m.Called(query)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Joins(query string, args ...interface{}) db.Database {
	callArgs := buildArgs(query, args)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Group(query string) db.Database {
	ret := m.Called(query)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Having(query interface{}, args ...interface{}) db.Database {
	callArgs := buildArgs(query, args)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Order(value interface{}) db.Database {
	ret := m.Called(value)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Scopes(funcs ...func(db.Database) db.Database) db.Database {
	ret := m.Called(funcs)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Take(dest interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(dest, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Last(dest interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(dest, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) UpdateColumns(values interface{}) error {
	ret := m.Called(values)
	return ret.Error(0)
}

func (m *MockDatabase) CreateInBatches(value interface{}, batchSize int) error {
	ret := m.Called(value, batchSize)
	return ret.Error(0)
}

func (m *MockDatabase) FindInBatches(dest interface{}, batchSize int, fc func(db.Database, int) error) error {
	ret := m.Called(dest, batchSize, fc)
	return ret.Error(0)
}

func (m *MockDatabase) UpdateInBatches(value interface{}, batchSize int) error {
	ret := m.Called(value, batchSize)
	return ret.Error(0)
}

func (m *MockDatabase) Transaction(fc func(db.Database) error, opts ...*sql.TxOptions) error {
	callArgs := buildArgsFromSlice(append([]interface{}{fc}, interfaceSliceFromTxOptions(opts)...))
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

// helper to convert []*sql.TxOptions to []interface{}
func interfaceSliceFromTxOptions(opts []*sql.TxOptions) []interface{} {
	res := make([]interface{}, 0, len(opts))
	for _, o := range opts {
		res = append(res, o)
	}
	return res
}

func (m *MockDatabase) Begin(opts ...*sql.TxOptions) db.Database {
	callArgs := buildArgsFromSlice(interfaceSliceFromTxOptions(opts))
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Commit() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDatabase) Rollback() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDatabase) Raw(sql string, values ...interface{}) db.Database {
	callArgs := buildArgs(sql, values)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Exec(sql string, values ...interface{}) error {
	callArgs := buildArgs(sql, values)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Pluck(column string, dest interface{}) error {
	ret := m.Called(column, dest)
	return ret.Error(0)
}

func (m *MockDatabase) Migrator() gorm.Migrator {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(gorm.Migrator)
}

func (m *MockDatabase) Distinct(args ...interface{}) db.Database {
	ret := m.Called(args)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Debug() db.Database {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Session(config *gorm.Session) db.Database {
	ret := m.Called(config)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) GetUnderlying() *gorm.DB {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDatabase) Unscoped() db.Database {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}
