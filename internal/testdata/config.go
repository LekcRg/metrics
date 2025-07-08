package testdata

import "github.com/LekcRg/metrics/internal/config"

const defaultAddr = "localhost:8080"
const defaultLogLvl = "debug"
const defaultIsDev = false
const defaultStoreInterval = 300
const defaultFileStoragePath = "store.json"
const defaultRestore = false

// const defaultReportInterval = 10
// const defaultPollInterval = 2
// const defaultHTTPS = false

// TODO: File with testData
var TestServerConfig = config.ServerConfig{
	CommonConfig: config.CommonConfig{
		LogLvl: defaultLogLvl,
		IsDev:  defaultIsDev,
	},
	Addr:            defaultAddr,
	StoreInterval:   defaultStoreInterval,
	FileStoragePath: defaultFileStoragePath,
	Restore:         defaultRestore,
	SyncSave:        false,
}
