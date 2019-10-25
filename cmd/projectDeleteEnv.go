package cmd

import (
	"fmt"
	"os"

	"github.com/amazeeio/lagoon-cli/graphql"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var projectDeleteEnvCmd = &cobra.Command{
	Use:   "environment [project name] [environment name]",
	Short: "Delete an environment",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 {
			fmt.Println("Not enough arguments. Requires: project name and environment.")
			cmd.Help()
			os.Exit(1)
		}
		projectName := args[0]
		projectEnvironment := args[1]

		fmt.Println(fmt.Sprintf("Deleting %s-%s", projectName, projectEnvironment))

		if yesNo() {
			var responseData DeleteEnvironmentResult
			err := graphql.GraphQLRequest(fmt.Sprintf(`mutation {
    deleteEnvironment(
      input: {
        project:"%s"
        name:"%s"
        execute:true
      }
    )
  }`, projectName, projectEnvironment), &responseData)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				if responseData.DeleteEnvironment == "success" {
					fmt.Println(fmt.Sprintf("Result: %s", aurora.Green(responseData.DeleteEnvironment)))
				} else {
					fmt.Println(fmt.Sprintf("Result: %s", aurora.Yellow(responseData.DeleteEnvironment)))
				}
			}
		}

	},
}

func init() {
	projectDeleteCmd.AddCommand(projectDeleteEnvCmd)
}
