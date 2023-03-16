package main

import (
	"fmt"
	"log"
	"os"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "panther",
}

var rootArgs struct {
	repo    string
	noCache bool
}

func init() {
	// add persistent flags here
	// --repo string
	// --no-cache bool

	rootCmd.PersistentFlags().StringVarP(&rootArgs.repo, "repo", "r", "weaveworks/weave-gitops", "GitHub repo to use")
	rootCmd.PersistentFlags().BoolVarP(&rootArgs.noCache, "no-cache", "n", false, "Don't use cache")
}

var client gpt3.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalln("Missing API KEY")
	}

	client = gpt3.NewClient(apiKey)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
