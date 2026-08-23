package common

// Project metadata.
const (
	COMMON_PROJECT_NAME = "gofurry-admin"
	COMMON_PROJECT_HELP = `
gofurry Admin is the maintenance console for gofurry public services.
Usage:
  ./gofurry-admin [command]
    install: install this service into systemd.
    uninstall: uninstall this service from systemd.
    reset-password: reset the single admin password.
    version: show the current version.
    help: show this help message.
`
)

// Time formats.
const (
	TIME_FORMAT_DIGIT_DAY = "20060102"
	TIME_FORMAT_DIGIT     = "20060102150405"
	TIME_FORMAT_DATE      = "2006-01-02 15:04:05"
	TIME_FORMAT_DAY       = "2006-01-02"
	TIME_FORMAT_LOG       = "2006-01-02 15:04:05.000"
)

// Response status flags.
const (
	RETURN_FAILED       = 0
	RETURN_SUCCESS      = 1
	RETURN_SUCCESS_CODE = 200
	RETURN_FAILED_CODE  = 500
)

// Common numeric constants.
const (
	JWT_RELET_NUM     = 2
	ONLINE_RELET_NUM  = 5
	EMAIL_CODE_LENGTH = 6
)

// Common HTTP headers.
const (
	USER_AGENT      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	ACCEPT_LANGUAGE = "en-US,en;q=0.9"
	APPLICATION     = "application/json"
)

// Event names.
const (
	GLOBAL_MSG = "GLOBAL_MSG"
	COMMON_MSG = "COMMON_MSG"
)
