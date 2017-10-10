package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"time"
)

type Marker struct {
	Name        string
	Type        int
	Description string
}

const (
	MARKER_TYPE_START = 0
	MARKER_TYPE_END   = 1
	MARKER_TYPE_ERR   = 2
)

const (
	ExitCodeOK = iota
	ExitCodeErr
)

var markers map[string]Marker = map[string]Marker{
	"dodeploy":      {"DoDeploy", MARKER_TYPE_START, "doing deploy"},
	"skipdeploy":    {"SkipDeploy", MARKER_TYPE_END, "disable auto-deploy"},
	"isdeploying":   {"IsDeploying", MARKER_TYPE_START, "deploying"},
	"deployed":      {"Deployed", MARKER_TYPE_START, "deployed"},
	"failed":        {"Failed", MARKER_TYPE_ERR, "deploy failed"},
	"isundeploying": {"IsUnDeploying", MARKER_TYPE_END, "disabling deploy"},
	"undeployed":    {"UnDeployed", MARKER_TYPE_END, "disable deploy"},
	"pending":       {"Pending", MARKER_TYPE_END, "Pending deploy"},
}

func GetWildflyState(war_path string) (res []Marker, err error) {
	basedir := filepath.Dir(war_path)

	files, err := ioutil.ReadDir(basedir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		arr := strings.Split(file.Name(), ".")
		ext := arr[len(arr)-1]

		if v, exists := markers[ext]; exists {
			res = append(res, v)
		}
	}
	return
}

func sendNotification(api_url, channel, msg string, marker_type int) error {
	hostname, _ := os.Hostname()
	fields := []AttachmentField{
		SetAttachmentField("HostName", hostname),
		SetAttachmentField("Message", msg),
	}

	var color string
	switch marker_type {
	case MARKER_TYPE_START:
		color = SLACK_ATTACHMENT_START
	case MARKER_TYPE_END:
		color = SLACK_ATTACHMENT_END
	case MARKER_TYPE_ERR:
		color = SLACK_ATTACHMENT_ERR
	}

	attachments := []Attachment{
		SetAttachment(msg, color, fields),
	}
	slack := setSlack(channel, attachments)
	err := SlackNotification(api_url, slack)

	return err
}

func monitorState(config Config, notify_markers map[string]bool) int {
	var now_state []Marker
	first := true

	signal_ch := make(chan os.Signal, 1)
	signal.Notify(signal_ch, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-signal_ch:
			break loop
		default:
			time.Sleep(config.App.Duration * time.Second)

			res, err := GetWildflyState(config.Wildfly.WarPath)
			if err != nil {
				log.Error(err)
				return ExitCodeErr
			}
			if first {
				now_state = res
				first = false
			}
			if reflect.DeepEqual(now_state, res) {
				continue
			}

			log.Info("Change State")
			for _, v := range res {
				log.WithFields(log.Fields{
					"name":        v.Name,
					"description": v.Description,
				}).Info("Change State")

				if len(notify_markers) == 0 {
					err := sendNotification(config.Slack.ApiUrl, config.Slack.Channel, v.Description, v.Type)
					if err != nil {
						log.Error(err)
					}
					log.Info("Send notification to slack.")
				} else {
					if _, exists := notify_markers[v.Name]; exists {
						err := sendNotification(config.Slack.ApiUrl, config.Slack.Channel, v.Description, v.Type)
						if err != nil {
							log.Error(err)
						}
						log.Info("Send notification to slack.")
					}
				}
			}
			now_state = res
		}
	}
	return ExitCodeOK
}

func settingLog(log_path string) error {
	log.SetFormatter(&log.JSONFormatter{})
	if log_path == "" {
		log.SetOutput(os.Stdout)
	} else {
		logfile, err := os.OpenFile(log_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		log.SetOutput(logfile)
	}
	return nil
}

func Run(args []string) int {
	var config_path string

	flags := flag.NewFlagSet("wfsm", flag.ContinueOnError)
	flags.StringVar(&config_path, "config", "config.yaml", "Specify config file path")

	if err := flags.Parse(args[1:]); err != nil {
		flags.PrintDefaults()
		return ExitCodeErr
	}
	config, err := loadConfig(config_path)
	if err != nil {
		log.Error(err)
		return ExitCodeErr
	}

	if err := settingLog(config.App.LogPath); err != nil {
		log.Error("Failed open log file. " + config.App.LogPath)
		return ExitCodeErr
	}

	var notify_markers = make(map[string]bool)

	for _, v := range config.App.NotifyMarker {
		if _, exists := markers[v]; !exists {
			log.Error("Error specify marker :" + v)
			return 1
		}
		notify_markers[markers[v].Name] = true
	}

	log.Info("Start Monitoring...")
	exit_code := monitorState(config, notify_markers)
	log.Info("End Monitoring")

	return exit_code
}

func main() {
	exit_code := Run(os.Args)
	os.Exit(exit_code)
}
