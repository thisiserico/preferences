package example

import "time"

type FirstAPI struct {
	clientName     string
	aPIKey         string
	requestTimeout time.Duration
	dumpDebugLogs  bool
	isDryRun       bool
}

type FirstAPIConfig struct {
	ClientName     string
	APIKey         string
	RequestTimeout time.Duration
	DumpDebugLogs  bool
	IsDryRun       bool
}

func NewFirstAPI(config FirstAPIConfig) FirstAPI {
	return FirstAPI{
		clientName:     config.ClientName,
		aPIKey:         config.APIKey,
		requestTimeout: config.RequestTimeout,
		dumpDebugLogs:  config.DumpDebugLogs,
		isDryRun:       config.IsDryRun,
	}
}
