# AlertIf: Conditional PagerDuty Alerts

This is intended to be run as a cron job or something as such for sanity checks on your server. 

[![Build Status](https://travis-ci.org/prasincs/alertif.svg?branch=master)](https://travis-ci.org/prasincs/alertif)


## Supports

- disk usage check
- tcp connection check
- http connection check

## Setup

```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GOROOT=/usr/local/go      # assuming go is installed at /usr/local/go
export PATH=$PATH:$GOROOT/bin
```

## Building using godep

If you haven't gotten godep yet, do that using 

`go get github.com/tools/godep`

```
go get github.com/prasincs/alertif
cd $GOPATH/src/github.com/prasincs/alertif
make all
```


## Usage

You can run the command using 

`./alertif -p <pagerduty service key> --disk -t 80 -i "/dev,/tmp" -s example,tcp,8888,dead`

This means 

* `-p` Run with a pagerduty service key that you've obtained from the site.
* `--disk` enable checking disk usage
* `-t,--disk-threshold` If there's any disk that's using more than 80 percent disk usage, send PagerDuty Alert
* `-i,--disk-ignore` Ignores "/dev" and "/tmp" mountpoints
* `-s,--service` If the service is dead, sends alert. Note the syntax, it goes as follows: Name,Type,Port,Action
* `-h, --hostname` The Hostname you want to use for service checks.

### Service Check format

A Service Command is a comma separated list. It is broken down into the following components:

"Name,Type,Port,Action"

The resulting alert would be sent under the name you've given the service. For `http` service type, the Action is the URL you want to GET.

## TODO

* [DONE] Adding service monitoring
* Add end to end tests
