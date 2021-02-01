package config

import "time"

const (
	EnvName = "demo.env"

	MachineRawData = "iii.dae.MachineRawData"
	Statistic      = "iii.dae.Statistics"
	EventLatest    = "iii.dae.EventLatest"
	EventHist      = "iii.dae.EventHist"
	GroupTopo      = "iii.cfg.GroupTopology"
	TPCList        = "iii.cfg.TPCList"
)

var (
	MongodbURL      string
	MongodbDatabase string
	MongodbUsername string
	MongodbPassword string

	TaipeiTimeZone *time.Location
	UTCTimeZone    *time.Location
)
