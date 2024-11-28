package config

import "fmt"

type EventStorageType string

var (
	EventStorageTypeMemory EventStorageType = "memory"
	EventStorageTypePg     EventStorageType = "pg"
)

func (t *EventStorageType) UnmarshalText(s []byte) error {
	switch string(s) {
	case string(EventStorageTypeMemory):
		*t = EventStorageTypeMemory
	case string(EventStorageTypePg):
		*t = EventStorageTypePg
	default:
		return fmt.Errorf("invalid event storage type '%s'", s)
	}

	return nil
}

func (t *EventStorageType) String() string {
	return string(*t)
}

type EventStoragePg struct {
	DataSource string `yaml:"data_source" env:"DATASOURCE"`
}
