package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lockb0x-llc/relayforge/internal/auth"
	"github.com/lockb0x-llc/relayforge/internal/models"
	"github.com/lockb0x-llc/relayforge/internal/workflow"
	"github.com/lockb0x-llc/relayforge/pkg/types"
)

type Server struct {
	db       *gorm.DB
	router   *gin.Engine
	auth     *auth.AuthService
	workflow *workflow.Service
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	// Database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "relayforge"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "relayforge"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Workflow{}, &models.Run{}, 
		&models.Job{}, &models.Step{}, &models.Log{}, &models.Runner{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize services
	authService := auth.NewAuthService(
		getEnv("GITHUB_CLIENT_ID", ""),
		getEnv("GITHUB_CLIENT_SECRET", ""),
		getEnv("JWT_SECRET", "your-secret-key"),
	)

	workflowService := workflow.NewService(db)

	router := gin.Default()
	
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	server := &Server{
		db:       db,
		router:   router,
		auth:     authService,
		workflow: workflowService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for demo
			},
		},
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "relayforge-api"})
	})

	// Auth routes
	auth := s.router.Group("/api/auth")
	{
		auth.GET("/github", s.githubAuth)
		auth.GET("/callback", s.githubCallback)
		auth.GET("/user", s.authMiddleware(), s.getUser)
	}

	// API routes
	api := s.router.Group("/api")
	api.Use(s.authMiddleware())
	{
		// Workflows
		api.GET("/workflows", s.getWorkflows)
		api.POST("/workflows", s.createWorkflow)
		api.GET("/workflows/:id", s.getWorkflow)
		api.PUT("/workflows/:id", s.updateWorkflow)
		api.DELETE("/workflows/:id", s.deleteWorkflow)

		// Runs
		api.GET("/workflows/:id/runs", s.getWorkflowRuns)
		api.POST("/workflows/:id/runs", s.createRun)
		api.GET("/runs/:id", s.getRun)
		api.POST("/runs/:id/cancel", s.cancelRun)

		// Runners
		api.GET("/runners", s.getRunners)
		api.POST("/runners/register", s.registerRunner)
	}

	// WebSocket for logs
	s.router.GET("/ws/logs/:runId", s.authMiddleware(), s.streamLogs)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Auth handlers
func (s *Server) githubAuth(c *gin.Context) {
	url := s.auth.GetGitHubAuthURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (s *Server) githubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	user, token, err := s.auth.HandleGitHubCallback(code, s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (s *Server) getUser(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		user, err := s.auth.ValidateToken(token, s.db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// Workflow handlers
func (s *Server) getWorkflows(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	workflows, err := s.workflow.GetUserWorkflows(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workflows": workflows})
}

func (s *Server) createWorkflow(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		YAMLContent string `json:"yaml_content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow := &models.Workflow{
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		YAMLContent: req.YAMLContent,
	}

	if err := s.workflow.CreateWorkflow(workflow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"workflow": workflow})
}

func (s *Server) getWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	workflow, err := s.workflow.GetWorkflow(uint(id), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workflow": workflow})
}

func (s *Server) updateWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		YAMLContent string `json:"yaml_content"`
		IsActive    *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := s.workflow.UpdateWorkflow(uint(id), user.ID, req.Name, req.Description, req.YAMLContent, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workflow": workflow})
}

func (s *Server) deleteWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	if err := s.workflow.DeleteWorkflow(uint(id), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted"})
}

// Run handlers
func (s *Server) getWorkflowRuns(c *gin.Context) {
	workflowID, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	runs, err := s.workflow.GetWorkflowRuns(uint(workflowID), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"runs": runs})
}

func (s *Server) createRun(c *gin.Context) {
	workflowID, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	var req types.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := s.workflow.CreateRun(uint(workflowID), user.ID, req.Inputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"run": run})
}

func (s *Server) getRun(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	run, err := s.workflow.GetRun(uint(id), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"run": run})
}

func (s *Server) cancelRun(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user := c.MustGet("user").(*models.User)

	if err := s.workflow.CancelRun(uint(id), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Run cancelled"})
}

// Runner handlers
func (s *Server) getRunners(c *gin.Context) {
	var runners []models.Runner
	if err := s.db.Find(&runners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"runners": runners})
}

func (s *Server) registerRunner(c *gin.Context) {
	var req types.RunnerRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	runner := &models.Runner{
		ID:      fmt.Sprintf("runner-%d", time.Now().Unix()),
		Name:    req.Name,
		Version: req.Version,
		Status:  "online",
	}

	if err := s.db.Create(runner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"runner": runner})
}

// WebSocket log streaming
func (s *Server) streamLogs(c *gin.Context) {
	runID, _ := strconv.Atoi(c.Param("runId"))
	user := c.MustGet("user").(*models.User)

	// Verify user has access to this run
	var run models.Run
	if err := s.db.Where("id = ? AND user_id = ?", runID, user.ID).First(&run).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
		return
	}

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Stream logs for this run
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastLogID := uint(0)

	for {
		select {
		case <-ticker.C:
			var logs []models.Log
			err := s.db.Joins("JOIN steps ON logs.step_id = steps.id").
				Joins("JOIN jobs ON steps.job_id = jobs.id").
				Where("jobs.run_id = ? AND logs.id > ?", runID, lastLogID).
				Order("logs.id ASC").
				Find(&logs).Error

			if err != nil {
				log.Printf("Error fetching logs: %v", err)
				continue
			}

			for _, logEntry := range logs {
				entry := types.LogEntry{
					RunID:     uint(runID),
					Content:   logEntry.Content,
					Level:     logEntry.Level,
					Timestamp: logEntry.Timestamp.Format(time.RFC3339),
				}

				if err := conn.WriteJSON(entry); err != nil {
					log.Printf("WebSocket write failed: %v", err)
					return
				}

				lastLogID = logEntry.ID
			}
		}
	}
}