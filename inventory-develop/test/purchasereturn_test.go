package test

// This test file assumes:

//   - package path "inventory/handlers" exports ReturnRequest and
//     HandleCreatePurchaseReturn (adjust the import path below if different)
//   - middleware.GetDB(c) reads the Database out of gin.Context via
//     c.Get("db") / similar — tests set it with c.Set("db", fake).
//     If middleware uses a different key or a dedicated setter function,
//     swap the two c.Set("db", ...) calls in newTestContext for that.
//   - models.Purchase has ID uint, MaterialID uint, Status string,
//     RejectionReason string; models.PurchaseReceived has ID uint;
//     models.Materialchunks has ID uint, Status string, and models.Materials has an 'available' count.

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "inventory/db"
	"inventory/handlers"
	"inventory/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// Fake DB — implements the full dbpkg.Database interface (matching *GormDB's
// method set exactly), but only the methods this handler actually calls do
// anything interesting. Everything else is a harmless stub.
// ---------------------------------------------------------------------------

type fakeDB struct {
	purchase         models.Purchase
	purchaseFirstErr error

	purchaseReceived         models.PurchaseReceived
	purchaseReceivedFirstErr error

	chunks  []models.Materialchunks
	findErr error

	countResult int64
	countErr    error

	updateErr error
	saveErr   error // Error to return on Save
	createErr error

	// recorded calls, for assertions
	updateCalls   []updateCall
	savedPurchase *models.Purchase
	created       []interface{}
}
type gormExpr struct{} // Dummy type for assertion

type updateCall struct {
	column string
	value  interface{}
}

// --- methods the handler actually exercises ---

func (f *fakeDB) Transaction(fc func(tx dbpkg.Database) error, opts ...*sql.TxOptions) error {
	return fc(f)
}

func (f *fakeDB) First(dest interface{}, conditions ...interface{}) error {
	switch d := dest.(type) {
	case *models.Purchase:
		if f.purchaseFirstErr != nil {
			return f.purchaseFirstErr
		}
		*d = f.purchase
		return nil
	case *models.PurchaseReceived:
		if f.purchaseReceivedFirstErr != nil {
			return f.purchaseReceivedFirstErr
		}
		*d = f.purchaseReceived
		return nil
	}
	return nil
}

func (f *fakeDB) Where(query interface{}, args ...interface{}) dbpkg.Database {
	return f
}

func (f *fakeDB) Find(dest interface{}, conds ...interface{}) error {
	if f.findErr != nil {
		return f.findErr
	}
	if d, ok := dest.(*[]models.Materialchunks); ok {
		*d = f.chunks
	}
	return nil
}

func (f *fakeDB) Model(value interface{}) dbpkg.Database {
	return f
}

func (f *fakeDB) Count(count *int64) dbpkg.Database {
	*count = f.countResult
	return f
}

func (f *fakeDB) Error() error {
	return f.countErr
}

func (f *fakeDB) Update(column string, value interface{}) error {
	f.updateCalls = append(f.updateCalls, updateCall{column: column, value: value})
	// Special handling to detect gorm.Expr's return type
	if _, ok := value.(clause.Expr); ok {
		f.updateCalls[len(f.updateCalls)-1].value = gormExpr{}
	}

	return f.updateErr
}

func (f *fakeDB) Save(value interface{}) error {
	if p, ok := value.(*models.Purchase); ok {
		saved := *p
		f.savedPurchase = &saved
	}
	return f.saveErr
}

func (f *fakeDB) Create(value interface{}) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, value)
	return nil
}

// --- stubs to satisfy the rest of the Database interface ---

func (f *fakeDB) Preload(column string, conditions ...interface{}) dbpkg.Database { return f }
func (f *fakeDB) Delete(value interface{}, conditions ...interface{}) error       { return nil }
func (f *fakeDB) AutoMigrate(values ...interface{}) error                         { return nil }
func (f *fakeDB) Offset(offset int) dbpkg.Database                                { return f }
func (f *fakeDB) Limit(limit int) dbpkg.Database                                  { return f }
func (f *fakeDB) Scan(dest interface{}) error                                     { return nil }
func (f *fakeDB) Select(query interface{}) dbpkg.Database                         { return f }
func (f *fakeDB) Table(name string) dbpkg.Database                                { return f }
func (f *fakeDB) Joins(query string, args ...interface{}) dbpkg.Database          { return f }
func (f *fakeDB) Group(query string) dbpkg.Database                               { return f }
func (f *fakeDB) Updates(values ...interface{}) error                             { return nil }
func (f *fakeDB) Association(column string) *gorm.Association                     { return nil }
func (f *fakeDB) Having(query interface{}, args ...interface{}) dbpkg.Database    { return f }
func (f *fakeDB) Order(value interface{}) dbpkg.Database                          { return f }
func (f *fakeDB) Scopes(funcs ...func(dbpkg.Database) dbpkg.Database) dbpkg.Database {
	return f
}
func (f *fakeDB) Take(dest interface{}, conditions ...interface{}) error { return nil }
func (f *fakeDB) Last(dest interface{}, conditions ...interface{}) error { return nil }
func (f *fakeDB) UpdateColumns(values interface{}) error                 { return nil }
func (f *fakeDB) CreateInBatches(value interface{}, batchSize int) error { return nil }
func (f *fakeDB) FindInBatches(dest interface{}, batchSize int, fc func(tx dbpkg.Database, batch int) error) error {
	return nil
}
func (f *fakeDB) UpdateInBatches(value interface{}, batchSize int) error  { return nil }
func (f *fakeDB) Begin(opts ...*sql.TxOptions) dbpkg.Database             { return f }
func (f *fakeDB) IsNotFound() bool                                        { return false }
func (f *fakeDB) Commit() error                                           { return nil }
func (f *fakeDB) Rollback() error                                         { return nil }
func (f *fakeDB) Raw(sqlStr string, values ...interface{}) dbpkg.Database { return f }
func (f *fakeDB) Exec(sqlStr string, values ...interface{}) error         { return nil }
func (f *fakeDB) Pluck(column string, dest interface{}) error             { return nil }
func (f *fakeDB) Migrator() gorm.Migrator                                 { return nil }
func (f *fakeDB) Distinct(args ...interface{}) dbpkg.Database             { return f }
func (f *fakeDB) Debug() dbpkg.Database                                   { return f }
func (f *fakeDB) Session(config *gorm.Session) dbpkg.Database             { return f }
func (f *fakeDB) GetUnderlying() *gorm.DB                                 { return nil }
func (f *fakeDB) Unscoped() dbpkg.Database                                { return f }

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestContext(t *testing.T, db dbpkg.Database, purchaseID string, body handlers.ReturnRequest) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/purchases/"+purchaseID+"/return", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: purchaseID}}

	if db != nil {
		c.Set("db", db) // ASSUMPTION: matches how middleware.GetDB reads the context
	}

	return c, w
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// NOTE: A "DB missing from context" test was intentionally omitted here.
// middleware.GetDB(c) calls c.MustGet("db"), which panics if the key was
// never set — it does not return nil gracefully. That means the handler's
// `if db == nil` branch isn't reachable via a simply-unset context key, and
// simulating it would require knowing GetDB's real internals (e.g. whether
// it ever returns a typed-nil value that legitimately compares == nil).
// Share middleware/database.go if you'd like this branch covered precisely.

func TestHandleCreatePurchaseReturn_InvalidBody(t *testing.T) {
	db := &fakeDB{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/purchases/1/return", bytes.NewReader([]byte("not-json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", db)

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreatePurchaseReturn_PurchaseNotFound(t *testing.T) {
	db := &fakeDB{
		purchaseFirstErr: gorm.ErrRecordNotFound,
	}
	c, w := newTestContext(t, db, "999", handlers.ReturnRequest{Reason: "not found case"})

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCreatePurchaseReturn_HappyPath_ChunksViaPurchaseReceived(t *testing.T) {
	db := &fakeDB{
		purchase:         models.Purchase{ID: 1, MaterialID: 42},
		purchaseReceived: models.PurchaseReceived{ID: 7},
		chunks: []models.Materialchunks{
			{ID: 101, Status: "available"},
			{ID: 102, Status: "available"},
			{ID: 103, Status: "reserved"},
		},
		countResult: 2, // 2 of the 3 chunks are "available"
	}
	c, w := newTestContext(t, db, "1", handlers.ReturnRequest{Reason: "wrong item shipped"})

	handlers.HandleCreatePurchaseReturn()(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, db.savedPurchase)
	assert.Equal(t, "Returned", db.savedPurchase.Status)
	assert.Equal(t, "wrong item shipped", db.savedPurchase.ReturnReason)
	require.Len(t, db.created, 1)

	// two Update calls expected: chunk status -> "returned", material available -> decremented
	require.Len(t, db.updateCalls, 2)
	assert.Equal(t, "status", db.updateCalls[0].column)
	assert.Equal(t, "returned", db.updateCalls[0].value)
	assert.Equal(t, "available", db.updateCalls[1].column, "Expected second update to be for 'available' column")
	assert.IsType(t, gormExpr{}, db.updateCalls[1].value, "Expected available count to be updated with a gorm.Expr")
}

func TestHandleCreatePurchaseReturn_FallsBackToPurchaseIDWhenNoReceivedRecord(t *testing.T) {
	db := &fakeDB{
		purchase:         models.Purchase{ID: 5, MaterialID: 42},
		purchaseReceived: models.PurchaseReceived{ID: 0}, // not found
		chunks: []models.Materialchunks{
			{ID: 201, Status: "available"},
		},
		countResult: 1,
	}
	c, w := newTestContext(t, db, "5", handlers.ReturnRequest{Reason: "damaged in transit"})

	handlers.HandleCreatePurchaseReturn()(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, db.savedPurchase)
	assert.Equal(t, "Returned", db.savedPurchase.Status)
}

func TestHandleCreatePurchaseReturn_NoChunksFound_StillUpdatesPurchase(t *testing.T) {
	db := &fakeDB{
		purchase: models.Purchase{ID: 9, MaterialID: 42},
	}
	c, w := newTestContext(t, db, "9", handlers.ReturnRequest{Reason: "no chunks to return"})

	handlers.HandleCreatePurchaseReturn()(c)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, db.updateCalls, "no chunk/material updates should happen when there are no chunks")
	require.NotNil(t, db.savedPurchase)
	assert.Equal(t, "Returned", db.savedPurchase.Status)
}

func TestHandleCreatePurchaseReturn_CountError_RollsBackWith500(t *testing.T) {
	db := &fakeDB{
		purchase: models.Purchase{ID: 1, MaterialID: 42},
		chunks:   []models.Materialchunks{{ID: 1, Status: "available"}},
		countErr: assertNewErr("count failed"),
	}
	c, w := newTestContext(t, db, "1", handlers.ReturnRequest{Reason: "x"})

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCreatePurchaseReturn_ChunkUpdateError_RollsBackWith500(t *testing.T) {
	db := &fakeDB{
		purchase:  models.Purchase{ID: 1, MaterialID: 42},
		chunks:    []models.Materialchunks{{ID: 1, Status: "available"}},
		updateErr: assertNewErr("update failed"),
	}
	c, w := newTestContext(t, db, "1", handlers.ReturnRequest{Reason: "x"})

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCreatePurchaseReturn_SaveError_RollsBackWith500(t *testing.T) {
	db := &fakeDB{
		purchase: models.Purchase{ID: 1, MaterialID: 42},
		saveErr:  assertNewErr("save failed"),
	}
	c, w := newTestContext(t, db, "1", handlers.ReturnRequest{Reason: "x"})

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCreatePurchaseReturn_ActivityCreateError_RollsBackWith500(t *testing.T) {
	db := &fakeDB{
		purchase:  models.Purchase{ID: 1, MaterialID: 42},
		createErr: assertNewErr("activity log failed"),
	}
	c, w := newTestContext(t, db, "1", handlers.ReturnRequest{Reason: "x"})

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCreatePurchaseReturn_MaterialUpdateError_RollsBackWith500(t *testing.T) {
	db := &fakeDB{
		purchase:    models.Purchase{ID: 1, MaterialID: 42},
		chunks:      []models.Materialchunks{{ID: 101, Status: "available"}},
		countResult: 1,
	}
	// First update (chunk status) succeeds, second (material available) fails.
	db.updateErr = assertNewErr("material update failed")

	c, w := newTestContext(t, db, "1", handlers.ReturnRequest{Reason: "test material update failure"})

	handlers.HandleCreatePurchaseReturn()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// Only the first update (chunk status) should have been attempted before the error.
	assert.Len(t, db.updateCalls, 1)
}

// small local helper so this file doesn't need "errors" just for test fixtures
func assertNewErr(msg string) error {
	return &simpleErr{msg: msg}
}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }
