package arg

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/dataproviders/arg/queries"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/cache"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	"time"
)

type ARGDataProviderFactory struct {
}

func NewARGDataProviderFactory () *ARGDataProviderFactory{
	return &ARGDataProviderFactory{}
}

func (factory *ARGDataProviderFactory) CreateARGDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider, argClient IARGClient, queryGenerator queries.IARGQueryGenerator, cacheClient cache.ICacheClient, configuration *ARGDataProviderConfiguration) (*ARGDataProvider, error){
	cacheExpirationTimeScannedResults, err := time.ParseDuration(configuration.CacheExpirationTimeScannedResults)
	if err != nil{
		return nil, errors.Wrapf(err, "Given CreateARGDataProvider.CacheExpirationTimeScannedResults string is not a valid time duration")
	}
	cacheExpirationTimeUnscannedResults, err := time.ParseDuration(configuration.CacheExpirationTimeUnscannedResults)
	if err != nil{
		return nil, errors.Wrapf(err, "Given CreateARGDataProvider.CacheExpirationTimeUnscannedResults string is not a valid time duration")
	}
	cacheExpirationTimeForErrors, err := time.ParseDuration(configuration.CacheExpirationTimeForErrors)
	if err != nil{
		return nil, errors.Wrapf(err, "Given CreateARGDataProvider.cacheExpirationTimeForErrors string is not a valid time duration")
	}
	argDataProviderCacheClient := &ARGDataProviderCacheClient{
		cacheClient: cacheClient,
		CacheExpirationTimeScannedResults: cacheExpirationTimeScannedResults,
		CacheExpirationTimeUnscannedResults: cacheExpirationTimeUnscannedResults,
		CacheExpirationTimeForErrors: cacheExpirationTimeForErrors,
	}
	return NewARGDataProvider(instrumentationProvider, argClient, queryGenerator, argDataProviderCacheClient), nil
}
