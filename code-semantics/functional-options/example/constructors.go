package example

import "time"

type SecondAPI struct {
	clientName     string
	aPIKey         string
	requestTimeout time.Duration
	dumpDebugLogs  bool
	isDryRun       bool
}

func NewSecondAPI(clientName, apiKey string) *SecondAPI {
	return &SecondAPI{
		clientName: clientName,
		aPIKey:     apiKey,
	}
}

func (api *SecondAPI) WithRequestTimeout(dur time.Duration) {
	api.requestTimeout = dur
}
