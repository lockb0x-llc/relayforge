package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "relayforge",
	Short: "RelayForge CLI - Infrastructure orchestration platform",
	Long: `RelayForge CLI is a command-line interface for the RelayForge 
infrastructure orchestration platform. Use it to manage workflows,
runs, and monitor your infrastructure automation.`,
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.relayforge.yaml)")
	rootCmd.PersistentFlags().String("api-url", "http://localhost:8080", "RelayForge API URL")
	rootCmd.PersistentFlags().String("token", "", "Authentication token")
	
	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".relayforge")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("RelayForge CLI v1.0.0")
	},
}

// Auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to RelayForge",
	Run: func(cmd *cobra.Command, args []string) {
		apiURL := viper.GetString("api-url")
		fmt.Printf("Open the following URL in your browser to login:\n")
		fmt.Printf("%s/api/auth/github\n", apiURL)
		fmt.Printf("\nAfter authentication, set your token using:\n")
		fmt.Printf("relayforge auth set-token <your-token>\n")
	},
}

var setTokenCmd = &cobra.Command{
	Use:   "set-token [token]",
	Short: "Set authentication token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]
		viper.Set("token", token)
		
		// Save to config file
		home, _ := os.UserHomeDir()
		configPath := fmt.Sprintf("%s/.relayforge.yaml", home)
		viper.WriteConfigAs(configPath)
		
		fmt.Printf("Token saved to %s\n", configPath)
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(setTokenCmd)
}

// Workflow commands
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Workflow management commands",
}

var listWorkflowsCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows",
	Run: func(cmd *cobra.Command, args []string) {
		workflows, err := apiCall("GET", "/api/workflows", nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		
		fmt.Printf("%-5s %-20s %-15s %-10s\n", "ID", "Name", "Description", "Status")
		fmt.Println("-----------------------------------------------------------")
		
		if workflowData, ok := workflows["workflows"].([]interface{}); ok {
			for _, w := range workflowData {
				if workflow, ok := w.(map[string]interface{}); ok {
					id := workflow["id"]
					name := workflow["name"]
					desc := workflow["description"]
					active := workflow["is_active"]
					status := "inactive"
					if active == true {
						status = "active"
					}
					fmt.Printf("%-5v %-20s %-15s %-10s\n", id, name, desc, status)
				}
			}
		}
	},
}

var createWorkflowCmd = &cobra.Command{
	Use:   "create [name] [yaml-file]",
	Short: "Create a new workflow",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		yamlFile := args[1]
		
		content, err := os.ReadFile(yamlFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}
		
		payload := map[string]string{
			"name":         name,
			"yaml_content": string(content),
		}
		
		result, err := apiCall("POST", "/api/workflows", payload)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		
		fmt.Printf("Workflow created successfully!\n")
		if workflow, ok := result["workflow"].(map[string]interface{}); ok {
			fmt.Printf("ID: %v\n", workflow["id"])
			fmt.Printf("Name: %v\n", workflow["name"])
		}
	},
}

func init() {
	workflowCmd.AddCommand(listWorkflowsCmd)
	workflowCmd.AddCommand(createWorkflowCmd)
}

// Run commands
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Workflow run commands",
}

var startRunCmd = &cobra.Command{
	Use:   "start [workflow-id]",
	Short: "Start a workflow run",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowID := args[0]
		
		payload := map[string]interface{}{
			"workflow_id": workflowID,
		}
		
		result, err := apiCall("POST", fmt.Sprintf("/api/workflows/%s/runs", workflowID), payload)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		
		fmt.Printf("Run started successfully!\n")
		if run, ok := result["run"].(map[string]interface{}); ok {
			fmt.Printf("Run ID: %v\n", run["id"])
			fmt.Printf("Status: %v\n", run["status"])
		}
	},
}

var listRunsCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List workflow runs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowID := args[0]
		
		runs, err := apiCall("GET", fmt.Sprintf("/api/workflows/%s/runs", workflowID), nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		
		fmt.Printf("%-5s %-15s %-20s %-20s\n", "ID", "Status", "Started", "Finished")
		fmt.Println("----------------------------------------------------------------")
		
		if runData, ok := runs["runs"].([]interface{}); ok {
			for _, r := range runData {
				if run, ok := r.(map[string]interface{}); ok {
					id := run["id"]
					status := run["status"]
					started := run["started_at"]
					finished := run["finished_at"]
					
					if started == nil {
						started = "-"
					}
					if finished == nil {
						finished = "-"
					}
					
					fmt.Printf("%-5v %-15v %-20v %-20v\n", id, status, started, finished)
				}
			}
		}
	},
}

func init() {
	runCmd.AddCommand(startRunCmd)
	runCmd.AddCommand(listRunsCmd)
}

// API helper function
func apiCall(method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	// This is a simplified implementation
	// In a real CLI, you would use HTTP client to make actual API calls
	fmt.Printf("API Call: %s %s\n", method, endpoint)
	if payload != nil {
		payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Printf("Payload: %s\n", payloadJSON)
	}
	
	// Return mock response for demo
	return map[string]interface{}{
		"message": "Success",
	}, nil
}