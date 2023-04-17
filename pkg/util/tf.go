package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"regexp"
)

func HandleCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Please provide a command.")
		os.Exit(1)
	}

	command := args[1]
	text := args[2]

	switch command {
	case "plan":
		ExplainCommand("plan")
	case "validate":
		ExplainCommand("validate")
	case "destroy":
		ExplainCommand("destroy")
	case "init":
		ExplainCommand("init")
	case "show":
		ExplainCommand("show")
	case "generate":
		GenerateCode(command, text)
	case "concept":
		if len(args) < 3 {
			fmt.Println("Please provide a concept.")
			os.Exit(1)
		}
		concept := args[2]
		ExplainConcept(concept)
	default:
		fmt.Printf("Unsupported command: %s\n", command)
		os.Exit(1)
	}
}

func GenerateCode(command string, text string) {
	code, err := GenerateCodeFromChatGPT(text)
	if err != nil {
		fmt.Printf("Error getting explanation from ChatGPT: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(Colorize("Terraform %s output\n\n", Green), command)
	fmt.Println("RAW")
	fmt.Println(code)
	fmt.Println("END RAW")
	fmt.Println(parseIac(code))
}

func parseIac(code string) string {
	pattern := regexp.MustCompile("`{3}([\\w]*)\n([\\S\\s]+?)\n`{3}")
	result := pattern.FindAllStringSubmatch(code, -1)
	var codes []string
	for _, v := range result {
		codes = append(codes, v[1:]...)
	}
	return strings.Join(codes, "\n")
}

func ExplainCommand(command string) {
	var cmd *exec.Cmd
	if command == "destroy" {
		cmd = exec.Command("terraform", "plan", "-destroy")
	} else {
		cmd = exec.Command("terraform", command)
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()

	if err != nil {
		errOutput := errBuf.String()
		explanation, err := GetExplanationFromChatGPT(errOutput, "command", command)
		if err != nil {
			fmt.Printf("Error getting explanation from ChatGPT: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf(Colorize("Error encountered while running 'terraform %s':\n", Red), command)
		fmt.Println(explanation)
	} else {
		output := outBuf.String()
		betweenDelimiters := false
		if command == "plan" || command == "destroy" || command == "show" {
			lines := strings.Split(output, "\n")
			var sb strings.Builder
			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if strings.Contains(trimmedLine, "#") {
					sb.WriteString(trimmedLine)
					sb.WriteString("\n")
				}
				if strings.Contains(trimmedLine, "Changes to Outputs:") {
					betweenDelimiters = true
					sb.WriteString(trimmedLine)
					sb.WriteString("\n")
					continue
				}

				if strings.Contains(trimmedLine, "─────────────────────") {
					betweenDelimiters = false
					break
				}

				if strings.Contains(trimmedLine, "No changes") {
					sb.WriteString(trimmedLine)
					break
				}

				if betweenDelimiters {
					sb.WriteString(trimmedLine)
					sb.WriteString("\n")
				}
			}
			output = sb.String()
		}
		explanation, err := GetExplanationFromChatGPT(output, "command", command)
		if err != nil {
			fmt.Printf("Error getting explanation from ChatGPT: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf(Colorize("Terraform %s output explained\n\n", Green), command)
		fmt.Println(explanation)
	}
}

func ExplainConcept(concept string) {
	explanation, err := GetExplanationFromChatGPT(concept, "concept", "command")
	if err != nil {
		fmt.Printf("Error getting explanation from ChatGPT: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(Colorize("Explain the following concept: %s \n\n", Green), concept)
	fmt.Println(explanation)
}
