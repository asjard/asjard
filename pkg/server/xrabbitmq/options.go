package xrabbitmq

type PublishOptions struct {
	Mandatory, Immediate bool
	Exchange, Key        string
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

func WithPublishExchange(exchange string) PublishOption {
	return func(opts *PublishOptions) {
		opts.Exchange = exchange
	}
}

func WithPublishKey(key string) PublishOption {
	return func(opts *PublishOptions) {
		opts.Key = key
	}
}
