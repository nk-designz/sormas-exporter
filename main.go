package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	contactsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_contacts",
		Help: "The contacts in sormas database.",
	})
	casesInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_cases",
		Help: "The cases in sormas database.",
	})
	eventsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_events",
		Help: "The events in sormas database.",
	})
	usersInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_users",
		Help: "The users in sormas database.",
	})
	personsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_persons",
		Help: "The persons in sormas database.",
	})
	sessionsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_sessions",
		Help: "The sessions in sormas database.",
	})
	tasksInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_usage_tasks",
		Help: "The tasks in sormas database.",
	})

	hostFlag     = flag.String("host", "localhost", "database host")
	userFlag     = flag.String("user", "sormas_user", "database user")
	passwordFlag = flag.String("password", "password", "database password")
	portFlag     = flag.Int("port", 5432, "database port")
	delayFlag    = flag.Int("delay", 30, "seconds between gathering")
	pathFlag     = flag.String("path", "/var/lib/node-exporter", "Where to store metrics in file system")
)

func main() {
	flag.Parse()
	New(*hostFlag, *userFlag, *passwordFlag, *portFlag, *delayFlag, *pathFlag).Run()
}

// Exporter type
type Exporter struct {
	db    *gorm.DB
	delay int
	reg   *prometheus.Registry
	path  string
}

// New exporter
func New(host, user, password string, port, delay int, path string) *Exporter {
	var err error
	e := new(Exporter)
	e.delay = delay
	e.path = path
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=sormas port=%d sslmode=disable TimeZone=Europe/Berlin",
		host,
		user,
		password,
		port,
	)
	e.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// Let's create a custom registry and then let it collect the metrics.
	// see https://github.com/prometheus/client_golang/blob/master/prometheus/examples_test.go
	// from line:345
	e.reg = prometheus.NewRegistry()

	return e
}

// Run exporter
func (e *Exporter) Run() {
	e.reg.MustRegister(contactsInSormas)
	e.reg.MustRegister(casesInSormas)
	e.reg.MustRegister(eventsInSormas)
	e.reg.MustRegister(usersInSormas)
	e.reg.MustRegister(personsInSormas)
	e.reg.MustRegister(sessionsInSormas)
	e.reg.MustRegister(tasksInSormas)
	go e.checkRoutine()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:3014", nil))
}

func (e *Exporter) checkRoutine() {
	for {
		var (
			contacts int64
			events   int64
			cases    int64
			users    int64
			persons  int64
			sessions int64
			tasks    int64
		)
		e.db.Table("contact").Count(&contacts)
		e.db.Table("events").Count(&events)
		e.db.Table("cases").Count(&cases)
		e.db.Table("users").Count(&users)
		e.db.Table("person").Count(&persons)
		e.db.Table("task").Count(&tasks)
		e.db.Table("pg_stat_activity").Count(&sessions)

		contactsInSormas.Set(float64(contacts))
		eventsInSormas.Set(float64(events))
		casesInSormas.Set(float64(cases))
		usersInSormas.Set(float64(users))
		personsInSormas.Set(float64(persons))
		sessionsInSormas.Set(float64(sessions))
		tasksInSormas.Set(float64(tasks))

		e.writeMetrics()

		log.Printf(
			"Contacts: %d, Events: %d, Cases: %d, User: %d, Persons: %d, Tasks: %d, Sessions: %d\n",
			contacts,
			events,
			cases,
			users,
			persons,
			tasks,
			sessions,
		)
		time.Sleep(time.Duration(e.delay) * time.Second)
	}
}

func (e *Exporter) writeMetrics() {
	filename := e.path + "/sormas-usage.prom.temp"
	f, err := os.Create(filename)
	if err != nil {
		panic("error creating file " + filename)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	metrics, err := e.reg.Gather()
	if err != nil || len(metrics) == 0 {
		panic("unexpected behavior of custom test registry")
	}
	for _, m := range metrics {
		w.WriteString(fmt.Sprintf("# HELP %s %s\n", *m.Name, *m.Help))
		w.WriteString(fmt.Sprintf("# TYPE %s %s\n", *m.Name, strings.ToLower(m.Type.String())))
		w.WriteString(fmt.Sprintf("%s %d\n", *m.Name, int(m.GetMetric()[0].GetGauge().GetValue())))
		w.WriteString("\n")
	}
	w.Flush()
	f.Close()
	os.Rename(filename, e.path+"/sormas-usage.prom")
}
