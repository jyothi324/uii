package handlers

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
)

// Regression coverage for a real bug: the installer field on a job used to
// be picked from the agent list (always came with a real
// InstallationSalesPersonID), then became a plain free-text box that never
// sends an ID. This function used to unconditionally blank the name
// whenever the ID was empty — which is now the normal case — silently
// discarding every typed installer name on save. It would still show for
// the rest of that session (only local state), then vanish on the next
// load once the API returned this now-blank field back.
func TestResolveInstallationSalesPersonName_NoID_PreservesFreeTextName(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	details := &models.EstimateDetails{
		InstallationSalesPersonID:   "",
		InstallationSalesPersonName: "Puru",
	}

	resolveInstallationSalesPersonName(mockDB, details)

	if details.InstallationSalesPersonName != "Puru" {
		t.Errorf("expected free-text installer name to survive with no ID, got %q", details.InstallationSalesPersonName)
	}
	mockDB.AssertNotCalled(t, "First", mock.Anything, mock.Anything, mock.Anything)
}

func TestResolveInstallationSalesPersonName_ValidID_ResolvesAgentName(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockDB.On("First", mock.AnythingOfType("*models.Agent"), "id = ?", "AGT-1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Agent)
		dest.Name = "Real Agent Name"
	}).Return(nil).Once()

	details := &models.EstimateDetails{
		InstallationSalesPersonID:   "AGT-1",
		InstallationSalesPersonName: "stale name from an earlier save",
	}

	resolveInstallationSalesPersonName(mockDB, details)

	if details.InstallationSalesPersonName != "Real Agent Name" {
		t.Errorf("expected name resolved from the agent record, got %q", details.InstallationSalesPersonName)
	}
	mockDB.AssertExpectations(t)
}

func TestResolveInstallationSalesPersonName_InvalidID_LeavesNameUntouched(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockDB.On("First", mock.AnythingOfType("*models.Agent"), "id = ?", "GONE").
		Return(errRecordNotFoundForTest{}).Once()

	details := &models.EstimateDetails{
		InstallationSalesPersonID:   "GONE",
		InstallationSalesPersonName: "whatever was sent",
	}

	resolveInstallationSalesPersonName(mockDB, details)

	if details.InstallationSalesPersonName != "whatever was sent" {
		t.Errorf("a stale/invalid ID should be non-fatal and leave the existing name as-is, got %q", details.InstallationSalesPersonName)
	}
}

type errRecordNotFoundForTest struct{}

func (errRecordNotFoundForTest) Error() string { return "record not found" }
