package main

import (
	"deplagene/avito-tech-internship/utils"
	"log/slog"
)

func main() {
	logger := utils.NewLogger("development")
	logger.Info("Starting application")
	slog.Info("This is a default logger message, for comparison.")
	// TODO: Передать logger в другие слои проекта.
}
