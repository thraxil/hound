package main

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type PageResponse struct {
	GraphiteBase string
	MetricBase   string
	Alerts       []*Alert
	Idx          int
}

type AlertsCollection struct {
	Alerts  []*Alert
	Emailer Emailer
}

func NewAlertsCollection(e Emailer) *AlertsCollection {
	return &AlertsCollection{Emailer: e}
}

func (ac *AlertsCollection) AddAlert(a *Alert) {
	ac.Alerts = append(ac.Alerts, a)
}

func (ac *AlertsCollection) CheckAll() {
	for _, a := range ac.Alerts {
		a.CheckMetric()
	}
}

func (ac *AlertsCollection) ProcessAll() {
	// fetch/calculate new status for all
	ac.CheckAll()
	alerts_sent := 0
	recoveries_sent := 0
	errors := 0
	failures := 0
	successes := 0

	for _, a := range ac.Alerts {
		s, rs, e, f, as := a.UpdateState(recoveries_sent)
		successes = successes + s
		recoveries_sent = recoveries_sent + rs
		errors = errors + e
		failures = failures + f
		alerts_sent = alerts_sent + as
	}
	if alerts_sent >= GLOBAL_THROTTLE {
		ac.Emailer.Throttled(failures, GLOBAL_THROTTLE, EMAIL_TO)
	}

	if recoveries_sent >= GLOBAL_THROTTLE {
		ac.Emailer.RecoveryThrottled(recoveries_sent, GLOBAL_THROTTLE, EMAIL_TO)
	}
	ac.HandleErrors(errors)
	LogToGraphite(alerts_sent, recoveries_sent, failures, errors, successes)
}

func (ac *AlertsCollection) HandleErrors(errors int) {
	if errors > 0 {
		d := backoff_time(GLOBAL_BACKOFF)
		window := LAST_ERROR_EMAIL.Add(d)
		if time.Now().After(window) {
			ac.Emailer.EncounteredErrors(errors, EMAIL_TO)
			LAST_ERROR_EMAIL = time.Now()
			GLOBAL_BACKOFF = intmin(GLOBAL_BACKOFF+1, len(BACKOFF_DURATIONS))
		}
	} else {
		GLOBAL_BACKOFF = 0
	}
}

func LogToGraphite(alerts_sent, recoveries_sent, failures, errors, successes int) {
	var clientGraphite net.Conn
	clientGraphite, err := net.Dial("tcp", CARBON_BASE)
	if err != nil || clientGraphite == nil {
		return
	}
	defer clientGraphite.Close()
	now := int32(time.Now().Unix())
	buffer := bytes.NewBufferString("")

	fmt.Fprintf(buffer, "%salerts_sent %d %d\n", METRIC_BASE, alerts_sent, now)
	fmt.Fprintf(buffer, "%srecoveries_sent %d %d\n", METRIC_BASE, recoveries_sent, now)
	fmt.Fprintf(buffer, "%sfailures %d %d\n", METRIC_BASE, failures, now)
	fmt.Fprintf(buffer, "%serrors %d %d\n", METRIC_BASE, errors, now)
	fmt.Fprintf(buffer, "%ssuccesses %d %d\n", METRIC_BASE, successes, now)
	fmt.Fprintf(buffer, "%sglobal_backoff %d %d\n", METRIC_BASE, GLOBAL_BACKOFF, now)
	clientGraphite.Write(buffer.Bytes())
}

func intmin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (ac *AlertsCollection) Run() {
	for {
		ac.ProcessAll()
		ac.DisplayAll()
		time.Sleep(time.Duration(CHECK_INTERVAL) * time.Minute)
	}
}

func (ac *AlertsCollection) DisplayAll() {
	for _, a := range ac.Alerts {
		fmt.Println(a)
	}
}

func (ac *AlertsCollection) MakePageResponse(idx int) PageResponse {
	pr := PageResponse{GraphiteBase: GRAPHITE_BASE,
		MetricBase: METRIC_BASE, Idx: idx}
	for _, a := range ac.Alerts {
		pr.Alerts = append(pr.Alerts, a)
	}
	return pr
}
