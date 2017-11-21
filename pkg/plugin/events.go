package plugin

import (
	"github.com/hellofresh/janus/pkg/router"
)

// Define the event names for the startup and shutdown events
const (
	StartupEvent         string = "startup"
	AdminAPIStartupEvent string = "admin_startup"

	ShutdownEvent string = "shutdown"
)

// OnAdminAPIStartup represents a event that happens when Janus starts up the admin API
type OnAdminAPIStartup struct {
	Router router.Router
}
