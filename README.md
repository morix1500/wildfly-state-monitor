# wildfly-state-monitor
This app is monitor to [Wildfly Marker](https://docs.jboss.org/author/display/WFLY10/Application+deployment).  
if changed marker, send message to slack.  

## Description
Wildfly is application server.  
When Wildfly's application is placed in the deploy directory,  
it generates a marker representing the deployed state.  
<https://docs.jboss.org/author/display/WFLY10/Application+deployment>  

wildfly-state-monitor will monitor that marker. And when it is changed send a message to Slack.  
The marker to be monitored can be specified in the setting file.  

With this application, it is possible to detect start/stop, error of Wildfly application.

## Usage
```
$ ./wildfly-state-monitor -h
Usage of wildfly-state-monitor:
  -c string
        Specify config file path (default "config.yaml")
  -config string
        Specify config file path (default "config.yaml")
  -v    Output version number.
  -version
        Output version number.
```

### coinfig
Support format is yaml only.  
Items in the configuration file are listed in [config.yaml.sample](./config.yaml.sample).  
If you want to setting config file, following execute command.

```
$ cp config.yaml.sample config.yaml
```

| Name | Description |
| --- | --- |
| slack.api_url | Slack Incomming Webhook URL |
| slack.channel | Target Slack channel |
| wildfly.war_path | Full path of Wildfly application war file |
| app.log_path | Where you want to output logs. Default: standard out |
| app.duration | Monitoring interval. Unit: seconds |
| app.notify_markers | Specify Wildfly markers to receive notification |


## Installation

```
$ wget https://github.com/morix1500/wildfly-state-monitor/releases/download/v1.0.0/wildfly-state-monitor_linux_amd64 -O /usr/local/bin/wildfly-state-monitor
$ chmod u+x /usr/local/bin/wildfly-state-monitor
```

## License
Please see the [LICENSE](./LICENSE) file for details.  

## Author
Shota Omori(Morix)  
https://github.com/morix1500
