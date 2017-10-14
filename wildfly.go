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
	"fmt"
)

// Marker -- wildfly marker
type Marker struct {
	Name        string
	Type        int
	Description string
}

const (
	// MarkerTypeStart -- Means wildfly start
	MarkerTypeStart = 0
	// MarkerTypeEnd -- Means wildfly end
	MarkerTypeEnd   = 1
	// MarkerTypeErr -- Means wildfly error
	MarkerTypeErr   = 2
)

const (
	// ExitCodeOK -- Success code
	ExitCodeOK = iota
	// ExitCodeErr -- Error code
	ExitCodeErr
)

var markerList = map[string]Marker{
	"dodeploy":      {"DoDeploy", MarkerTypeStart, "doing deploy"},
	"skipdeploy":    {"SkipDeploy", MarkerTypeEnd, "disable auto-deploy"},
	"isdeploying":   {"IsDeploying", MarkerTypeStart, "deploying"},
	"deployed":      {"Deployed", MarkerTypeStart, "deployed"},
	"failed":        {"Failed", MarkerTypeErr, "deploy failed"},
	"isundeploying": {"IsUnDeploying", MarkerTypeEnd, "disabling deploy"},
	"undeployed":    {"UnDeployed", MarkerTypeEnd, "disable deploy"},
	"pending":       {"Pending", MarkerTypeEnd, "Pending deploy"},
}

func getWildflyState(warPath string) (res []Marker, err error) {
	basedir := filepath.Dir(warPath)

	files, err := ioutil.ReadDir(basedir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		arr := strings.Split(file.Name(), ".")
		ext := arr[len(arr)-1]

		if v, exists := markerList[ext]; exists {
			res = append(res, v)
		}
	}
	return
}

func sendNotification(apiURL, channel, msg string, markerType int) error {
	hostname, _ := os.Hostname()
	fields := []AttachmentField{
		SetAttachmentField("HostName", hostname),
		SetAttachmentField("Message", msg),
	}

	var color string
	switch markerType {
	case MarkerTypeStart:
		color = SlackAttachementStart
	case MarkerTypeEnd:
		color = SlackAttachementEnd
	case MarkerTypeErr:
		color = SlackAttachementErr
	}

	attachments := []Attachment{
		SetAttachment(msg, color, fields),
	}
	slack := SetSlack(channel, attachments)
	err := SlackNotification(apiURL, slack)

	return err
}

func monitorState(config Config, notifyMarkers map[string]bool) int {
	var nowState []Marker
	first := true

	signalch := make(chan os.Signal, 1)
	signal.Notify(signalch, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-signalch:
			break loop
		default:
			time.Sleep(config.App.Duration * time.Second)

			res, err := getWildflyState(config.Wildfly.WarPath)
			if err != nil {
				log.Error(err)
				return ExitCodeErr
			}
			if first {
				nowState = res
				first = false
			}
			if reflect.DeepEqual(nowState, res) {
				continue
			}

			log.Info("Change State")
			for _, v := range res {
				log.WithFields(log.Fields{
					"name":        v.Name,
					"description": v.Description,
				}).Info("Change State")

				if len(notifyMarkers) == 0 {
					err := sendNotification(config.Slack.APIURL, config.Slack.Channel, v.Description, v.Type)
					if err != nil {
						log.Error(err)
					}
					log.Info("Send notification to slack.")
				} else {
					if _, exists := notifyMarkers[v.Name]; exists {
						err := sendNotification(config.Slack.APIURL, config.Slack.Channel, v.Description, v.Type)
						if err != nil {
							log.Error(err)
						}
						log.Info("Send notification to slack.")
					}
				}
			}
			nowState = res
		}
	}
	return ExitCodeOK
}

func settingLog(logPath string) error {
	log.SetFormatter(&log.JSONFormatter{})
	if logPath == "" {
		log.SetOutput(os.Stdout)
	} else {
		logfile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		log.SetOutput(logfile)
	}
	return nil
}

func run(args []string) int {
	var configPath string
	var version bool

	flags := flag.NewFlagSet("wildfly-state-monitor", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&configPath, "config", "config.yaml", "Specify config file path")
	flags.StringVar(&configPath, "c", "config.yaml", "Specify config file path")
	flags.BoolVar(&version, "v", false, "Output version number.")
	flags.BoolVar(&version, "version", false, "Output version number.")

	if err := flags.Parse(args[1:]); err != nil {
		flags.PrintDefaults()
		return ExitCodeErr
	}

	if version {
		fmt.Println(Version)
		return ExitCodeOK
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Error(err)
		return ExitCodeErr
	}

	if err := settingLog(config.App.LogPath); err != nil {
		log.Error("Failed open log file. " + config.App.LogPath)
		return ExitCodeErr
	}

	var notifyMarkers = make(map[string]bool)

	for _, v := range config.App.NotifyMarker {
		if _, exists := markerList[v]; !exists {
			log.Error("Error specify marker :" + v)
			return ExitCodeErr
		}
		notifyMarkers[markerList[v].Name] = true
	}

	log.Info("Start Monitoring...")
	exitCode := monitorState(config, notifyMarkers)
	log.Info("End Monitoring")

	return exitCode
}

func main() {
	exitCode := run(os.Args)
	os.Exit(exitCode)
}
