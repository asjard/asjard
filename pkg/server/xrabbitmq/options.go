package xrabbitmq

type PublishOptions struct {
	Mandatory, Immediate bool
}

type PublishOption func(opts *PublishOptions)

func WithPublishMandatory() PublishOption {
	return func(opts *PublishOptions) {
		opts.Mandatory = true
	}
}

func WithPublishImmediate() PublishOption {
	return func(opts *PublishOptions) {
		opts.Immediate = true
	}
}
