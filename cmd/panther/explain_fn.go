package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var explainFnCmd = &cobra.Command{
	Use:   "explain-fn",
	Short: "Explain function",
	Long:  "Explain function",
	Args:  cobra.ExactArgs(1),
	RunE:  runExplainFunctionCmd,
}

const explainFunctionPromptSuffix = `
---
Explain the above code.
You start by by giving a high level explanation of what the code does 
then go into details, and give a reason why the code does that.
`

func init() {
	rootCmd.AddCommand(explainFnCmd)
}

type searchResult struct {
	FilePath     string
	FunctionName string
	Content      string
}

func searchFunction(path, targetFunctionPattern string) ([]searchResult, error) {
	// Validate the targetFunctionRegex pattern
	targetFunctionRegex, err := regexp.Compile(targetFunctionPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid targetFunctionRegex pattern: %v", err)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	var results []searchResult

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		filePath := filepath.Join(path, file.Name())

		if file.IsDir() {
			subResults, err := searchFunction(filePath, targetFunctionPattern)
			if err == nil && len(subResults) > 0 {
				results = append(results, subResults...)
			}
		} else if strings.HasSuffix(file.Name(), ".go") {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("error reading file: %v", err)
			}

			fileSet := token.NewFileSet()
			fileAst, err := parser.ParseFile(fileSet, filePath, string(content), parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("error parsing file: %v", err)
			}

			for _, decl := range fileAst.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					funcName := d.Name.Name
					matched := targetFunctionRegex.MatchString(funcName)
					if matched {
						fileContent := string(content[d.Pos()-1 : d.End()-1])
						if d.Doc != nil {
							fileContent = string(content[d.Doc.Pos()-1:d.Doc.End()-1]) + fileContent
						}
						result := searchResult{
							FilePath:     filePath,
							FunctionName: funcName,
							Content:      fileContent,
						}
						results = append(results, result)
					}
				}
			}
		}
	}

	if len(results) == 0 {
		return nil, errors.New("function not found")
	}

	return results, nil
}

func runExplainFunctionCmd(cmd *cobra.Command, args []string) error {
	dirPath := "."
	targetFunctionPattern := args[0]

	functionContents, err := searchFunction(dirPath, targetFunctionPattern)
	if err != nil {
		return err
	}

	if len(functionContents) > 1 {
		fmt.Println("Multiple functions found, please specify the function name more precisely.")
		for _, functionContent := range functionContents {
			fmt.Println("  - ", functionContent.FunctionName, "in", functionContent.FilePath)
		}
		return nil
	}

	_, err = client.ChatCompletion(context.Background(), gpt3.ChatCompletionRequest{
		Model: gpt3.GPT3Dot5Turbo0301,
		Messages: []gpt3.ChatCompletionRequestMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: systemPrompt,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error getting system init: %v", err)
	}

	contents := []string{
		fmt.Sprintf("# Explain Function `%s`\n", functionContents[0].FunctionName),
	}

	fmt.Println("Explaining function", functionContents[0].FunctionName, "...")
	resp, err := client.ChatCompletion(context.Background(), gpt3.ChatCompletionRequest{
		Messages: []gpt3.ChatCompletionRequestMessage{
			{
				Role:    "user",
				Content: functionContents[0].Content + "\n" + explainFunctionPromptSuffix,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error processing function:%v", err)
	}

	contents = append(contents, resp.Choices[0].Message.Content)
	content := strings.Join(contents, "")
	if err := os.WriteFile("fn_"+functionContents[0].FunctionName+"_explained.md", []byte(content), 0644); err != nil {
		return err
	}

	model, err := newExample(content)
	if err != nil {
		return err
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		tea.WithMouseAllMotion(),
		tea.WithInputTTY(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
