package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/models"
)

type JobService struct{}

func NewJobService() *JobService {
	return &JobService{}
}

func (s *JobService) UpsertJobFromLead(database db.Database, lead models.Lead) error {

	type StageWithTasks struct {
		models.Stage
		StageTaskNames  []string `json:"stageTaskNames"`
		StageTaskStates []string `json:"stageTaskStates"`
	}

	// Look up existing job record first — needed for fallback paths and upsert.
	var existingJob models.Job
	jobFound := database.Where("lead_id = ?", lead.ID).First(&existingJob) == nil

	// Fetch every stage of the lead, then narrow to this job's ID below. Reading
	// the lead first keeps a newly created job from capturing only the stages
	// that already carried its job ID (which can be a lone auto-provisioned
	// Template stage) and silently dropping the rest of the job.
	var stages []models.Stage

	// 1. Primary: fetch by lead_id.
	_ = database.Where("lead_id = ?", lead.ID).Find(&stages)

	// 2. Fallback: fetch by job_id (set at stage-creation time, always reliable).
	if len(stages) == 0 && jobFound && existingJob.JobID != "" {
		_ = database.Where("job_id = ?", existingJob.JobID).Find(&stages)
	}

	// 3. Fallback2: extract IDs from existing blob and fetch rows directly.
	if len(stages) == 0 && jobFound && len(existingJob.Stages) > 0 {
		type minGroup struct {
			Stages []struct {
				ID string `json:"ID"`
			} `json:"stages"`
		}
		var groups []minGroup
		if json.Unmarshal(existingJob.Stages, &groups) == nil {
			var blobIDs []string
			for _, g := range groups {
				for _, st := range g.Stages {
					if st.ID != "" {
						blobIDs = append(blobIDs, st.ID)
					}
				}
			}
			if len(blobIDs) > 0 {
				_ = database.Where("id IN ?", blobIDs).Find(&stages)
			}
		}
	}

	if len(stages) == 0 {
		return nil
	}

	// ── Fetch tasks ─────────────────────────────────────────────────────────────

	var stageIDs []string
	for _, stage := range stages {
		stageIDs = append(stageIDs, stage.ID)
	}

	var stageTasks []models.StageTask
	if err := database.Where("stage_id IN ?", stageIDs).Find(&stageTasks); err != nil {
		return err
	}

	stageTaskMap := make(map[string][]string)
	stageTaskStateMap := make(map[string][]string)
	for _, task := range stageTasks {
		stageTaskMap[task.StageID] = append(stageTaskMap[task.StageID], task.Name)
		state := "In Progress"
		if task.Completed {
			state = "Completed"
		}else if task.Cancelled {
			state = "Cancelled"
		}
		stageTaskStateMap[task.StageID] = append(stageTaskStateMap[task.StageID], state)
	}

	isCompleted := len(stages) > 0
	for _, stage := range stages {
		if stage.State != "Completed" {
			isCompleted = false
			break
		}
	}

	// ── Resolve jobID ───────────────────────────────────────────────────────────

	jobID := existingJob.JobID
	if jobID == "" {
		jobID = stages[0].JobID
	}

	// A lead has one job, so every stage of every residential entry must carry
	// that job ID. Estimate revisions used to mint their own ID, which split one
	// job's entries across several IDs; this collapses them back onto the job so
	// projects added later stay part of it.
	if jobID != "" {
		_ = database.Model(&models.Stage{}).
			Where("lead_id = ? AND job_id <> ?", lead.ID, jobID).
			Update("job_id", jobID)
		for i := range stages {
			stages[i].JobID = jobID
		}
	}

	// ── Group stages by ResidentialEntryID ─────────────────────────────────────

	grouped := make(map[uint][]models.Stage)
	for _, stage := range stages {
		grouped[stage.ResidentialEntryID] = append(grouped[stage.ResidentialEntryID], stage)
	}

	type ResidentialStageGroup struct {
		ResidentialEntryID      uint             `json:"residentialEntryId"`
		ResidentialEntryName    string           `json:"residentialEntryName"`
		ResidentialEntryDetails string           `json:"residentialEntryDetails"`
		Stages                  []StageWithTasks `json:"stages"`
	}

	var final []ResidentialStageGroup

	for reID, stgs := range grouped {
		var residential models.ResidentialEntry
		if err := database.Preload("ManagerResidentialEntry").First(&residential, reID); err != nil {
			continue
		}

		var enrichedStages []StageWithTasks
		for _, stage := range stgs {
			taskNames := stageTaskMap[stage.ID]
			if taskNames == nil {
				taskNames = []string{}
			}
			taskStates := stageTaskStateMap[stage.ID]
			if taskStates == nil {
				taskStates = []string{}
			}
			enrichedStages = append(enrichedStages, StageWithTasks{
				Stage:           stage,
				StageTaskNames:  taskNames,
				StageTaskStates: taskStates,
			})
		}

		final = append(final, ResidentialStageGroup{
			ResidentialEntryID:      reID,
			ResidentialEntryName:    residential.ManagerResidentialEntry.ResidentialEntryName,
			ResidentialEntryDetails: residential.Details,
			Stages:                  enrichedStages,
		})
	}

	stagesJSON, _ := json.Marshal(final)

	// ── Estimate and total ──────────────────────────────────────────────────────

	var estimate models.Estimate
	var estimateID *string
	if err := database.Where("lead_id = ?", lead.ID).First(&estimate); err == nil {
		estimateID = &estimate.ID
	}

	var jobTotal float64
	var latestEst models.Estimate
	if err := database.
		Where("lead_id = ? AND id LIKE ?", lead.ID, "%-v%").
		Order("updated_at desc").
		First(&latestEst); err == nil {
		jobTotal = latestEst.Details.Total
	} else if estimateID != nil {
		var baseEst models.Estimate
		if err := database.Where("id = ?", *estimateID).First(&baseEst); err == nil {
			jobTotal = baseEst.Details.Total
		}
	}

	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"jobName": lead.JobName,
		"Total":   jobTotal,
	})

	// ── Upsert ──────────────────────────────────────────────────────────────────

	if jobFound {
		existingJob.Stages = stagesJSON
		existingJob.EstimateID = estimateID
		existingJob.Details = detailsJSON
		existingJob.Completed = isCompleted
		return database.Save(&existingJob)
	}

	job := models.Job{
		ID:         jobID,
		JobID:      jobID,
		LeadID:     lead.ID,
		EstimateID: estimateID,
		Stages:     stagesJSON,
		Details:    detailsJSON,
		Completed:  isCompleted,
	}
	return database.Create(&job)
}

// SyncJobByID rebuilds the job blob for the given job_id and returns the
// refreshed job data in the same shape as GetAllJobs.

func (s *JobService) SyncJobByID(database db.Database, jobID string) (map[string]interface{}, error) {
	// 1. Find the job record.
	var job models.Job
	if err := database.Where("job_id = ?", jobID).First(&job); err != nil {
		// Try by id column as fallback.
		if err2 := database.Where("id = ?", jobID).First(&job); err2 != nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
	}

	// 2. Find the lead.
	var lead models.Lead
	if err := database.Where("id = ?", job.LeadID).First(&lead); err != nil {
		return nil, fmt.Errorf("lead not found for job %s: %v", jobID, err)
	}

	// 3. Rebuild and persist the blob.
	if err := s.UpsertJobFromLead(database, lead); err != nil {
		return nil, fmt.Errorf("UpsertJobFromLead failed: %v", err)
	}

	// 4. Re-fetch the freshly updated job record.
	var updated models.Job
	if err := database.Where("job_id = ?", jobID).First(&updated); err != nil {
		if err2 := database.Where("id = ?", jobID).First(&updated); err2 != nil {
			return nil, fmt.Errorf("failed to re-fetch job after sync: %v", err2)
		}
	}

	// 5. Parse the blob and enrich with live stage states (same as GetAllJobs).
	type StageWithTasks struct {
		ID                  string      `json:"ID"`
		Name                string      `json:"Name"`
		State               string      `json:"State"`
		Description         string      `json:"Description"`
		Date                string      `json:"Date"`
		Notes               interface{} `json:"Notes"`
		Team                string      `json:"Team"`
		ResidentialEntryID  uint        `json:"ResidentialEntryID"`
		LeadID              string      `json:"LeadID"`
		JobID               string      `json:"JobID"`
		StageTaskNames      []string    `json:"stageTaskNames"`
		StageTaskStates     []string    `json:"stageTaskStates"`
		StageTaskCreatedAts []time.Time `json:"stageTaskCreatedAts"`
		StageTaskUpdatedAts []time.Time `json:"stageTaskUpdatedAts"`
	}
	type ResidentialStageGroup struct {
		ResidentialEntryID      uint             `json:"residentialEntryId"`
		ResidentialEntryName    string           `json:"residentialEntryName"`
		ResidentialEntryDetails string           `json:"residentialEntryDetails"`
		Stages                  []StageWithTasks `json:"stages"`
	}

	var stageGroups []ResidentialStageGroup
	if len(updated.Stages) > 0 {
		_ = json.Unmarshal(updated.Stages, &stageGroups)
	}

	// Collect stage IDs for live-state override and task enrichment.
	var stageIDs []string
	for _, g := range stageGroups {
		for _, st := range g.Stages {
			stageIDs = append(stageIDs, st.ID)
		}
	}

	// Live state map — read directly from stages table (not blob).
	liveStateMap := make(map[string]string)
	liveDateMap := make(map[string]string)
	if len(stageIDs) > 0 {
		var liveStages []models.Stage
		if err := database.Where("id IN ?", stageIDs).Find(&liveStages); err == nil {
			for _, ls := range liveStages {
				liveStateMap[ls.ID] = ls.State
				if ls.Date != nil {
					liveDateMap[ls.ID] = ls.Date.Format(time.RFC3339)
				}
			}
		}
	}

	// Task enrichment maps.
	stageTaskMap := make(map[string][]string)
	stageTaskStateMap := make(map[string][]string)
	stageTaskCreatedAtMap := make(map[string][]time.Time)
	stageTaskUpdatedAtMap := make(map[string][]time.Time)
	if len(stageIDs) > 0 {
		var stageTasks []models.StageTask
		if err := database.Where("stage_id IN ?", stageIDs).Find(&stageTasks); err == nil {
			for _, task := range stageTasks {
				stageTaskMap[task.StageID] = append(stageTaskMap[task.StageID], task.Name)
				stageTaskCreatedAtMap[task.StageID] = append(stageTaskCreatedAtMap[task.StageID], task.CreatedAt)
				stageTaskUpdatedAtMap[task.StageID] = append(stageTaskUpdatedAtMap[task.StageID], task.UpdatedAt)
				state := "In Progress"
				if task.Completed {
					state = "Completed"
				}
				stageTaskStateMap[task.StageID] = append(stageTaskStateMap[task.StageID], state)
			}
		}
	}

	// Apply live state + task data.
	for gi := range stageGroups {
		for si := range stageGroups[gi].Stages {
			id := stageGroups[gi].Stages[si].ID
			if s, ok := liveStateMap[id]; ok {
				stageGroups[gi].Stages[si].State = s
			}
			if d, ok := liveDateMap[id]; ok {
				stageGroups[gi].Stages[si].Date = d
			}
			names := stageTaskMap[id]
			if names == nil {
				names = []string{}
			}
			states := stageTaskStateMap[id]
			if states == nil {
				states = []string{}
			}
			createdAts := stageTaskCreatedAtMap[id]
			if createdAts == nil {
				createdAts = []time.Time{}
			}
			updatedAts := stageTaskUpdatedAtMap[id]
			if updatedAts == nil {
				updatedAts = []time.Time{}
			}
			stageGroups[gi].Stages[si].StageTaskNames = names
			stageGroups[gi].Stages[si].StageTaskStates = states
			stageGroups[gi].Stages[si].StageTaskCreatedAts = createdAts
			stageGroups[gi].Stages[si].StageTaskUpdatedAts = updatedAts
		}
	}

	// Resolve estimate total (same logic as GetAllJobs).
	var detailsMap map[string]interface{}
	if err := json.Unmarshal(updated.Details, &detailsMap); err != nil {
		detailsMap = make(map[string]interface{})
	}
	var allEstimates []models.Estimate
	if err := database.
		Where("lead_id = ?", updated.LeadID).
		Order("updated_at desc").
		Find(&allEstimates); err == nil {
		total := float64(0)
		set := false
		for _, est := range allEstimates {
			if !set && strings.Contains(est.ID, "-v") {
				total = est.Details.Total
				set = true
			}
		}
		if !set && len(allEstimates) > 0 {
			total = allEstimates[0].Details.Total
		}
		detailsMap["Total"] = total
	}
	enrichedDetails, _ := json.Marshal(detailsMap)

	return map[string]interface{}{
		"id":          updated.ID,
		"job_id":      updated.JobID,
		"lead_id":     updated.LeadID,
		"estimate_id": updated.EstimateID,
		"details":     json.RawMessage(enrichedDetails),
		"stages":      stageGroups,
		"created_at":  updated.CreatedAt,
		"updated_at":  updated.UpdatedAt,
	}, nil
}

func (s *JobService) GetAllJobs(database db.Database, status string, page, limit int, search string) ([]map[string]interface{}, int, error) {
	if database == nil {
		return nil, 0, fmt.Errorf("database instance is nil")
	}

	query := database.Order("created_at desc")

	if status == "completed" {
		query = query.Where("completed = ?", true)
	} else if status == "pending" {
		query = query.Where("completed = ?", false)
	}

	if search != "" {
		searchLike := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(details->>'jobName') LIKE ? OR LOWER(job_id) LIKE ?",
			searchLike, searchLike,
		)
	}

	var allJobs []models.Job
	if err := query.Find(&allJobs); err != nil {
		return nil, 0, err
	}

	type StageWithTasks struct {
		ID                  string      `json:"ID"`
		Name                string      `json:"Name"`
		State               string      `json:"State"`
		Description         string      `json:"Description"`
		Date                string      `json:"Date"`
		Notes               interface{} `json:"Notes"`
		Team                string      `json:"Team"`
		ResidentialEntryID  uint        `json:"ResidentialEntryID"`
		LeadID              string      `json:"LeadID"`
		JobID               string      `json:"JobID"`
		StageTaskNames      []string    `json:"stageTaskNames"`
		StageTaskStates     []string    `json:"stageTaskStates"`
		StageTaskCreatedAts []time.Time `json:"stageTaskCreatedAts"`
		StageTaskUpdatedAts []time.Time `json:"stageTaskUpdatedAts"`
	}

	type ResidentialStageGroup struct {
		ResidentialEntryID      uint             `json:"residentialEntryId"`
		ResidentialEntryName    string           `json:"residentialEntryName"`
		ResidentialEntryDetails string           `json:"residentialEntryDetails"`
		Stages                  []StageWithTasks `json:"stages"`
	}

	type enrichedJob struct {
		job             models.Job
		stageGroups     []ResidentialStageGroup
		latestUpdatedAt time.Time
	}

	// ── Pass 1: collect all stage IDs across all blobs ──────────────────────────

	allBlobStageIDs := make([]string, 0)
	for _, job := range allJobs {
		if len(job.Stages) == 0 {
			continue
		}
		type minGroup struct {
			Stages []struct {
				ID string `json:"ID"`
			} `json:"stages"`
		}
		var groups []minGroup
		if json.Unmarshal(job.Stages, &groups) != nil {
			continue
		}
		for _, g := range groups {
			for _, st := range g.Stages {
				if st.ID != "" {
					allBlobStageIDs = append(allBlobStageIDs, st.ID)
				}
			}
		}
	}

	// ── Pass 2: batch-fetch live State + Date from stages table ─────────────────

	liveStateMap := make(map[string]string)
	liveDateMap := make(map[string]string)
	if len(allBlobStageIDs) > 0 {
		// Use models.Stage so GORM resolves the correct table name.
		// A local struct would cause GORM to infer a non-existent table name
		// and silently return empty, making the override a no-op.
		var liveStages []models.Stage
		if err := database.Where("id IN ?", allBlobStageIDs).Find(&liveStages); err == nil {
			for _, ls := range liveStages {
				liveStateMap[ls.ID] = ls.State
				if ls.Date != nil {
					liveDateMap[ls.ID] = ls.Date.Format(time.RFC3339)
				}
			}
		}
		// Failure here degrades gracefully to blob values.
	}

	// ── Pass 3: build response with live-state override ──────────────────────────

	var matchingJobs []enrichedJob

	for _, job := range allJobs {
		var stageGroups []ResidentialStageGroup
		if len(job.Stages) > 0 {
			if err := json.Unmarshal(job.Stages, &stageGroups); err != nil {
				stageGroups = []ResidentialStageGroup{}
			}
		}
		if len(stageGroups) == 0 {
			continue
		}

		var stageIDs []string
		for _, group := range stageGroups {
			for _, stage := range group.Stages {
				stageIDs = append(stageIDs, stage.ID)
			}
		}

		stageTaskMap := make(map[string][]string)
		stageTaskStateMap := make(map[string][]string)
		StageTaskCreatedAtMap := make(map[string][]time.Time)
		StageTaskUpdatedAtMap := make(map[string][]time.Time)
		if len(stageIDs) > 0 {
			var stageTasks []models.StageTask
			if err := database.Where("stage_id IN ?", stageIDs).Find(&stageTasks); err != nil {
				return nil, 0, err
			}
			for _, task := range stageTasks {
				stageTaskMap[task.StageID] = append(stageTaskMap[task.StageID], task.Name)
				StageTaskCreatedAtMap[task.StageID] = append(StageTaskCreatedAtMap[task.StageID], task.CreatedAt)
				StageTaskUpdatedAtMap[task.StageID] = append(StageTaskUpdatedAtMap[task.StageID], task.UpdatedAt)
				state := "In Progress"
				if task.Completed {
					state = "Completed"
				}else if task.Cancelled {
					state = "Cancelled"
				}
				stageTaskStateMap[task.StageID] = append(stageTaskStateMap[task.StageID], state)
			}
		}

		for gi := range stageGroups {
			for si := range stageGroups[gi].Stages {
				stageID := stageGroups[gi].Stages[si].ID

				// Override blob State/Date with authoritative DB value.
				if liveState, ok := liveStateMap[stageID]; ok {
					stageGroups[gi].Stages[si].State = liveState
				}
				if liveDate, ok := liveDateMap[stageID]; ok {
					stageGroups[gi].Stages[si].Date = liveDate
				}

				names := stageTaskMap[stageID]
				if names == nil {
					names = []string{}
				}
				states := stageTaskStateMap[stageID]
				if states == nil {
					states = []string{}
				}
				createdAts := StageTaskCreatedAtMap[stageID]
				if createdAts == nil {
					createdAts = []time.Time{}
				}
				updatedAts := StageTaskUpdatedAtMap[stageID]
				if updatedAts == nil {
					updatedAts = []time.Time{}
				}
				stageGroups[gi].Stages[si].StageTaskNames = names
				stageGroups[gi].Stages[si].StageTaskStates = states
				stageGroups[gi].Stages[si].StageTaskCreatedAts = createdAts
				stageGroups[gi].Stages[si].StageTaskUpdatedAts = updatedAts
			}
		}

		
			latestUpdatedAt := job.UpdatedAt

		matchingJobs = append(matchingJobs, enrichedJob{
			job:             job,
			stageGroups:     stageGroups,
			latestUpdatedAt: latestUpdatedAt,
		})
	}

	totalCount := len(matchingJobs)

	var pageJobs []enrichedJob
	if limit == 0 {
		pageJobs = matchingJobs
	} else {
		offset := (page - 1) * limit
		if offset >= totalCount {
			return []map[string]interface{}{}, totalCount, nil
		}
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		pageJobs = matchingJobs[offset:end]
	}

	leadIDs := make([]string, 0, len(pageJobs))
	for _, ej := range pageJobs {
		leadIDs = append(leadIDs, ej.job.LeadID)
	}

	estimateTotalByLead := make(map[string]float64)
	if len(leadIDs) > 0 {
		var allEstimates []models.Estimate
		if err := database.
			Where("lead_id IN ?", leadIDs).
			Order("updated_at desc").
			Find(&allEstimates); err == nil {

			baseEstimateByLead := make(map[string]models.Estimate)
			for _, est := range allEstimates {
				if est.LeadID == nil {
					continue
				}
				leadID := *est.LeadID
				if _, exists := baseEstimateByLead[leadID]; !exists {
					baseEstimateByLead[leadID] = est
				}
			}
			for _, est := range allEstimates {
				if est.LeadID == nil {
					continue
				}
				leadID := *est.LeadID
				if _, alreadySet := estimateTotalByLead[leadID]; !alreadySet {
					if strings.Contains(est.ID, "-v") {
						estimateTotalByLead[leadID] = est.Details.Total
					}
				}
			}
			for leadID, baseEst := range baseEstimateByLead {
				if _, exists := estimateTotalByLead[leadID]; !exists {
					estimateTotalByLead[leadID] = baseEst.Details.Total
				}
			}
		}
	}

	response := make([]map[string]interface{}, 0, len(pageJobs))
	for _, ej := range pageJobs {
		var detailsMap map[string]interface{}
		if err := json.Unmarshal(ej.job.Details, &detailsMap); err != nil {
			detailsMap = make(map[string]interface{})
		}
		detailsMap["Total"] = estimateTotalByLead[ej.job.LeadID]
		enrichedDetails, _ := json.Marshal(detailsMap)

		response = append(response, map[string]interface{}{
			"id":          ej.job.ID,
			"job_id":      ej.job.JobID,
			"lead_id":     ej.job.LeadID,
			"estimate_id": ej.job.EstimateID,
			"details":     json.RawMessage(enrichedDetails),
			"stages":      ej.stageGroups,
			"created_at":  ej.job.CreatedAt,
			"updated_at":  ej.latestUpdatedAt,
		})
	}

	return response, totalCount, nil
}