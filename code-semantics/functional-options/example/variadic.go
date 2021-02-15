package example

import "time"

type ThirdAPI struct {
	clientName     string
	aPIKey         string
	requestTimeout time.Duration
	dumpDebugLogs  bool
	isDryRun       bool
}

type Option func(*ThirdAPI)

func NewThirdAPI(clientName, apiKey string, opts ...Option) *ThirdAPI {
	api := &ThirdAPI{
		clientName: clientName,
		aPIKey:     apiKey,
	}

	for _, option := range opts {
		option(api)
	}

	return api
}

func WithRequestTimeout(dur time.Duration) Option {
	return func(api *ThirdAPI) {
		api.requestTimeout = dur
	}
}

func WithDebugLogsEnabled() Option {
	return func(api *ThirdAPI) {
		api.dumpDebugLogs = true
	}
}
func RunningOnDryRunMode() Option {
	return func(api *ThirdAPI) {
		api.isDryRun = true
	}
}
