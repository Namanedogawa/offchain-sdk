package baseapp

import (
	"errors"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/berachain/offchain-sdk/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// AppBuilder is a builder for an app.
type AppBuilder struct {
	appName   string
	jobs      []job.Basic
	db        ethdb.KeyValueStore
	ethClient eth.Client
	svr       *server.Server
	metrics   telemetry.Metrics
}

// NewAppBuilder creates a new app builder.
func NewAppBuilder(appName string) *AppBuilder {
	return &AppBuilder{
		appName: appName,
		jobs:    []job.Basic{},
		metrics: telemetry.NewNoopMetrics(),
	}
}

// AppName returns the name of the app.
func (ab *AppBuilder) AppName() string {
	return ab.appName
}

// RegisterJob adds a job to the builder.
func (ab *AppBuilder) RegisterJob(job job.Basic) {
	ab.jobs = append(ab.jobs, job)
}

// RegisterDB registers the database.
func (ab *AppBuilder) RegisterDB(db ethdb.KeyValueStore) {
	ab.db = db
}

// RegisterMetrics registers the metrics with optional telemetry configuration.
func (ab *AppBuilder) RegisterMetrics(cfg *telemetry.Config) error {
	var err error
	ab.metrics, err = telemetry.NewMetrics(cfg)
	if err != nil {
		return err
	}

	// Enable metrics on eth client if it's an instance of ChainProviderImpl.
	if chainProvider, ok := ab.ethClient.(*eth.ChainProviderImpl); ok {
		chainProvider.EnableMetrics(ab.metrics)
	}
	return nil
}

// RegisterHTTPServer registers the HTTP server.
func (ab *AppBuilder) RegisterHTTPServer(svr *server.Server) {
	ab.svr = svr
}

// RegisterHTTPHandler registers an HTTP handler.
func (ab *AppBuilder) RegisterHTTPHandler(handler *server.Handler) error {
	if ab.svr == nil {
		return errors.New("must enable the HTTP server to register a handler")
	}
	ab.svr.RegisterHandler(handler)
	return nil
}

// RegisterMiddleware registers middleware for the HTTP server.
func (ab *AppBuilder) RegisterMiddleware(m server.Middleware) error {
	if ab.svr == nil {
		return errors.New("must enable the HTTP server to register middleware")
	}
	ab.svr.RegisterMiddleware(m)
	return nil
}

// RegisterPrometheusTelemetry registers a Prometheus metrics HTTP server.
func (ab *AppBuilder) RegisterPrometheusTelemetry() error {
	if ab.svr == nil {
		return errors.New("must enable the HTTP server to register Prometheus")
	}
	return ab.RegisterHTTPHandler(&server.Handler{Path: "/metrics", Handler: promhttp.Handler()})
}

// RegisterEthClient registers the Ethereum client.
func (ab *AppBuilder) RegisterEthClient(ethClient eth.Client) {
	ab.ethClient = ethClient
}

// BuildApp builds the app.
func (ab *AppBuilder) BuildApp(logger log.Logger) *BaseApp {
	return New(
		ab.appName,
		logger,
		ab.ethClient,
		ab.jobs,
		ab.db,
		ab.svr,
		ab.metrics,
	)
}
