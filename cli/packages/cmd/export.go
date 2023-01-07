/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Infisical/infisical-merge/packages/models"
	"github.com/Infisical/infisical-merge/packages/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	FormatDotenv string = "dotenv"
	FormatJson   string = "json"
	FormatCSV    string = "csv"
	FormatYaml   string = "yaml"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:                   "export",
	Short:                 "Used to export environment variables to a file",
	DisableFlagsInUseLine: true,
	Example:               "infisical export --env=prod --format=json > secrets.json",
	Args:                  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		toggleDebug(cmd, args)
		util.RequireLogin()
		util.RequireLocalWorkspaceFile()
	},
	Run: func(cmd *cobra.Command, args []string) {
		envName, err := cmd.Flags().GetString("env")
		if err != nil {
			util.HandleError(err)
		}

		shouldExpandSecrets, err := cmd.Flags().GetBool("expand")
		if err != nil {
			util.HandleError(err)
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			util.HandleError(err)
		}

		secrets, err := util.GetAllEnvironmentVariables(envName)
		if err != nil {
			util.HandleError(err, "Unable to fetch secrets")
		}

		var output string
		if shouldExpandSecrets {
			substitutions := util.SubstituteSecrets(secrets)
			output, err = formatEnvs(substitutions, format)
			if err != nil {
				util.HandleError(err)
			}
		} else {
			output, err = formatEnvs(secrets, format)
			if err != nil {
				util.HandleError(err)
			}
		}

		fmt.Print(output)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("env", "e", "dev", "Set the environment (dev, prod, etc.) from which your secrets should be pulled from")
	exportCmd.Flags().Bool("expand", true, "Parse shell parameter expansions in your secrets")
	exportCmd.Flags().StringP("format", "f", "dotenv", "Set the format of the output file (dotenv, json, csv)")
}

// Format according to the format flag
func formatEnvs(envs []models.SingleEnvironmentVariable, format string) (string, error) {
	switch strings.ToLower(format) {
	case FormatDotenv:
		return formatAsDotEnv(envs), nil
	case FormatJson:
		return formatAsJson(envs), nil
	case FormatCSV:
		return formatAsCSV(envs), nil
	case FormatYaml:
		return formatAsYaml(envs), nil
	default:
		return "", fmt.Errorf("invalid format type: %s. Available format types are [%s]", format, []string{FormatDotenv, FormatJson, FormatCSV, FormatYaml})
	}
}

// Format environment variables as a CSV file
func formatAsCSV(envs []models.SingleEnvironmentVariable) string {
	csvString := &strings.Builder{}
	writer := csv.NewWriter(csvString)
	writer.Write([]string{"Key", "Value"})
	for _, env := range envs {
		writer.Write([]string{env.Key, env.Value})
	}
	writer.Flush()
	return csvString.String()
}

// Format environment variables as a dotenv file
func formatAsDotEnv(envs []models.SingleEnvironmentVariable) string {
	var dotenv string
	for _, env := range envs {
		dotenv += fmt.Sprintf("%s='%s'\n", env.Key, env.Value)
	}
	return dotenv
}

func formatAsYaml(envs []models.SingleEnvironmentVariable) string {
	var dotenv string
	for _, env := range envs {
		dotenv += fmt.Sprintf("%s: %s\n", env.Key, env.Value)
	}
	return dotenv
}

// Format environment variables as a JSON file
func formatAsJson(envs []models.SingleEnvironmentVariable) string {
	// Dump as a json array
	json, err := json.Marshal(envs)
	if err != nil {
		log.Errorln("Unable to marshal environment variables to JSON")
		log.Debugln(err)
		return ""
	}
	return string(json)
}
