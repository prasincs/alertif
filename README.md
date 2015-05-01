# AlertIf: Conditional PagerDuty Alerts

This is intended to be run as a cron job or something as such for sanity checks on your server. 

![Travis CI Build Status](https://travis-ci.org/prasincs/alertif.svg)


## Supports

- disk usage check

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

## TODO

* [DONE] Adding service monitoring
* CPU check
