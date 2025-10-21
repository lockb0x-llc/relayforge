package workflow

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/lockb0x-llc/relayforge/internal/models"
	"github.com/lockb0x-llc/relayforge/pkg/types"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Workflow management
func (s *Service) GetUserWorkflows(userID uint) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error
	return workflows, err
}

func (s *Service) CreateWorkflow(workflow *models.Workflow) error {
	// Validate YAML content
	var spec types.WorkflowSpec
	if err := yaml.Unmarshal([]byte(workflow.YAMLContent), &spec); err != nil {
		return fmt.Errorf("invalid YAML content: %v", err)
	}

	return s.db.Create(workflow).Error
}

func (s *Service) GetWorkflow(id, userID uint) (*models.Workflow, error) {
	var workflow models.Workflow
	err := s.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Runs").
		First(&workflow).Error
	return &workflow, err
}

func (s *Service) UpdateWorkflow(id, userID uint, name, description, yamlContent string, isActive *bool) (*models.Workflow, error) {
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&workflow).Error; err != nil {
		return nil, err
	}

	if name != "" {
		workflow.Name = name
	}
	if description != "" {
		workflow.Description = description
	}
	if yamlContent != "" {
		// Validate YAML content
		var spec types.WorkflowSpec
		if err := yaml.Unmarshal([]byte(yamlContent), &spec); err != nil {
			return nil, fmt.Errorf("invalid YAML content: %v", err)
		}
		workflow.YAMLContent = yamlContent
	}
	if isActive != nil {
		workflow.IsActive = *isActive
	}

	err := s.db.Save(&workflow).Error
	return &workflow, err
}

func (s *Service) DeleteWorkflow(id, userID uint) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Workflow{}).Error
}

// Run management
func (s *Service) GetWorkflowRuns(workflowID, userID uint) ([]models.Run, error) {
	var runs []models.Run
	err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).
		Preload("Jobs").
		Order("created_at DESC").
		Find(&runs).Error
	return runs, err
}

func (s *Service) CreateRun(workflowID, userID uint, inputs map[string]string) (*models.Run, error) {
	// Get workflow
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return nil, err
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("workflow is not active")
	}

	// Parse workflow YAML
	var spec types.WorkflowSpec
	if err := yaml.Unmarshal([]byte(workflow.YAMLContent), &spec); err != nil {
		return nil, fmt.Errorf("invalid workflow YAML: %v", err)
	}

	// Create run
	run := &models.Run{
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "pending",
	}

	if err := s.db.Create(run).Error; err != nil {
		return nil, err
	}

	// Create jobs
	for jobName, jobSpec := range spec.Jobs {
		job := &models.Job{
			RunID:  run.ID,
			Name:   jobName,
			Status: "pending",
		}

		if err := s.db.Create(job).Error; err != nil {
			return nil, err
		}

		// Create steps
		for i, stepSpec := range jobSpec.Steps {
			stepName := stepSpec.Name
			if stepName == "" {
				stepName = fmt.Sprintf("Step %d", i+1)
			}

			step := &models.Step{
				JobID:   job.ID,
				Name:    stepName,
				Command: stepSpec.Run,
				Status:  "pending",
			}

			if err := s.db.Create(step).Error; err != nil {
				return nil, err
			}
		}
	}

	// Load the complete run with jobs
	if err := s.db.Preload("Jobs.Steps").First(run, run.ID).Error; err != nil {
		return nil, err
	}

	// Queue the run for execution (in a real system, this would be sent to a job queue)
	go s.executeRun(run)

	return run, nil
}

func (s *Service) GetRun(id, userID uint) (*models.Run, error) {
	var run models.Run
	err := s.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Workflow").
		Preload("Jobs.Steps.Logs").
		First(&run).Error
	return &run, err
}

func (s *Service) CancelRun(id, userID uint) error {
	var run models.Run
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&run).Error; err != nil {
		return err
	}

	if run.Status == "running" || run.Status == "pending" {
		run.Status = "cancelled"
		finishedAt := time.Now()
		run.FinishedAt = &finishedAt
		return s.db.Save(&run).Error
	}

	return fmt.Errorf("run cannot be cancelled in current status: %s", run.Status)
}

// Simulate run execution (in real system, this would be handled by runners)
func (s *Service) executeRun(run *models.Run) {
	// Update run status
	run.Status = "running"
	startedAt := time.Now()
	run.StartedAt = &startedAt
	s.db.Save(run)

	// Execute jobs sequentially (simplified)
	var jobs []models.Job
	s.db.Where("run_id = ?", run.ID).Preload("Steps").Find(&jobs)

	runSuccess := true
	for _, job := range jobs {
		job.Status = "running"
		jobStartedAt := time.Now()
		job.StartedAt = &jobStartedAt
		s.db.Save(&job)

		// Execute steps
		jobSuccess := true
		for _, step := range job.Steps {
			step.Status = "running"
			stepStartedAt := time.Now()
			step.StartedAt = &stepStartedAt
			s.db.Save(&step)

			// Simulate step execution
			time.Sleep(2 * time.Second)

			// Create log entry
			log := &models.Log{
				StepID:  step.ID,
				Content: fmt.Sprintf("Executing: %s", step.Command),
				Level:   "info",
			}
			s.db.Create(log)

			// Simulate step completion
			step.Status = "success"
			stepFinishedAt := time.Now()
			step.FinishedAt = &stepFinishedAt
			s.db.Save(&step)

			// Create completion log
			completionLog := &models.Log{
				StepID:  step.ID,
				Content: fmt.Sprintf("Step completed successfully"),
				Level:   "info",
			}
			s.db.Create(completionLog)
		}

		// Update job status
		if jobSuccess {
			job.Status = "success"
		} else {
			job.Status = "failed"
			runSuccess = false
		}
		jobFinishedAt := time.Now()
		job.FinishedAt = &jobFinishedAt
		s.db.Save(&job)
	}

	// Update run status
	if runSuccess {
		run.Status = "success"
	} else {
		run.Status = "failed"
	}
	finishedAt := time.Now()
	run.FinishedAt = &finishedAt
	s.db.Save(run)
}