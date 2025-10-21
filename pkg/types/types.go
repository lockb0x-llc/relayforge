package types

// WorkflowSpec represents the YAML workflow specification
type WorkflowSpec struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description,omitempty"`
	On          map[string]interface{} `yaml:"on,omitempty"`
	Jobs        map[string]JobSpec `yaml:"jobs"`
}

// JobSpec represents a job in the workflow
type JobSpec struct {
	Name     string     `yaml:"name,omitempty"`
	RunsOn   string     `yaml:"runs-on"`
	Needs    []string   `yaml:"needs,omitempty"`
	If       string     `yaml:"if,omitempty"`
	Steps    []StepSpec `yaml:"steps"`
	Env      map[string]string `yaml:"env,omitempty"`
	Timeout  string     `yaml:"timeout,omitempty"`
}

// StepSpec represents a step in a job
type StepSpec struct {
	Name      string            `yaml:"name,omitempty"`
	Uses      string            `yaml:"uses,omitempty"`
	Run       string            `yaml:"run,omitempty"`
	With      map[string]string `yaml:"with,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	If        string            `yaml:"if,omitempty"`
	Continue  bool              `yaml:"continue-on-error,omitempty"`
	Timeout   string            `yaml:"timeout,omitempty"`
	WorkingDir string           `yaml:"working-directory,omitempty"`
}

// RunRequest represents a request to start a workflow run
type RunRequest struct {
	WorkflowID uint              `json:"workflow_id"`
	Inputs     map[string]string `json:"inputs,omitempty"`
	Ref        string            `json:"ref,omitempty"`
}

// LogEntry represents a log entry for streaming
type LogEntry struct {
	RunID     uint   `json:"run_id"`
	JobID     uint   `json:"job_id"`
	StepID    uint   `json:"step_id"`
	Content   string `json:"content"`
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
}

// RunnerRegistration represents runner registration request
type RunnerRegistration struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Tags    []string `json:"tags"`
}

// JobAssignment represents a job assignment to a runner
type JobAssignment struct {
	JobID    uint         `json:"job_id"`
	RunID    uint         `json:"run_id"`
	JobSpec  JobSpec      `json:"job_spec"`
	Workflow WorkflowSpec `json:"workflow"`
}

// JobResult represents the result of job execution
type JobResult struct {
	JobID     uint   `json:"job_id"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

// StepResult represents the result of step execution
type StepResult struct {
	StepID     uint   `json:"step_id"`
	Status     string `json:"status"`
	ExitCode   int    `json:"exit_code"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}