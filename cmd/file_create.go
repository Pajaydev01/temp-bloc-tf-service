package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Default project structure
var baseDirs = []string{
	"cmd",
	"config",
	"internal",
	"api",
	"db",
	"scripts",
	"test",
	"pkg/user/api",
	"pkg/user/init",
	"pkg/user/model",
	"pkg/user/repository",
	"pkg/user/test",
	"pkg/user/usecase",
}

// Module structure
var moduleSubDirs = []string{"api", "init", "model", "repository", "test", "usecase"}

// InitProject initializes a new project
func InitProject(projectName string) {
	fmt.Println("Initializing project:", projectName)
	os.Mkdir(projectName, 0755)

	// Change directory into the project
	os.Chdir(projectName)

	// Create Go module
	runCommand("go mod init " + projectName)

	// Create base directories
	for _, dir := range baseDirs {
		createDir(dir)
	}

	// Create config.yaml
	createFile("config/config.yaml", "app_name: \""+projectName+"\"\nport: 8080\n")

	// Create main.go
	createFile("main.go", `package main

import "fmt"

func main() {
    fmt.Println("Welcome to `+projectName+`!")
}
`)

	fmt.Println("✅ Project", projectName, "initialized successfully!")
}

// CreateModule creates a new module
func CreateModule(moduleName string) {
	modulePath := filepath.Join("pkg", moduleName)

	// Check if the module already exists
	if _, err := os.Stat(modulePath); !os.IsNotExist(err) {
		fmt.Println("Module already exists:", moduleName)
		return
	}

	fmt.Println("Creating module:", moduleName)
	for _, subDir := range moduleSubDirs {
		createDir(filepath.Join(modulePath, subDir))
	}

	// Generate default files
	createFile(filepath.Join(modulePath, "model", moduleName+".go"), "package model\n\ntype "+capitalize(moduleName)+" struct {\n\tID   int\n\tName string\n}\n")
	createFile(filepath.Join(modulePath, "repository", moduleName+"_repository.go"), "package repository\n\n// "+capitalize(moduleName)+"Repository handles database operations")
	createFile(filepath.Join(modulePath, "usecase", moduleName+"_usecase.go"), "package usecase\n\n// "+capitalize(moduleName)+"UseCase handles business logic")
	createFile(filepath.Join(modulePath, "api", moduleName+"_handler.go"), "package api\n\n// "+capitalize(moduleName)+"Handler handles HTTP requests")
	createFile(filepath.Join(modulePath, "api", moduleName+"_router.go"), "package router\n\n// "+capitalize(moduleName)+"Handler handles routing")
	createFile(filepath.Join(modulePath, "test", moduleName+"_test.go"), "package test\n\nimport \"testing\"\n\nfunc Test"+capitalize(moduleName)+"(t *testing.T) {\n\t// Test cases here\n}")
	createFile(filepath.Join(modulePath, "init", "init.go"), "package init\n\n// Init function for "+capitalize(moduleName))

	fmt.Println("✅ Module", moduleName, "created successfully!")
}

// Utility function to create a directory
func createDir(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println("Error creating directory:", dir, err)
	}
}

// Utility function to create a file with default content
func createFile(filePath, content string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", filePath, err)
		return
	}
	defer file.Close()

	file.WriteString(content)
}

// Run a shell command
func runCommand(command string) {

	err := exec.Command("sh", "-c", command).Run()
	//err := os.Command("sh", "-c", command).Run()
	if err != nil {
		fmt.Println("Error running command:", command, err)
	}
}

// Capitalize first letter of a string
func capitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	return string(str[0]-32) + str[1:]
}

// handles file creation or project initialization
func HandleInit() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go init <project_name>      # Initialize project")
		fmt.Println("  go run main.go create module <name>    # Create a module")
		return
	}

	command := os.Args[1]
	arg := os.Args[2]
	fmt.Println("command:", command)
	switch command {
	case "init":
		InitProject(arg)
	case "create":
		if len(os.Args) < 4 || os.Args[2] != "module" {
			fmt.Println("Usage: go run main.go create module <name>")
			return
		}
		moduleName := os.Args[3]
		CreateModule(moduleName)
	default:
		fmt.Println("Unknown command:", command)
	}
}
