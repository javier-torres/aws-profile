package handlers

import (
	"flag"
	"fmt"
	"os/exec"
	"os"
		"strings"
	"gopkg.in/ini.v1"
)

type SetHandler struct {
	FlagSet *flag.FlagSet
	Flags   SetCommandFlags
}

type SetCommandFlags struct {
	CredentialsFilePath   *string
	ConfigFilePath   *string
	Pattern *string
}

func NewSetHandler() SetHandler {
	flagSet := flag.NewFlagSet("set", flag.ExitOnError)

	credentialsFilePath := flagSet.String("credentials-path", "~/.aws/credentials", "Path to AWS Credentials file")
	configFilePath := flagSet.String("config-path", "~/.aws/config", "Path to AWS Config file")
	pattern := flagSet.String("pattern", "", "Start the fzf finder with the given query")

	return SetHandler{
		FlagSet: flagSet,
		Flags:   SetCommandFlags{
			CredentialsFilePath: credentialsFilePath,
			ConfigFilePath: configFilePath,
			Pattern: pattern,
		},
	}
}

func getAWSProfiles(credentialsFile *ini.File, configFile *ini.File) []string {
	var profiles []string

	for _, section := range credentialsFile.Sections() {
		if !strings.EqualFold(section.Name(), "default") {
			profiles = append(profiles, section.Name())
		}
	}

	for _, section := range configFile.Sections() {
		if !strings.EqualFold(section.Name(), "default") &&
			section.HasKey("role_arn") &&
			section.HasKey("source_profile") {
			profiles = append(profiles, section.Name())
		}
	}

	return profiles
}

func (handler SetHandler) Handle(arguments []string) {
	flagSet := handler.FlagSet
	flagSet.Parse(arguments)
	if flagSet.Parsed() {
		credentialsFile, err := ReadFile(*handler.Flags.CredentialsFilePath)
		if err != nil {
			fmt.Printf("Fail to read AWS credentials file: %v", err)
			os.Exit(1)
		}

		configFile, err := ReadFile(*handler.Flags.ConfigFilePath)
		if err != nil {
			fmt.Printf("Fail to read AWS config file: %v", err)
			os.Exit(1)
		}

		profiles := getAWSProfiles(credentialsFile, configFile)

		joinedProfiles := strings.Join(profiles, "\n")

		fzfCommand := fmt.Sprintf("echo -e '%s' | fzf-tmux --height 30%% --reverse -1 -0 --header 'Select AWS profile' --query '%s'",
								joinedProfiles,
								*handler.Flags.Pattern)
		shellCommand := exec.Command("bash", "-c", fzfCommand)
		shellCommand.Stdin = os.Stdin
		shellCommand.Stderr = os.Stderr

		shellOutput, err := shellCommand.Output()
		if err != nil {
			// should only exit with code 0 when the error is caused by Ctrl+C
			// temporarily assume all the errors are caused by Ctrl+C for now
			os.Exit(0)
		}

		selectedProfile := string(shellOutput)
		fmt.Printf("%s", selectedProfile)
	}
}
