package marketdata

import "time"

// NewServiceWithSource selects a fetcher by source id and returns a Service.
// Supported sources: "yahoo_http" (real HTTP), "yahoo" (stub/synthetic).
func NewServiceWithSource(source string) *Service {
    switch source {
    case "yahoo_http":
        return NewService(NewYahooHTTPFetcher(8 * time.Second))
    case "yahoo":
        fallthrough
    default:
        return NewService(NewYahooFetcher())
    }
}


