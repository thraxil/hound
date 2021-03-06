#!/bin/sh

set -e

make

echo "Running hound..."

HOUND_GRAPHITE_BASE="https://graphite.ctl.columbia.edu/render/" \
HOUND_CARBON_BASE="graphite.ctl.columbia.edu:2003" \
HOUND_METRIC_BASE="app.gauges.hounddev." \
HOUND_EMAIL_FROM="hound@thraxil.org" \
HOUND_EMAIL_TO="anders@columbia.edu" \
HOUND_CHECK_INTERVAL=1 \
HOUND_GLOBAL_THROTTLE=10 \
HOUND_HTTP_PORT=9998 \
HOUND_TEMPLATE_FILE="index.html" \
HOUND_ALERT_TEMPLATE_FILE="alert.html" \
HOUND_EMAIL_ON_ERROR=false \
HOUND_SMTP_SERVER=postgres \
HOUND_SMTP_PORT=25 \
HOUND_LOG_LEVEL=DEBUG \
HOUND_WINDOW=2hours \
\
./hound -config=config.json
