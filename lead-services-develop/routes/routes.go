package routes

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/services"
)

type Handler struct {
	DB db.Database
}

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	h := handlers.NewHandler()
	api.GET("/company/check", h.CheckCompanyExists)

	// NEW SaaS billing routes (additive — does not alter existing handlers)
	RegisterBillingRoutes(api)

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-2"
	}
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "tulipfiles"
	}
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if awsAccessKey == "" || awsSecretKey == "" {
		log.Fatal("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			awsAccessKey,
			awsSecretKey,
			"",
		),
	})
	if err != nil {
		log.Fatal("Failed to create AWS session:", err)
	}

	s3Client := s3.New(sess)
	leadHandler := handlers.NewLeadHandler(s3Client, s3Bucket, awsRegion)
	fileHandler := handlers.NewFileHandler(s3Client, s3Bucket, awsRegion)
	jobHandler := handlers.NewJobHandler()
	aiHandler := handlers.NewAIHandler(services.NewHTTPGeminiClient())
	{
		api.POST("/login", h.Login)
		api.GET("/check-mfa", h.CheckMFA)
		api.POST("/send-otp", h.SendOTP)
		api.POST("/verify-otp", h.VerifyOTP)
		api.POST("/enable-totp", h.EnableTOTP)
		api.GET("/checkemailexists", h.CheckEmailExists)
		api.GET("/checkphoneexists", h.CheckPhoneExists)
		api.POST("/agent", h.CreateAgent)
		api.PUT("/reset-password", h.UpdatePassword)
		// CustomerProfile routes
		api.POST("/profiles", h.CreateProfile)
		api.GET("/profiles", h.GetAllProfiles)
		api.POST("/profiles/:phoneoremail", h.GetProfile)
		api.PUT("/profiles/:id", h.UpdateProfile)
		api.DELETE("/profiles/:id", h.DeleteProfile)
		// UserGroup routes
		api.POST("/usergroups", h.CreateUserGroup)
		api.PUT("/usergroups/:id", h.UpdateUserGroup)
		api.DELETE("/usergroups/:id", h.DeleteUserGroup)
		api.GET("/usergroups/all", h.GetAllUserGroups)
		api.DELETE("/usergroups/:id/users/:user_id", h.RemoveUserFromGroup)
		// Permission routes
		api.PUT("/permissions/:id", h.UpdatePermission)
		api.GET("/permissions/:id", h.CheckPermission)
		api.GET("/permissions/all", h.GetAllPermissions)
		api.PUT("/permissions/bulk", h.BulkUpdatePermissions)
		// api.PATCH("/agent/:id/show-prices", h.UpdateShowPrices)

		api.Use(middleware.AuthMiddleware)
		api.GET("/agents", h.ListAgents)
		api.GET("/auth/verify", h.VerifyAuth)
		api.POST("/logout", h.Logout)
		api.POST("/ai/ask", aiHandler.Ask)
		auth := api.Group("/auth")
		auth.Use(middleware.RequireRole("MANAGER"))
		{
			auth.PUT("/superuser/:id", h.UpdateSuperuserStatus)
			auth.DELETE("/agent/:id", h.DeleteAgent)
			auth.PUT("/agent/:id", h.UpdateAgent)
			// CountryTax routes
			auth.POST("/countytaxes", h.CreateCountyTax)
			auth.PUT("/countytaxes/:id", h.UpdateCountyTax)
			auth.DELETE("/countytaxes/:id", h.DeleteCountyTax)
			auth.GET("/countytaxes/all", h.GetAllCountyTaxes)
			// // BranchLocation routes
			auth.POST("/pickuplocations", h.CreatePickup)
			auth.PUT("/pickuplocations/:id", h.UpdatePickup)
			auth.DELETE("/pickuplocations/:id", h.DeletePickup)
			auth.GET("/pickuplocations/all", h.ListPickupLocations)
			// // Service routes
			auth.POST("/services", h.CreateService)
			auth.PUT("/services/:id", h.UpdateService)
			auth.DELETE("/services/:id", h.DeleteService)
			auth.GET("/services/all", h.GetAllServices)
			// // ServiceGroup routes
			auth.POST("/servicegroups", h.CreateServiceGroup)
			auth.PUT("/servicegroups/:id", h.UpdateServiceGroup)
			auth.DELETE("/servicegroups/:id", h.DeleteServiceGroup)
			auth.GET("/servicegroups/all", h.GetAllServiceGroups)

			// ServiceCategory routes
			auth.POST("/servicecategories", h.CreateServiceCategory)
			auth.PUT("/servicecategories/:id", h.UpdateServiceCategory)
			auth.DELETE("/servicecategories/:id", h.DeleteServiceCategory)
			auth.GET("/servicecategories/all", h.GetAllServiceCategories)

			// Customer routes
			auth.PUT("/customers/:id", h.UpdateCustomer)
			auth.DELETE("/customers/:id", h.DeleteCustomer)
            // Lead Control routes
			auth.GET("/lead-control/all", h.GetLeadControlAll)
			auth.POST("/lead-control/create", h.CreateOrUpdateLeadControl)
			auth.DELETE("/lead-control/:id", h.DeleteLeadControl)
		}
		// Agent routes
		api.GET("/agent/:id", h.GetAgent)
		api.PUT("/agent/:id", h.UpdateAgent)
		api.DELETE("/agent/:id", h.DeleteAgent)
		api.POST("/leads", leadHandler.CreateLead)
		api.GET("/leads/:id", leadHandler.GetLead)
		api.GET("/leads/findByPhoneOrEmail", leadHandler.GetLeadByPhoneOrEmail)
		api.PUT("/leads/:id", leadHandler.UpdateLead)
		api.PUT("/leads/:id/schedule", leadHandler.UpdateLeadSchedule)
		api.POST("/leads/:id/notes", leadHandler.AddLeadNote)
		api.DELETE("/leads/:id", leadHandler.DeleteLead)
		api.GET("/leads/all", leadHandler.GetAllLeads)
		api.GET("/leads/serve/:id", leadHandler.ServeProfilePic)
		api.PUT("/leads/fav/:id", leadHandler.ToggleFavorite)
		api.PUT("/leads/updatestatus/:id", leadHandler.UpdateStatus)
		api.GET("/serve/:filename", leadHandler.ServeFile)
		api.PUT("/leads/details/:id", leadHandler.LeadDetails)
		// Job routes
		api.GET("/jobs", jobHandler.GetAllJobs)
		api.PUT("/jobs/:id/sync", jobHandler.SyncJob)
		// // Task routes
		api.POST("/tasks", h.CreateTask)
		api.PUT("/tasks/:id/photos", fileHandler.UploadTaskPhotos)
		api.PUT("/tasks/:id", h.UpdateTask)
		api.DELETE("/tasks/:id", h.DeleteTask)
		api.GET("/tasks", h.GetAllTasks)
		api.GET("/associatecustomers", h.GetAllAssociateCustomers)
		// customers
		api.GET("/customers/all", h.GetAllCustomer)
		api.POST("/customers", h.CreateCustomer)
		// // File routes
		api.POST("/files", fileHandler.CreateFile)
		api.GET("/files/:id", fileHandler.GetFile)
		api.PUT("/files/:id", fileHandler.UpdateFile)
		api.DELETE("/files/:id", fileHandler.DeleteFile)
		api.PUT("/updatecategory/:id", fileHandler.UpdateFileCategory)
		// ResidentialEntry routes
		api.DELETE("/residentialentries/:id", h.DeleteResidentialEntry)
		api.GET("/residentialentries", h.GetAllResidentialEntry)
		api.GET("/residentialentries/lead/:id", h.GetAllByLead)
		api.GET("/residentialentries/onlylead/:id", h.GetOnlyByLead)
		// // Estimate routes
		api.POST("/estimates", h.CreateEstimate)
		api.GET("/estimates/:id", h.GetEstimate)
		api.PUT("/estimates/:id", h.UpdateEstimate)
		api.PUT("/estimates/:id/schedule", h.UpdateEstimateSchedule)
		api.POST("/estimates/:id/notes", h.AddEstimateNote)
		api.PUT("/jobestimates/:id", h.UpdateJobEstimate)
		api.POST("/estimates/:id/stages", h.CreateStagesForEstimate)
		api.GET("/estimates/all", h.GetAllEstimates)
		api.POST("/duplicates/:id", h.DuplicateEstimate)
		api.POST("/duplicatesjob/:id", h.DuplicateEstimateJob)
		api.GET("/versions/:id", h.GetEstimateVersions)
		api.POST("/versiontobase/:id", h.PromoteVersionToBase)
		api.GET("/estimates/:id/activity", h.GetEstimateActivity)
		api.GET("/jobs/:id/activity", jobHandler.GetJobActivity)
		// Estimate Project Management routes
		api.GET("/estimates/:id/project-management", h.GetEstimateProjectManagement)
		api.POST("/estimates/:id/project-management", h.MutateEstimateProjectManagement)
		api.GET("/materialrows/:id", h.GetMaterialRow)
		//stage routes
		api.GET("/lead/:lead_id", fileHandler.GetStagesByLeadID())
		api.PUT("/lead/:lead_id/stage/:stage_id/status", h.UpdateStageStatus())
		api.POST("/lead/:lead_id/stage/:stage_id/task", h.CreateStageTask())
		api.PUT("/lead/:lead_id/stage/:stage_id/task/:task_id", h.UpdateStageTask())
		api.POST("/lead/:lead_id/stage/:stage_id/user", h.AddStageUser())
		api.PUT("/lead/:lead_id/stage/:stage_id/user", h.UpdateStageUser())
		api.GET("/lead/:lead_id/stage/:stage_id/users", h.GetStageUsers())
		api.DELETE("/lead/:lead_id/stage/:stage_id/user/:user_id", h.DeleteStageUser())
		api.POST("/task/:id/timer/start", h.StartTaskTimer())
		api.POST("/task/:id/timer/pause", h.PauseTaskTimer())
		api.POST("/task/:id/timer/complete", h.CompleteTaskTimer())
		api.POST("/task/:id/timer/cancel", h.CancelTask())
		api.DELETE("/task/:id", h.DeleteTaskRecord())
		api.GET("/stagetask/:id", h.TaskCompletionStatus)
		api.GET("/task/:id/timer", h.GetTaskTimerStatus())
		api.GET("/agentwithstages", h.AgentWithStages())
		api.GET("/stagetasks", h.GetAllStageTasks)
		api.POST("/lead/:lead_id/stage/:stage_id/note", h.AddStageNote())
		api.DELETE("/lead/:lead_id/stage/:stage_id/note/:note_id", h.DeleteStageNote())
		api.GET("/stages", h.GetAllStages())
		// Report email route
		api.POST("/reports/send-email", h.SendReportEmail)

		scratchpads := api.Group("/scratchpads")
		{
			scratchpads.POST("", h.CreateScratchpad)
			scratchpads.GET("", h.GetAllScratchpads)
			scratchpads.GET("/:id", h.GetScratchpad)
			scratchpads.PUT("/:id", h.UpdateScratchpad)
			scratchpads.DELETE("/:id", h.DeleteScratchpad)
			scratchpads.POST("/:id/send-email", h.SendAsEmail)
			scratchpads.GET("/by-agent/:agent_id", h.GetScratchpadsByAgent)
		}
	}
}
