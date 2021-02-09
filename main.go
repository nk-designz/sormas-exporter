package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	contactsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_contacts",
		Help: "The contacts in sormas database",
	})
	casesInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_cases",
		Help: "The cases in sormas database",
	})
	eventsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_events",
		Help: "The events in sormas database",
	})
	usersInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_users",
		Help: "The users in sormas database",
	})
	personsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_persons",
		Help: "The persons in sormas database",
	})
	sessionsInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_sessions",
		Help: "The sessions in sormas database",
	})
	tasksInSormas = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sormas_exporter_tasks",
		Help: "The tasks in sormas database",
	})

	hostFlag = flag.String("host", "localhost", "database host")
	userFlag = flag.String("user", "sormas_user", "database user")
	passwordFlag = flag.String("password", "password", "database password")
	portFlag = flag.Int("port", 5432, "database port")
	retrysFlag = flag.Int("retry", 5, "seconds after refresh")
)

func main() {
	flag.Parse()
	New(*hostFlag, *userFlag, *passwordFlag, *portFlag, *retrysFlag).Run()
}

// Exporter type
type Exporter struct {
	db *gorm.DB
	retrys int
}

// New exporter
func New(host, user, password string, port, retrys int) *Exporter {
	var err error
	e := new(Exporter)
	e.retrys = retrys
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
	return e
}

// Run exporter
func (e *Exporter) Run() {
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
		time.Sleep(time.Duration(e.retrys) * time.Second)
	}
}
