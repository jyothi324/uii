package routes

import (
	"inventory/handlers"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-2"
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
	r.POST("/visualize", handlers.GeminiImageHandler)
	s3Client := s3.New(sess)
	fileHandler := handlers.NewFileHandler(s3Client, awsRegion)
	customerHandler := handlers.NewVisualizeCustomerHandler(s3Client, awsRegion)
	r.POST("/visualizer/customers", customerHandler.CreateCustomer)
	r.GET("/visualizer/check-email", customerHandler.CheckEmailExists)
	r.GET("/visualizer/check-phone", customerHandler.CheckPhoneExists)
	r.GET("/visualizer/customers/:id", customerHandler.GetCustomerByID)
	r.PUT("/visualizer/customers/:id/images", customerHandler.UpdateCustomerImages)
	r.GET("/visualizer/customers", customerHandler.GetAllCustomers)
	r.GET("/activity/recent", handlers.GetRecentActivity)

	materialRoutes := r.Group("/material")
	{
		materialRoutes.GET("/all", fileHandler.HandleFetchMaterials())
		materialRoutes.POST("/create", fileHandler.HandleCreateMaterial())
		materialRoutes.POST("/chunks/status", fileHandler.HandleBulkChunkStatus())
		materialRoutes.PUT("/chunk/:id", fileHandler.HandleMaterialChunkUpdate())
		materialRoutes.PATCH("/chunk/:id/status", fileHandler.HandleChunkStatusUpdate())
		materialRoutes.DELETE("/chunk/:id", fileHandler.HandleDeleteMaterialChunk())

		// ── Sub-group for /:id to avoid wildcard conflicts ──
		byID := materialRoutes.Group("/:id")
		{
			byID.DELETE("/image", fileHandler.HandleDeleteMaterialImage())
			byID.POST("/image/replace", fileHandler.HandleReplaceMaterialImage())
			byID.GET("", fileHandler.HandleMaterialGetByID())
			byID.PUT("", fileHandler.HandleMaterialUpdate())
			byID.DELETE("", fileHandler.DeleteMaterialByID())
		}
	}

	purchaseRoutes := r.Group("/purchase")
	{
		purchaseRoutes.GET("/all", handlers.HandleFetchPurchases())
		purchaseRoutes.POST("/create", handlers.HandleCreatePurchase())
		purchaseRoutes.GET("/:id", handlers.HandlePurchaseGetByID())
		purchaseRoutes.PUT("/:id", handlers.HandlePurchaseUpdate())
		purchaseRoutes.DELETE("/:id", handlers.DeletePurchaseByID())
	}
	prHandler := handlers.NewPurchaseReceivedHandler(s3Client, awsRegion)
	purchaseOrderReceived := r.Group("/purchase-received")
	{
		purchaseOrderReceived.GET("/all", prHandler.HandleFetchPurchaseReceiveds())
		purchaseOrderReceived.POST("/create", handlers.HandleCreatePurchaseReceived())
		purchaseOrderReceived.POST("/create-with-image", prHandler.HandleCreatePurchaseReceivedWithImage())
		purchaseOrderReceived.GET("/:id", prHandler.HandlePurchaseReceivedGetByID())
		purchaseOrderReceived.PUT("/:id", handlers.HandlePurchaseReceivedUpdate())
		purchaseOrderReceived.PUT("/:id/with-image", prHandler.HandlePurchaseReceivedUpdateWithImage())
		purchaseOrderReceived.DELETE("/:id", handlers.DeletePurchaseReceivedByID())
	}
	supplierRoutes := r.Group("/supplier")
	{
		supplierRoutes.GET("/all", handlers.GetAllSuppliers())
		supplierRoutes.POST("/create", handlers.CreateSupplier())
		supplierRoutes.GET("/:id", handlers.GetSupplierByID())
		supplierRoutes.PUT("/:id", handlers.UpdateSupplier())
		supplierRoutes.DELETE("/:id", handlers.DeleteSupplier())
	}
	feedbackRoutes := r.Group("/feedback")
	{
		feedbackRoutes.POST("/create", handlers.CreateFeedback())
	}
	materialmetaRoutes := r.Group("/metadata")
	{
		materialmetaRoutes.POST("/create", handlers.HandleCreateMetadata())
		materialmetaRoutes.GET("/all", handlers.GetAllMetadata())
		materialmetaRoutes.DELETE("/:id", handlers.DeleteMetadataByID())
	}

	settingsRoutes := r.Group("/settings")
	{
		settingsRoutes.GET("/stage-colors", handlers.GetStageColors())
		settingsRoutes.POST("/stage-colors", handlers.HandleSaveStageColors())
		settingsRoutes.GET("/terms", handlers.HandleGetTerms())
		settingsRoutes.POST("/terms", handlers.HandleSaveTerms())
		settingsRoutes.GET("/verbiage", handlers.HandleGetVerbiage())
		settingsRoutes.POST("/verbiage", handlers.HandleSaveVerbiage())
		settingsRoutes.GET("/email-templates", handlers.HandleListEmailTemplates())
		settingsRoutes.POST("/email-templates", handlers.HandleCreateEmailTemplate())
		settingsRoutes.PUT("/email-templates/:id", handlers.HandleUpdateEmailTemplate())
		settingsRoutes.DELETE("/email-templates/:id", handlers.HandleDeleteEmailTemplate())
		settingsRoutes.GET("/message-templates", handlers.HandleListMessageTemplates())
		settingsRoutes.POST("/message-templates", handlers.HandleCreateMessageTemplate())
		settingsRoutes.PUT("/message-templates/:id", handlers.HandleUpdateMessageTemplate())
		settingsRoutes.DELETE("/message-templates/:id", handlers.HandleDeleteMessageTemplate())
		settingsRoutes.GET("/job-form-questions", handlers.HandleListJobFormQuestions())
		settingsRoutes.POST("/job-form-questions", handlers.HandleCreateJobFormQuestion())
		settingsRoutes.PUT("/job-form-questions/:id", handlers.HandleUpdateJobFormQuestion())
		settingsRoutes.DELETE("/job-form-questions/:id", handlers.HandleDeleteJobFormQuestion())

		settingsRoutes.GET("/job-form-titles", handlers.HandleGetJobFormSectionTitles())
		settingsRoutes.POST("/job-form-titles", handlers.HandleUpsertJobFormSectionTitles())
	}

	formRoutes := r.Group("/forms")
	{
		formRoutes.GET("/tasks", handlers.HandleListJobTaskForms())
		formRoutes.POST("/tasks", handlers.HandleUpsertJobTaskForm())
	}

	onboardingRoutes := r.Group("/onboarding")
	{
		onboardingRoutes.POST("/create", handlers.NewOnboardingHandler(s3Client, awsRegion).HandleCreateCustomerOnboarding())
		onboardingRoutes.GET("/all", handlers.NewOnboardingHandler(s3Client, awsRegion).HandleGetAllCustomerOnboardings())
		onboardingRoutes.PUT("/:id", handlers.NewOnboardingHandler(s3Client, awsRegion).HandleUpdateCustomerOnboarding())
	}
	adminRoutes := r.Group("/admin")
	{
		adminRoutes.POST("/login", handlers.AdminLoginHandler)
	}
	tenantRoutes := r.Group("/tenant")
	{
		tenantRoutes.GET("/brand", handlers.NewOnboardingHandler(s3Client, awsRegion).GetTenantBrand())
		tenantRoutes.GET("/logo-proxy", handlers.NewOnboardingHandler(s3Client, awsRegion).ProxyLogo())
		tenantRoutes.GET("/address", handlers.NewOnboardingHandler(s3Client, awsRegion).GetTenantAddress())
		tenantRoutes.POST("/address", handlers.NewOnboardingHandler(s3Client, awsRegion).SaveTenantAddress())
	}
	purchaseReturnRoutes := r.Group("/purchase-return")
	{
		purchaseReturnRoutes.POST("/:id", handlers.HandleCreatePurchaseReturn())
	}

	auditRoutes := r.Group("/audit")
{
	auditRoutes.GET("/meta", handlers.HandleAuditMeta())

	auditRoutes.GET("/sessions", handlers.HandleAuditSessionsCollection())
	auditRoutes.POST("/sessions", handlers.HandleAuditSessionsCollection())

	auditRoutes.DELETE("/sessions/:id", handlers.HandleAuditSessionItem())
}
}
