package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	// Define the command line parameters
	output := flag.String(
		"output", "",
		"The output path to generate the golang documentation.",
	)
	functions := flag.String(
		"functions", "", "The functions to get the documentation for.",
	)
	pkg := flag.String(
		"pkg", "", "The package name to generate the golang documentation file for.",
	)
	flag.Parse()

	if *output == "" || *functions == "" || *pkg == "" {
		log.Fatal("error: output, functions and pkg flags are required")
	}

	result, err := generateDocs(*functions, *pkg)
	if err != nil {
		log.Fatal(err)
	}

	// Create file or truncate existing.
	log.Println("Writing result to file...")
	log.Println(result)

	err = os.WriteFile(*output, []byte(result), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func generateDocs(functions, pkg string) (string, error) {
	// Split the functions string into a slice
	funcSlice := strings.Split(functions, ",")

	funcNames := make([]string, 0, len(funcSlice))
	funcInfo := make(map[string][]string)
	for _, function := range funcSlice {
		functionName, comments, err := getFunctionInfo(function)
		if err != nil {
			return "", err
		}

		funcNames = append(funcNames, functionName)
		funcInfo[functionName] = comments
	}

	var sb strings.Builder
	sb.WriteString("// Code generated automatically by ocp-go-utils/docs. DO NOT EDIT.\n\n")
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkg))

	for i, funcName := range funcNames {
		comments := funcInfo[funcName]
		for _, comment := range comments {
			sb.WriteString(comment + "\n")
		}

		sb.WriteString(fmt.Sprintf("func %s() {}", funcName))

		if i < len(funcNames)-1 {
			sb.WriteString("\n\n")
		} else {
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// getFunctionInfo returns the function name and the comments above.
func getFunctionInfo(f string) (string, []string, error) {
	// Execute go doc command to get the function information
	cmd := exec.Command("bash", "-c", "go doc -src -short "+f)
	bOutput, err := cmd.Output()
	if err != nil {
		return "", nil, err
	}

	output := string(bOutput)

	// Use a regular expression to match the function signature
	funcRegex := regexp.MustCompile(`func \(\w* \*\w*\) (\w*)\(`)
	funcMatch := funcRegex.FindStringSubmatch(output)

	// Use a regular expression to match the comments
	commentRegex := regexp.MustCompile(`//.*`)
	commentMatches := commentRegex.FindAllString(output, -1)

	if len(funcMatch) == 0 {
		return "", nil, errors.New("error: no function signature found")
	}

	if len(commentMatches) == 0 {
		return "", nil, errors.New("error: no function comments found")
	}

	return funcMatch[1], commentMatches, nil
}
