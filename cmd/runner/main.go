package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lockb0x-llc/relayforge/pkg/types"
)

type Runner struct {
	ID      string
	Name    string
	Version string
	Tags    []string
	ApiURL  string
	Token   string
	client  *http.Client
}

func main() {
	runner := &Runner{
		ID:      fmt.Sprintf("runner-%d", time.Now().Unix()),
		Name:    getEnv("RUNNER_NAME", "relayforge-runner"),
		Version: "1.0.0",
		Tags:    strings.Split(getEnv("RUNNER_TAGS", "linux,shell"), ","),
		ApiURL:  getEnv("API_URL", "http://localhost:8080"),
		Token:   getEnv("RUNNER_TOKEN", ""),
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	log.Printf("Starting RelayForge Runner %s", runner.ID)

	// Register runner
	if err := runner.register(); err != nil {
		log.Fatal("Failed to register runner:", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down runner...")
		cancel()
	}()

	// Start job polling
	runner.startJobPolling(ctx)
}

func (r *Runner) register() error {
	registration := types.RunnerRegistration{
		Name:    r.Name,
		Version: r.Version,
		Tags:    r.Tags,
	}

	payload, _ := json.Marshal(registration)
	req, err := http.NewRequest("POST", r.ApiURL+"/api/runners/register", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", body)
	}

	log.Printf("Runner registered successfully: %s", r.ID)
	return nil
}

func (r *Runner) startJobPolling(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.pollForJobs()
		}
	}
}

func (r *Runner) pollForJobs() {
	// In a real implementation, this would poll the API for available jobs
	// For this demo, we'll simulate job execution
	log.Println("Polling for jobs...")
}

func (r *Runner) executeJob(assignment types.JobAssignment) {
	log.Printf("Executing job %d for run %d", assignment.JobID, assignment.RunID)

	// Parse workflow and job
	jobSpec := assignment.JobSpec

	// Execute steps
	for i, step := range jobSpec.Steps {
		if err := r.executeStep(assignment.JobID, uint(i+1), step); err != nil {
			log.Printf("Step failed: %v", err)
			r.reportJobResult(types.JobResult{
				JobID:  assignment.JobID,
				Status: "failed",
				Error:  err.Error(),
			})
			return
		}
	}

	// Report success
	r.reportJobResult(types.JobResult{
		JobID:  assignment.JobID,
		Status: "success",
	})
}

func (r *Runner) executeStep(jobID, stepID uint, step types.StepSpec) error {
	log.Printf("Executing step: %s", step.Name)
	
	if step.Run == "" {
		return fmt.Errorf("no command specified for step")
	}

	// Prepare command
	cmd := exec.Command("sh", "-c", step.Run)
	
	// Set working directory if specified
	if step.WorkingDir != "" {
		cmd.Dir = step.WorkingDir
	}

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range step.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start step
	startTime := time.Now()
	r.reportStepResult(types.StepResult{
		StepID:    stepID,
		Status:    "running",
		StartedAt: startTime.Format(time.RFC3339),
	})

	// Execute command
	err := cmd.Run()
	finishTime := time.Now()

	// Prepare result
	result := types.StepResult{
		StepID:     stepID,
		Output:     stdout.String(),
		FinishedAt: finishTime.Format(time.RFC3339),
	}

	if err != nil {
		result.Status = "failed"
		result.Error = stderr.String()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.Status = "success"
		result.ExitCode = 0
	}

	// Report result
	r.reportStepResult(result)

	return err
}

func (r *Runner) reportJobResult(result types.JobResult) {
	payload, _ := json.Marshal(result)
	log.Printf("Reporting job result: %s", payload)
	// In real implementation, this would POST to the API
}

func (r *Runner) reportStepResult(result types.StepResult) {
	payload, _ := json.Marshal(result)
	log.Printf("Reporting step result: %s", payload)
	// In real implementation, this would POST to the API
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}