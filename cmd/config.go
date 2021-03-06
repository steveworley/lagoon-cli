package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/amazeeio/lagoon-cli/pkg/output"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// LagoonConfigFlags .
type LagoonConfigFlags struct {
	Lagoon   string `json:"lagoon,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`
	GraphQL  string `json:"graphql,omitempty"`
	Token    string `json:"token,omitempty"`
	UI       string `json:"ui,omitempty"`
	Kibana   string `json:"kibana,omitempty"`
}

func parseLagoonConfig(flags pflag.FlagSet) LagoonConfigFlags {
	configMap := make(map[string]interface{})
	flags.VisitAll(func(f *pflag.Flag) {
		if flags.Changed(f.Name) {
			configMap[f.Name] = f.Value
		}
	})
	jsonStr, _ := json.Marshal(configMap)
	parsedFlags := LagoonConfigFlags{}
	json.Unmarshal(jsonStr, &parsedFlags)
	return parsedFlags
}

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	Short:   "Configure Lagoon CLI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

var configDefaultCmd = &cobra.Command{
	Use:     "default",
	Aliases: []string{"d"},
	Short:   "Set the default Lagoon to use",
	Run: func(cmd *cobra.Command, args []string) {
		lagoonConfig := parseLagoonConfig(*cmd.Flags())
		if lagoonConfig.Lagoon == "" {
			fmt.Println("Not enough arguments")
			cmd.Help()
			os.Exit(1)
		}
		viper.Set("default", strings.TrimSpace(string(lagoonConfig.Lagoon)))
		err := viper.WriteConfigAs(filepath.Join(configFilePath, configName+configExtension))
		handleError(err)

		resultData := output.Result{
			Result: "success",
			ResultData: map[string]interface{}{
				"default-lagoon": lagoonConfig.Lagoon,
			},
		}
		output.RenderResult(resultData, outputOptions)
	},
}

var configLagoonsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "View all configured Lagoon instances",
	Run: func(cmd *cobra.Command, args []string) {
		lagoons := viper.Get("lagoons")
		lagoonsMap := reflect.ValueOf(lagoons).MapKeys()
		if !outputOptions.CSV && !outputOptions.JSON {
			fmt.Println("You have the following Lagoon instances configured:")
			for _, lagoon := range lagoonsMap {
				fmt.Println(fmt.Sprintf("%s: %s", aurora.Yellow("Name"), lagoon))
				fmt.Println(fmt.Sprintf(" - %s: %s", aurora.Yellow("Hostname"), viper.GetString("lagoons."+lagoon.String()+".hostname")))
				fmt.Println(fmt.Sprintf(" - %s: %s", aurora.Yellow("GraphQL"), viper.GetString("lagoons."+lagoon.String()+".graphql")))
				fmt.Println(fmt.Sprintf(" - %s: %d", aurora.Yellow("Port"), viper.GetInt("lagoons."+lagoon.String()+".port")))
				fmt.Println(fmt.Sprintf(" - %s: %s", aurora.Yellow("UI"), viper.GetString("lagoons."+lagoon.String()+".ui")))
				fmt.Println(fmt.Sprintf(" - %s: %s", aurora.Yellow("Kibana"), viper.GetString("lagoons."+lagoon.String()+".kibana")))
			}
			fmt.Println("\nYour default Lagoon is:")
			fmt.Println(fmt.Sprintf("%s: %s\n", aurora.Yellow("Name"), viper.Get("default")))
			fmt.Println("Your current lagoon is:")
			fmt.Println(fmt.Sprintf("%s: %s", aurora.Yellow("Name"), viper.Get("current")))
		} else {
			var lagoonsData []map[string]interface{}
			for _, lagoon := range lagoonsMap {
				lagoonMapData := map[string]interface{}{
					"name":     fmt.Sprintf("%s", lagoon),
					"hostname": viper.GetString("lagoons." + lagoon.String() + ".hostname"),
					"graphql":  viper.GetString("lagoons." + lagoon.String() + ".graphql"),
					"port":     viper.GetString("lagoons." + lagoon.String() + ".port"),
					"ui":       viper.GetString("lagoons." + lagoon.String() + ".ui"),
					"kibana":   viper.GetString("lagoons." + lagoon.String() + ".Kibana"),
				}
				lagoonsData = append(lagoonsData, lagoonMapData)
			}
			returnedData := map[string]interface{}{
				"lagoons":        lagoonsData,
				"default-lagoon": viper.Get("default"),
				"current-lagoon": viper.Get("current"),
			}
			output.RenderJSON(returnedData, outputOptions)
		}
	},
}

var configAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add information about an additional Lagoon instance to use",
	Run: func(cmd *cobra.Command, args []string) {
		lagoonConfig := parseLagoonConfig(*cmd.Flags())
		if lagoonConfig.Lagoon == "" {
			fmt.Println("Missing arguments: Lagoon name is not defined")
			cmd.Help()
			os.Exit(1)
		}

		if lagoonConfig.Hostname != "" && lagoonConfig.Port != "" && lagoonConfig.GraphQL != "" {
			viper.Set("lagoons."+lagoonConfig.Lagoon+".hostname", lagoonConfig.Hostname)
			viper.Set("lagoons."+lagoonConfig.Lagoon+".port", lagoonConfig.Port)
			viper.Set("lagoons."+lagoonConfig.Lagoon+".graphql", lagoonConfig.GraphQL)
			if lagoonConfig.UI != "" {
				viper.Set("lagoons."+lagoonConfig.Lagoon+".ui", lagoonConfig.UI)
			}
			if lagoonConfig.Kibana != "" {
				viper.Set("lagoons."+lagoonConfig.Lagoon+".kibana", lagoonConfig.Kibana)
			}
			if lagoonConfig.Token != "" {
				viper.Set("lagoons."+lagoonConfig.Lagoon+".token", lagoonConfig.Token)
			}
			err := viper.WriteConfigAs(filepath.Join(configFilePath, configName+configExtension))
			if err != nil {
				output.RenderError(err.Error(), outputOptions)
				os.Exit(1)
			}
			resultData := output.Result{
				Result: "success",
				ResultData: map[string]interface{}{
					"lagoon":   lagoonConfig.Lagoon,
					"hostname": lagoonConfig.Hostname,
					"graphql":  lagoonConfig.GraphQL,
					"port":     lagoonConfig.Port,
					"ui":       lagoonConfig.UI,
					"kibana":   lagoonConfig.Kibana,
				},
			}
			output.RenderResult(resultData, outputOptions)
		} else {
			output.RenderError("Must have Hostname, Port, and GraphQL endpoint", outputOptions)
		}
	},
}

var configDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete a Lagoon instance configuration",
	Run: func(cmd *cobra.Command, args []string) {
		lagoonConfig := parseLagoonConfig(*cmd.Flags())

		if lagoonConfig.Lagoon == "" {
			fmt.Println("Missing arguments: Lagoon name is not defined")
			cmd.Help()
			os.Exit(1)
		}
		if yesNo(fmt.Sprintf("You are attempting to delete config for lagoon '%s', are you sure?", lagoonConfig.Lagoon)) {
			err := unset(lagoonConfig.Lagoon)
			if err != nil {
				output.RenderError(err.Error(), outputOptions)
				os.Exit(1)
			}
		}
	},
}

var configFeatureSwitch = &cobra.Command{
	Use:     "feature",
	Aliases: []string{"f"},
	Short:   "Enable or disable CLI features",
	Run: func(cmd *cobra.Command, args []string) {
		switch updateCheck {
		case "true":
			viper.Set("updateCheckDisable", true)
		case "false":
			viper.Set("updateCheckDisable", false)
		}
		switch projectDirectoryCheck {
		case "true":
			viper.Set("projectDirectoryCheckDisable", true)
		case "false":
			viper.Set("projectDirectoryCheckDisable", false)
		}
		err := viper.WriteConfigAs(filepath.Join(configFilePath, configName+configExtension))
		if err != nil {
			output.RenderError(err.Error(), outputOptions)
			os.Exit(1)
		}
	},
}

var configGetCurrent = &cobra.Command{
	Use:     "current",
	Aliases: []string{"cur"},
	Short:   "Display the current lagoon that commands would be executed against",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.GetString("current"))
	},
}

var updateCheck string
var projectDirectoryCheck string

func init() {
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configGetCurrent)
	configCmd.AddCommand(configDefaultCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configFeatureSwitch)
	configCmd.AddCommand(configLagoonsCmd)
	configAddCmd.Flags().StringVarP(&lagoonHostname, "hostname", "H", "", "Lagoon SSH hostname")
	configAddCmd.Flags().StringVarP(&lagoonPort, "port", "P", "", "Lagoon SSH port")
	configAddCmd.Flags().StringVarP(&lagoonGraphQL, "graphql", "g", "", "Lagoon GraphQL endpoint")
	configAddCmd.Flags().StringVarP(&lagoonToken, "token", "t", "", "Lagoon GraphQL token")
	configAddCmd.Flags().StringVarP(&lagoonUI, "ui", "u", "", "Lagoon UI location (https://ui-lagoon-master.ch.amazee.io)")
	configAddCmd.PersistentFlags().BoolVarP(&createConfig, "create-config", "", false, "Create the config file if it is non existent (to be used with --config-file)")
	configAddCmd.Flags().StringVarP(&lagoonKibana, "kibana", "k", "", "Lagoon Kibana URL (https://logs-db-ui-lagoon-master.ch.amazee.io)")
	configFeatureSwitch.Flags().StringVarP(&updateCheck, "disable-update-check", "", "", "Enable or disable checking of updates (true/false)")
	configFeatureSwitch.Flags().StringVarP(&projectDirectoryCheck, "disable-project-directory-check", "", "", "Enable or disable checking of local directory for lagoon project (true/false)")
}
