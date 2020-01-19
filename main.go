package main

import (
	"bytes"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

var version string
const defaultConfigPath = "/etc/wrapslack.yaml"

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func start() error {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "version",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "slack-token",
			EnvVars: []string{"SLACK_TOKEN"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "slack-channel",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "slack-message-template",
			Value: "`{{.hostName}}`: `{{.command}}` exited with `{{.exitCode}}`",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "slack-icon-emoji",
			Value: ":robot_face:",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "slack-username",
			Value: "wrapslack",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "notify-exit-codes",
			Value: "",
			Usage: "Comma-separated exit status codes to notify to Slack (empty '' for all status)",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "ignore-exit-codes",
			Value: "0",
			Usage: "Comma-separated exit status codes not to notify to Slack",
		}),
	}

	app := &cli.App{
		Name:   "wrapslack",
		Action: action,
		Flags:  flags,
	}

	if _, err := os.Stat(defaultConfigPath); err == nil {
		app.Before = altsrc.InitInputSource(flags, func() (altsrc.InputSourceContext, error) {
			return altsrc.NewYamlSourceFromFile(defaultConfigPath)
		})
	}

	return app.Run(os.Args)
}

func action(c *cli.Context) error {
	if c.Bool("version") {
		fmt.Printf("wrapslack %s\n", version)
		return nil
	}

	slackToken := c.String("slack-token")
	if slackToken == "" {
		return fmt.Errorf("slack-token is required")
	}

	slackChannel := c.String("slack-channel")
	if slackChannel == "" {
		return fmt.Errorf("slack-channel is required")
	}

	if c.NArg() == 0 {
		return fmt.Errorf("command is required")
	}

	slackMessageTemplate, err := template.New("").Parse(c.String("slack-message-template"))
	if err != nil {
		return err
	}

	ignoreCodes, err := parseCommaSeparatedToInts(c.String("ignore-exit-codes"))
	if err != nil {
		return err
	}

	notifyCodes, err := parseCommaSeparatedToInts(c.String("notify-exit-codes"))
	if err != nil {
		return err
	}

	// Run a command

	args := c.Args().Slice()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	var exitCode int
	exitCode = -1

	if err := cmd.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			exitCode = eerr.ExitCode()
		}
	} else {
		exitCode = 0
	}

	for _, i := range ignoreCodes {
		if i == exitCode {
			return nil
		}
	}

	if len(notifyCodes) > 0 {
		notify := false
		for _, i := range notifyCodes {
			if i == exitCode {
				notify = true
				break
			}
		}
		if !notify {
			return nil
		}
	}

	// Build message

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	templateData := map[string]interface{}{
		"hostName": hostname,
		"command":  args,
		"exitCode": exitCode,
	}

	var buf bytes.Buffer
	if err := slackMessageTemplate.Execute(&buf, templateData); err != nil {
		return err
	}

	// Notify

	slackApi := slack.New(slackToken)
	_, _, err = slackApi.PostMessage(slackChannel,
		slack.MsgOptionText(buf.String(), false),
		slack.MsgOptionIconEmoji(c.String("slack-icon-emoji")),
		slack.MsgOptionUsername(c.String("slack-username")))
	if err != nil {
		return err
	}

	return nil
}

func parseCommaSeparatedToInts(str string) ([]int, error) {
	var ints []int
	if str == "" {
		return ints, nil
	}

	for _, s := range strings.Split(str, ",") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		ints = append(ints, i)
	}
	return ints, nil
}
