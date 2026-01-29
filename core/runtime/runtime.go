package runtime

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/google/uuid"
)

const (
	// website is a template used to construct the service's project URL.
	website = "https://github.com/%s/%s"
)

// APP represents the global identity and deployment context of the application.
type APP struct {
	// App is the project/owner name.
	App string `json:"app"`
	// Environment is the deployment stage (e.g., prod, staging, dev).
	Environment string `json:"environment"`
	// Region is the physical geographic area (e.g., us-east-1).
	Region string `json:"region"`
	// AZ (Available Zone) is the specific data center isolation zone.
	AZ string `json:"avaliablezone"`
	// Website is the auto-generated project URL.
	Website string `json:"website"`
	// Favicon for UI-related service dashboards.
	Favicon string `json:"favicon"`
	// Desc is a short description of the service's purpose.
	Desc string `json:"desc"`
	// Instance contains the unique details of this specific running process.
	Instance Instance `json:"instance"`
}

// Instance describes a specific running member of a service group.
type Instance struct {
	// ID is a unique UUID generated at startup to identify this specific container/process.
	ID string
	// SystemCode is a numeric identifier (100-999) used for internal error codes or routing.
	SystemCode uint32 `json:"systemCode"`
	// Name is the specific deployment or entrypoint name (e.g., "svc-example-api", "svc-example-openapi").
	// It is used for fine-grained routing and service discovery.
	Name string `json:"name"`
	// Group identifies the logical service entity (e.g., "svc-example").
	// Multiple entrypoints (Names) can share the same Group to indicate they
	// belong to the same business logic and code base.
	Group string `json:"group"`
	// Version follows semantic versioning (e.g., "1.2.3").
	Version string `json:"version"`
	// Shareable indicates if this service can be used across different organizational units.
	Shareable bool `json:"shareable"`
	// MetaData stores arbitrary key-value pairs for flexible service discovery filtering.
	MetaData map[string]string `json:"metadata"`
}

var (
	// app holds the singleton instance of the runtime context.
	app = APP{
		App:         constant.Framework,
		Environment: "dev",
		Region:      "default",
		AZ:          "default",
		Favicon:     "favicon.ico",
		Instance: Instance{
			SystemCode: 100,
			Name:       constant.Framework,
			Version:    "1.0.0",
			MetaData:   make(map[string]string),
		},
	}
	// Exit is a global channel used to signal all background goroutines to stop.
	Exit    = make(chan struct{})
	appOnce sync.Once
)

// GetAPP retrieves the singleton application context.
// It panics if called before the configuration system is ready, ensuring that
// metadata is correctly loaded from config files or environment variables.
func GetAPP() APP {
	if !config.IsLoaded() {
		panic("config not loaded")
	}
	appOnce.Do(func() {
		// Unmarshal settings from the config source (e.g., asjard.service prefix).
		if err := config.GetWithUnmarshal(constant.ConfigServicePrefix, &app); err != nil {
			logger.Error("get instance conf fail", "err", err)
		}
		// Validate SystemCode range.
		if app.Instance.SystemCode < 100 || app.Instance.SystemCode > 999 {
			app.Instance.SystemCode = 100
		}
		// Generate unique ID and link project website.
		app.Instance.ID = uuid.NewString()
		app.Website = fmt.Sprintf(website, app.App, app.Instance.Name)

		// If Group is not specified, the service is treated as a standalone entity
		// where the entrypoint (Name) and the logical group (Group) are identical.
		if app.Instance.Group == "" {
			app.Instance.Group = app.Instance.Name
		}

		// Synchronize metadata into the constant package for easy access in other modules.
		constant.APP.Store(app.App)
		constant.Region.Store(app.Region)
		constant.AZ.Store(app.AZ)
		constant.Env.Store(app.Environment)
		constant.ServiceName.Store(app.Instance.Name)

		logger.Debug("get app", "app", app)
	})
	return app
}

// bufPool optimizes memory allocation by reusing byte buffers for key generation.
var bufPool = sync.Pool{
	New: func() any {
		b := bytes.NewBuffer(make([]byte, 0, 128))
		return b
	},
}

// ResourceKey generates a structured, delimited string useful for resource isolation.
// Format: {app}/{resource}/{env}/{service-group}/{version}/{region}/{az}/{key}
//
// Examples:
// - Redis Key: "asjard/caches/prod/svc-user/v1/us-east/az1/user:123"
// - Distributed Lock: "asjard/lock/staging/order-svc/v2/orders:lock"
func (app APP) ResourceKey(resource, key string, opts ...Option) string {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	if resource == "" {
		resource = "resource"
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Internal helper to handle delimiter logic.
	write := func(s string) {
		if buf.Len() == 0 && options.startWithDelimiter {
			buf.WriteString(options.delimiter)
		}
		buf.WriteString(s)
		buf.WriteString(options.delimiter)
	}

	// Build the path based on inclusion/exclusion options.
	if !options.withoutApp {
		write(app.App)
	}
	write(resource)
	if !options.withoutEnv {
		write(app.Environment)
	}
	if !options.withoutService {
		if options.withServiceId {
			write(app.Instance.ID)
		} else {
			write(app.Instance.Group)
		}
	}
	if !options.withoutVersion {
		write(app.Instance.Version)
	}
	if !options.withoutRegion {
		write(app.Region)
	}
	if !options.withoutAz {
		write(app.AZ)
	}
	if key != "" {
		write(key)
	}

	// Clean up trailing delimiters.
	if !options.endWithDelimiter && buf.Len() > 0 {
		buf.Truncate(buf.Len() - len(options.delimiter))
	}

	s := buf.String()
	bufPool.Put(buf)
	return s
}
