package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var summarizePrCmd = &cobra.Command{
	Use:   "summarize-pr",
	Short: "Summarize PR",
	Long:  "Summarize PR",
	Args:  cobra.ExactArgs(1),
	RunE:  runSummarizePrCmd,
}

const summarizePromptSuffix = `
---
Summarize the PR explanation into points. 
You must start each point (2 indents) by saying that "<2 indents>- In file <file name>, this PR <adds / changes / fixes>".
File name must be always in backticks.
Each point should be a single sentence, include what and why. If there's a CVE show its number.
If you have more than 1 sub-points, you can use a sub-bullet (4 indents) point.
`

func runSummarizePrCmd(cmd *cobra.Command, args []string) error {
	// read PR number from args
	prNumber := args[0]
	explained, err := os.ReadFile("pr_" + prNumber + "_explained.md")
	if err != nil {
		return err
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

	contents := []string{
		fmt.Sprintf("# PR Summary #%s\n", prNumber),
	}

	fmt.Println("Summarizing PR ...")
	resp, err := client.ChatCompletion(context.Background(), gpt3.ChatCompletionRequest{
		Messages: []gpt3.ChatCompletionRequestMessage{
			{
				Role:    "user",
				Content: string(explained) + "\n" + summarizePromptSuffix,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error summarizing: %v", err)
	}
	contents = append(contents, resp.Choices[0].Message.Content)

	content := strings.Join(contents, "")
	if err := os.WriteFile("pr_"+prNumber+"_summary.md", []byte(content), 0644); err != nil {
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
	// summarizePrCmd setup
	rootCmd.AddCommand(summarizePrCmd)
}
