package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var explainPrCmd = &cobra.Command{
	Use:   "explain-pr",
	Short: "Explain PR",
	Long:  "Explain PR",
	Args:  cobra.ExactArgs(1),
	RunE:  runExplainPrCmd,
}

func generateDiffURL(repo string, prNumber int) string {
	return fmt.Sprintf("https://patch-diff.githubusercontent.com/raw/%s/pull/%d.diff", repo, prNumber)
}

func downloadDiff(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

var ignoreSuffixes = []string{
	".swagger.json",
	".pb.go",
	".pb.gw.go",
	".pb.ts",
	".md",
	".mdx",
	"go.sum",
	"package-lock.json",
	"yarn.lock",
}

func splitDiffIntoChunks(diff string) []string {
	chunks := strings.Split(diff, "diff --git")
	// Remove empty chunk caused by the first "diff --git"
	chunks = chunks[1:]

	var result []string
	for _, chunk := range chunks {
		parts := strings.SplitN(chunk, "\n", 2)
		ignore := false
		for _, suffix := range ignoreSuffixes {
			if strings.HasSuffix(strings.TrimSpace(parts[0]), suffix) {
				ignore = true
				break
			}
		}
		if !ignore {
			result = append(result, "diff --git"+chunk)
		}
	}

	return result
}

const systemPrompt = `
Please act as a Senior Software Engineer, who is specialized in the Go programming language, and Kubernetes.
You have been working in the Go compiler team and Kubernetes team for a long time.
You are very smart and best at reviewing Go codes. The codes to be explained will be in the Git diff format.
`

const promptSuffix = `
---
Explain the above diff. 
You must start each paragraph by saying that "In file <file name>, this PR <adds / changes / fixes>".
You explain by giving a high level description, 
then go into details for each change, and give a reason why you need the change. 
Do not need to explain about whitespace changes.
`

// runExplainPrCmd runs the explain-pr command.
func runExplainPrCmd(cmd *cobra.Command, args []string) error {
	repo := rootArgs.repo

	prNumber := args[0]

	prNumberInt, err := strconv.Atoi(prNumber)
	if err != nil {
		return fmt.Errorf("error invalid PR number: %v", err)
	}

	url := generateDiffURL(repo, prNumberInt)
	fmt.Println("Downloading:", url)

	diff, err := downloadDiff(url)
	if err != nil {
		return fmt.Errorf("error downloading diff: %v", err)
	}

	chunks := splitDiffIntoChunks(diff)

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
		fmt.Sprintf("# Explain PR #%d\n", prNumberInt),
	}

	for i, chunk := range chunks {
		fmt.Println("Processing chunk:", i+1, "of", len(chunks), "...")
		resp, err := client.ChatCompletion(context.Background(), gpt3.ChatCompletionRequest{
			Messages: []gpt3.ChatCompletionRequestMessage{
				{
					Role:    "user",
					Content: chunk + "\n" + promptSuffix,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("error processing chunk: %d, %v", i, err)
		}
		contents = append(contents, resp.Choices[0].Message.Content)
	}

	content := strings.Join(contents, "")
	if err := os.WriteFile("pr_"+prNumber+"_explained.md", []byte(content), 0644); err != nil {
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

func init() {
	rootCmd.AddCommand(explainPrCmd)
}
