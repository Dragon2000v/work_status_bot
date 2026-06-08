package database

import (
	"log/slog"
)

func LogMongoConnect(log *slog.Logger, databaseName, outcome string, err error) {
	if log == nil {
		log = slog.Default()
	}
	attrs := []any{
		"event", "mongodb.connect",
		"operation", "mongodb",
		"database", databaseName,
		"outcome", outcome,
	}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
		log.Error("mongodb connect", attrs...)
		return
	}
	log.Info("mongodb connect", attrs...)
}

func LogMongoIndexes(log *slog.Logger, databaseName, outcome string, err error) {
	if log == nil {
		log = slog.Default()
	}
	attrs := []any{
		"event", "mongodb.indexes",
		"operation", "mongodb",
		"database", databaseName,
		"outcome", outcome,
	}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
		log.Error("mongodb indexes", attrs...)
		return
	}
	log.Info("mongodb indexes", attrs...)
}
