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
		altsrc.NewInt64SliceFlag(&cli.Int64SliceFlag{
			Name:  "notify-exit-code",
			Value: cli.NewInt64Slice(),
			Usage: "Exit status codes to notify to Slack",
		}),
		altsrc.NewInt64SliceFlag(&cli.Int64SliceFlag{
			Name:  "ignore-exit-code",
			Value: cli.NewInt64Slice(0),
			Usage: "Exit status codes not to notify to Slack",
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

	// Run a command

	args := c.Args().Slice()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	var exitCode int64
	exitCode = -1

	if err := cmd.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			exitCode = int64(eerr.ExitCode())
		}
	} else {
		exitCode = 0
	}

	for _, i := range c.Int64Slice("ignore-exit-code") {
		if i == exitCode {
			return nil
		}
	}
	if codes := c.Int64Slice("notify-exit-code"); len(codes) > 0 {
		notify := false
		for _, i := range codes {
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
