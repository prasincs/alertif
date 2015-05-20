#!/bin/bash
PAGERDUTY_KEY="ENTER_KEY"
/usr/local/bin/alertif -s $PAGERDUTY_KEY --disk -t 70 -i "/dev"
