package runtime

// Options defines how a structured string (like a path or key) should be constructed
// from the application's runtime metadata.
type Options struct {
	// startWithDelimiter if true, the resulting string will begin with the delimiter (e.g., "/app/...")
	startWithDelimiter bool
	// endWithDelimiter if true, the resulting string will end with the delimiter (e.g., ".../service/")
	endWithDelimiter bool
	// delimiter is the character used to join different metadata components (default is "/")
	delimiter string
	// withoutApp if true, the APP name will be omitted from the output
	withoutApp bool
	// withoutRegion if true, the geographical region will be omitted
	withoutRegion bool
	// withoutAz if true, the Availability Zone (AZ) will be omitted
	withoutAz bool
	// withoutEnv if true, the environment (prod/dev) will be omitted
	withoutEnv bool
	// withoutService if true, the service name will be omitted
	withoutService bool
	// withoutVersion if true, the service version will be omitted
	withoutVersion bool
	// withServiceId if true, the unique Service ID will be included (usually replacing or following the name)
	withServiceId bool
}

// Option is a function type used to modify the Options struct.
type Option func(options *Options)

// WithStartWithDelimiter sets whether the generated string starts with a delimiter.
func WithStartWithDelimiter(value bool) Option {
	return func(options *Options) {
		options.startWithDelimiter = value
	}
}

// WithEndWithDelimiter sets whether the generated string ends with a delimiter.
func WithEndWithDelimiter(value bool) Option {
	return func(options *Options) {
		options.endWithDelimiter = value
	}
}

// WithDelimiter specifies the separator character (e.g., ".", "/", or ":").
func WithDelimiter(delimiter string) Option {
	return func(options *Options) {
		options.delimiter = delimiter
	}
}

// WithoutApp toggles the inclusion of the Application name.
func WithoutApp(value bool) Option {
	return func(options *Options) {
		options.withoutApp = value
	}
}

// WithoutRegion toggles the inclusion of the Region information.
func WithoutRegion(value bool) Option {
	return func(options *Options) {
		options.withoutRegion = value
	}
}

// WithoutAz toggles the inclusion of the Availability Zone.
func WithoutAz(value bool) Option {
	return func(options *Options) {
		options.withoutAz = value
	}
}

// WithoutEnv toggles the inclusion of the Environment name.
func WithoutEnv(value bool) Option {
	return func(options *Options) {
		options.withoutEnv = value
	}
}

// WithoutService toggles the inclusion of the Service name.
func WithoutService(value bool) Option {
	return func(options *Options) {
		options.withoutService = value
	}
}

// WithoutVersion toggles the inclusion of the Service version.
func WithoutVersion(value bool) Option {
	return func(options *Options) {
		options.withoutVersion = value
	}
}

// WithServiceId determines if the unique Service Instance ID should be used.
func WithServiceId(value bool) Option {
	return func(options *Options) {
		options.withServiceId = value
	}
}

// defaultOptions provides the initial state for string construction,
// defaulting to a "/" delimiter (standard for URL paths or Unix file paths).
func defaultOptions() *Options {
	return &Options{
		delimiter: "/",
	}
}
