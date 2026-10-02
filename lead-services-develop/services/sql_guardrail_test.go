package services

import (
	"strings"
	"testing"
)

func TestValidateAIGeneratedSQL_AcceptsSimpleSelect(t *testing.T) {
	sql := "SELECT COUNT(*) AS count FROM leads WHERE deleted_at IS NULL"
	got, err := ValidateAIGeneratedSQL(sql)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(got, "LIMIT") {
		t.Errorf("expected LIMIT to be injected, got: %s", got)
	}
}

func TestValidateAIGeneratedSQL_PreservesExistingLimit(t *testing.T) {
	sql := "SELECT id FROM leads WHERE deleted_at IS NULL LIMIT 10"
	got, err := ValidateAIGeneratedSQL(sql)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.Count(got, "LIMIT") != 1 {
		t.Errorf("expected exactly one LIMIT clause, got: %s", got)
	}
}

func TestValidateAIGeneratedSQL_AcceptsWithClause(t *testing.T) {
	sql := "WITH recent AS (SELECT id FROM leads WHERE deleted_at IS NULL) SELECT COUNT(*) FROM recent"
	if _, err := ValidateAIGeneratedSQL(sql); err != nil {
		t.Fatalf("expected WITH-clause query to be accepted, got %v", err)
	}
}

func TestValidateAIGeneratedSQL_RejectsNonSelect(t *testing.T) {
	cases := []string{
		"DELETE FROM leads WHERE id = '1'",
		"UPDATE leads SET status = 'x'",
		"INSERT INTO leads (id) VALUES ('x')",
		"DROP TABLE leads",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err == nil {
			t.Errorf("expected rejection for %q, got no error", sql)
		}
	}
}

func TestValidateAIGeneratedSQL_RejectsMultipleStatements(t *testing.T) {
	sql := "SELECT 1; DROP TABLE leads;"
	if _, err := ValidateAIGeneratedSQL(sql); err == nil {
		t.Error("expected multi-statement query to be rejected")
	}
}

func TestValidateAIGeneratedSQL_RejectsDeniedKeywordInsideSelect(t *testing.T) {
	// A SELECT statement that smuggles a denied function call should still
	// be rejected — the keyword scan applies to the whole query text, not
	// just the leading verb.
	sql := "SELECT pg_sleep(10) FROM leads"
	if _, err := ValidateAIGeneratedSQL(sql); err == nil {
		t.Error("expected query containing a denied keyword to be rejected")
	}
}

func TestValidateAIGeneratedSQL_RejectsDisallowedTable(t *testing.T) {
	sql := "SELECT * FROM pg_catalog.pg_tables"
	if _, err := ValidateAIGeneratedSQL(sql); err == nil {
		t.Error("expected query referencing a disallowed table to be rejected")
	}
}

func TestValidateAIGeneratedSQL_RejectsEmptyQuery(t *testing.T) {
	if _, err := ValidateAIGeneratedSQL("   "); err == nil {
		t.Error("expected empty query to be rejected")
	}
}

func TestValidateAIGeneratedSQL_AllowsExtractFromDateFunction(t *testing.T) {
	// Regression test: EXTRACT(MONTH FROM created_at) contains the keyword
	// FROM but is not a table reference — a real question ("how many leads
	// were created in June") produced this exact shape and was incorrectly
	// rejected before nonTableFromFuncPattern was added.
	sql := "SELECT COUNT(*) AS lead_count FROM leads WHERE EXTRACT(MONTH FROM created_at) = 6 AND deleted_at IS NULL"
	if _, err := ValidateAIGeneratedSQL(sql); err != nil {
		t.Fatalf("expected EXTRACT(... FROM ...) query to be accepted, got %v", err)
	}
}

func TestValidateAIGeneratedSQL_AllowsSubstringAndTrimFromUsage(t *testing.T) {
	cases := []string{
		"SELECT SUBSTRING(name FROM 1 FOR 3) FROM customers",
		"SELECT TRIM(BOTH ' ' FROM name) FROM customers",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err != nil {
			t.Errorf("expected %q to be accepted, got %v", sql, err)
		}
	}
}

func TestValidateAIGeneratedSQL_StillRejectsDisallowedTableNearDateFunction(t *testing.T) {
	// Make sure stripping EXTRACT(...) doesn't accidentally make the
	// allow-list check blind to a real disallowed table elsewhere in the
	// same query.
	sql := "SELECT COUNT(*) FROM secret_table WHERE EXTRACT(MONTH FROM created_at) = 6"
	if _, err := ValidateAIGeneratedSQL(sql); err == nil {
		t.Error("expected query referencing a disallowed table to still be rejected")
	}
}

func TestValidateAIGeneratedSQL_AllowsAllowlistedJoins(t *testing.T) {
	sql := "SELECT l.id FROM leads l JOIN estimates e ON e.lead_id = l.id WHERE l.deleted_at IS NULL"
	if _, err := ValidateAIGeneratedSQL(sql); err != nil {
		t.Fatalf("expected join across allow-listed tables to be accepted, got %v", err)
	}
}

func TestValidateAIGeneratedSQL_RejectsAgentsSensitiveColumns(t *testing.T) {
	cases := []string{
		"SELECT password FROM agents WHERE email = 'x@example.com'",
		"SELECT password_hashed FROM agents",
		"SELECT a.details FROM agents a",
		"SELECT id, password FROM agents",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err == nil {
			t.Errorf("expected %q to be rejected", sql)
		}
	}
}

func TestValidateAIGeneratedSQL_RejectsAgentsWildcardSelect(t *testing.T) {
	cases := []string{
		"SELECT * FROM agents",
		"SELECT a.* FROM agents a WHERE a.deleted_at IS NULL",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err == nil {
			t.Errorf("expected %q to be rejected", sql)
		}
	}
}

func TestValidateAIGeneratedSQL_AllowsAgentsSafeColumns(t *testing.T) {
	cases := []string{
		"SELECT id, name, email, role FROM agents WHERE deleted_at IS NULL",
		"SELECT l.id, a.name FROM leads l JOIN agents a ON a.id = l.sales1_person_id WHERE l.deleted_at IS NULL",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err != nil {
			t.Errorf("expected %q to be accepted, got %v", sql, err)
		}
	}
}

func TestValidateAIGeneratedSQL_WildcardSelectStillAllowedForNonAgentsTables(t *testing.T) {
	sql := "SELECT * FROM leads WHERE deleted_at IS NULL"
	if _, err := ValidateAIGeneratedSQL(sql); err != nil {
		t.Fatalf("expected wildcard select against a non-agents table to be accepted, got %v", err)
	}
}

func TestValidateAIGeneratedSQL_AllowsExpandedScopeTables(t *testing.T) {
	cases := []string{
		"SELECT COUNT(*) FROM stage_tasks WHERE completed = false AND deleted_at IS NULL",
		"SELECT st.name, su.name AS assigned_to FROM stage_tasks st " +
			"JOIN stage_users su ON su.stage_id = st.stage_id WHERE st.deleted_at IS NULL",
		"SELECT name, tax_percentage FROM county_taxes WHERE name = 'Cook'",
		"SELECT name FROM pickup_locations WHERE deleted_at IS NULL",
		"SELECT re.id, mre.residential_entry_name FROM residential_entries re " +
			"JOIN manager_residential_entries mre ON mre.id = re.manager_residential_entry_id " +
			"WHERE re.lead_id = 'L1'",
		"SELECT details->>'price' AS price FROM service_rows WHERE estimate_id = 'E1'",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err != nil {
			t.Errorf("expected %q to be accepted, got %v", sql, err)
		}
	}
}

func TestValidateAIGeneratedSQL_AllowsPhoneGroupedProfileQuery(t *testing.T) {
	// The "which customer/profile has the most leads" pattern — grouping by
	// details->>'Phone' is how this product actually defines a "profile"
	// (see ui/app/leads/profiles/page.tsx), not a join to customers.
	sql := "SELECT details->>'Phone' AS phone, COUNT(*) AS lead_count FROM leads " +
		"WHERE deleted_at IS NULL AND details->>'Phone' IS NOT NULL " +
		"GROUP BY details->>'Phone' ORDER BY lead_count DESC LIMIT 5"
	if _, err := ValidateAIGeneratedSQL(sql); err != nil {
		t.Fatalf("expected phone-grouped profile query to be accepted, got %v", err)
	}
}

func TestValidateAIGeneratedSQL_AllowsJobsTable(t *testing.T) {
	cases := []string{
		"SELECT COUNT(*) FROM jobs",
		"SELECT j.job_id, j.details->>'Total' AS total FROM jobs j " +
			"JOIN leads l ON l.id = j.lead_id WHERE l.deleted_at IS NULL",
		"SELECT COUNT(*) FROM leads WHERE status = 'COMPLETED' AND deleted_at IS NULL",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err != nil {
			t.Errorf("expected %q to be accepted, got %v", sql, err)
		}
	}
}

func TestValidateAIGeneratedSQL_AllowsInventoryTables(t *testing.T) {
	// Physical warehouse inventory (owned by the separate `inventory`
	// service, see the comments in ai_schema.go) — added so Bloom can
	// answer stock/supplier/purchase-order questions, which it couldn't
	// before since these tables weren't on the allow-list at all.
	cases := []string{
		"SELECT name, available FROM materials WHERE available <= threshold",
		"SELECT status, COUNT(*) FROM materialchunks WHERE material_id = 1 GROUP BY status",
		"SELECT name, category FROM suppliers WHERE details->'address'->>'city' = 'Dallas'",
		"SELECT p.id, s.name AS supplier_name FROM purchases p " +
			"JOIN suppliers s ON s.id = p.supplier_id WHERE p.status = 'Purchase'",
		"SELECT pr.invoice_id, pr.received_quantity FROM purchase_receiveds pr " +
			"JOIN purchases p ON p.id = pr.purchase_id WHERE p.material_id = 1",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err != nil {
			t.Errorf("expected %q to be accepted, got %v", sql, err)
		}
	}
}

func TestValidateAIGeneratedSQL_ExcludedTablesStillRejected(t *testing.T) {
	// profiles/addresses/files were deliberately not added to the allow-list
	// (see ai_schema.go); payments/permissions exist live (discovered via
	// schema.sql drift) and haven't been reviewed for Bloom at all — confirm
	// the guardrail still rejects all of them.
	cases := []string{
		"SELECT * FROM profiles",
		"SELECT * FROM addresses",
		"SELECT file_path FROM files",
		"SELECT * FROM payments",
		"SELECT * FROM permissions",
	}
	for _, sql := range cases {
		if _, err := ValidateAIGeneratedSQL(sql); err == nil {
			t.Errorf("expected %q to be rejected", sql)
		}
	}
}

func TestTruncateResults_NoOpUnderLimit(t *testing.T) {
	rows := []map[string]interface{}{{"id": 1}, {"id": 2}}
	got := TruncateResults(rows)
	if len(got) != 2 {
		t.Errorf("expected 2 rows unchanged, got %d", len(got))
	}
}

func TestTruncateResults_TruncatesOverLimit(t *testing.T) {
	rows := make([]map[string]interface{}, DefaultAIQueryRowLimit+50)
	for i := range rows {
		rows[i] = map[string]interface{}{"id": i}
	}
	got := TruncateResults(rows)
	if len(got) != DefaultAIQueryRowLimit {
		t.Errorf("expected truncation to %d rows, got %d", DefaultAIQueryRowLimit, len(got))
	}
}
