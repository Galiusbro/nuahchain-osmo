# 🎯 Custom Oracle Data Source - Technical Specification & Roadmap

## 📋 Project Overview

**Goal:** Create a custom oracle data source system that collects real-time market data from primary sources without relying on external APIs (except for CoinGecko). This system will provide independent, reliable price feeds for the NuahChain trading ecosystem.

**Timeline:** 4 weeks (28 days)
**Complexity:** Medium to High
**Dependencies:** Web scraping, API integration, consensus algorithms

---

## 🎯 Core Objectives

1. **Independence:** No dependency on external paid APIs
2. **Reliability:** Multiple data sources with validation
3. **Scalability:** Easy addition of new data sources
4. **Decentralization:** Distributed oracle network
5. **Cost Efficiency:** Minimal operational costs

---

## 📊 Phase 1: Basic Web Scraping (HTML Parsing)
**Duration:** 1-2 days
**Complexity:** Low
**Status:** Pending

### Objectives
- Implement simple HTML parsing for static websites
- Create basic data extraction mechanisms
- Establish foundation for more complex scraping

### Data Sources
| Source | Asset Type | URL Pattern | Selector |
|--------|------------|-------------|----------|
| investing.com | Forex | `/currencies/{symbol}` | `.instrument-price_last__KQzyA` |
| x-rates.com | Forex | `/table/?from=USD&amount=1` | `td.rtRates` |
| coinmarketcap.com | Crypto | `/currencies/{symbol}/` | `.priceValue` |
| kitco.com | Commodities | `/gold-price-today-usa/` | `.spot-price` |
| yahoo.com/finance | Stocks | `/quote/{symbol}` | `[data-field="regularMarketPrice"]` |

### Technical Implementation
```go
type SimpleScraper struct {
    client    *http.Client
    userAgent string
    rateLimit time.Duration
}

type ScrapingResult struct {
    Symbol    string    `json:"symbol"`
    Price     float64   `json:"price"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
    Success   bool      `json:"success"`
    Error     string    `json:"error,omitempty"`
}
```

### Deliverables
- [ ] Basic HTTP client with rate limiting
- [ ] HTML parser for 5 primary sources
- [ ] Error handling and retry mechanisms
- [ ] Unit tests for each data source
- [ ] Configuration system for selectors

---

## 📊 Phase 2: API Integration with Free Sources
**Duration:** 2-3 days
**Complexity:** Medium
**Status:** Pending

### Objectives
- Integrate with free API services
- Implement rate limiting and quota management
- Create unified interface for all data sources

### Free API Sources
| Service | Rate Limit | Quota | Asset Types | Cost |
|---------|------------|-------|-------------|------|
| CoinGecko | 50 req/min | Free | Crypto | Free |
| Alpha Vantage | 5 req/min | Free | Stocks, Forex | Free |
| Fixer.io | 100 req/month | Free | Forex | Free |
| CurrencyLayer | 1000 req/month | Free | Forex | Free |
| TwelveData | 800 req/day | Free | All | Free |

### Technical Implementation
```go
type DataSource interface {
    GetPrice(symbol string) (*PriceData, error)
    GetName() string
    GetRateLimit() time.Duration
    IsAvailable() bool
    GetQuota() (used, limit int)
}

type PriceData struct {
    Symbol     string    `json:"symbol"`
    Price      float64   `json:"price"`
    Timestamp  time.Time `json:"timestamp"`
    Source     string    `json:"source"`
    Confidence float64   `json:"confidence"` // 0-1
    Metadata   map[string]interface{} `json:"metadata"`
}
```

### API Implementations
```go
type CoinGeckoSource struct {
    apiKey string
    client *http.Client
    quota  QuotaManager
}

type AlphaVantageSource struct {
    apiKey string
    client *http.Client
    quota  QuotaManager
}

type FixerSource struct {
    apiKey string
    client *http.Client
    quota  QuotaManager
}
```

### Deliverables
- [ ] API client implementations for 5 services
- [ ] Quota management system
- [ ] Rate limiting with backoff
- [ ] Unified data interface
- [ ] Configuration management
- [ ] Monitoring and alerting

---

## 📊 Phase 3: Advanced Parsing (JavaScript, Dynamic Content)
**Duration:** 3-4 days
**Complexity:** High
**Status:** Pending

### Objectives
- Handle Single Page Applications (SPA)
- Parse JavaScript-rendered content
- Implement proxy rotation for anti-bot measures
- Create robust error handling for dynamic content

### Advanced Sources
| Source | Type | Challenge | Solution |
|--------|------|-----------|----------|
| TradingView | Charts | JavaScript rendering | Playwright |
| Bloomberg | News/Data | Anti-bot protection | Proxy rotation |
| Reuters | Financial | Dynamic loading | Wait strategies |
| MarketWatch | Real-time | Rate limiting | Distributed scraping |

### Technical Implementation
```go
type AdvancedScraper struct {
    browser     *playwright.Browser
    proxies     []string
    currentProxy int
    userAgents  []string
    currentUA   int
}

type ScrapingConfig struct {
    URL           string            `json:"url"`
    Selector      string            `json:"selector"`
    WaitFor       time.Duration     `json:"wait_for"`
    JavaScript    string            `json:"javascript"`
    Headers       map[string]string `json:"headers"`
    Cookies       []http.Cookie     `json:"cookies"`
    ProxyRequired bool              `json:"proxy_required"`
    AntiBot       bool              `json:"anti_bot"`
}
```

### Browser Automation
```go
func (s *AdvancedScraper) GetTradingViewPrice(symbol string) (float64, error) {
    page, err := s.browser.NewPage()
    if err != nil {
        return 0, err
    }

    // Configure page
    page.SetExtraHTTPHeaders(map[string]string{
        "User-Agent": s.getRandomUserAgent(),
    })

    // Navigate with proxy
    err = page.Goto(fmt.Sprintf("https://www.tradingview.com/symbols/%s/"),
        playwright.PageGotoOptions{
            WaitUntil: playwright.WaitUntilStateNetworkidle,
        })
    if err != nil {
        return 0, err
    }

    // Wait for price element
    page.WaitForSelector(".tv-symbol-price-quote__value")

    // Extract price via JavaScript
    price, err := page.Evaluate(`
        () => {
            const priceEl = document.querySelector('.tv-symbol-price-quote__value');
            return priceEl ? parseFloat(priceEl.textContent.replace(/[^0-9.-]/g, '')) : null;
        }
    `)

    return price.(float64), nil
}
```

### Deliverables
- [ ] Playwright integration
- [ ] Proxy rotation system
- [ ] User agent rotation
- [ ] Anti-bot detection and evasion
- [ ] JavaScript execution framework
- [ ] Dynamic content handling
- [ ] Performance optimization

---

## 📊 Phase 4: Data Aggregation and Validation
**Duration:** 4-5 days
**Complexity:** High
**Status:** Pending

### Objectives
- Implement intelligent data aggregation
- Create validation mechanisms for data quality
- Develop confidence scoring system
- Build outlier detection algorithms

### Aggregation Algorithms
```go
type PriceAggregator struct {
    sources    []DataSource
    weights    map[string]float64
    validators []PriceValidator
    config     AggregationConfig
}

type AggregationConfig struct {
    MinSources     int           `json:"min_sources"`
    MaxAge         time.Duration `json:"max_age"`
    OutlierThreshold float64     `json:"outlier_threshold"`
    ConfidenceThreshold float64  `json:"confidence_threshold"`
}
```

### Validation System
```go
type PriceValidator interface {
    Validate(price *PriceData) error
    GetConfidence(price *PriceData) float64
    GetWeight() float64
}

// Outlier Detection
type OutlierValidator struct {
    maxDeviation float64
    method        string // "iqr", "zscore", "modified_zscore"
}

// Timestamp Validation
type TimestampValidator struct {
    maxAge time.Duration
    timezone string
}

// Cross-Source Validation
type CrossSourceValidator struct {
    minSources int
    maxDeviation float64
}
```

### Aggregation Logic
```go
func (a *PriceAggregator) AggregatePrice(symbol string) (*AggregatedPrice, error) {
    var prices []*PriceData
    var errors []error

    // Collect data from all sources
    for _, source := range a.sources {
        if !source.IsAvailable() {
            continue
        }

        price, err := source.GetPrice(symbol)
        if err != nil {
            errors = append(errors, err)
            continue
        }

        // Validate price
        valid := true
        for _, validator := range a.validators {
            if err := validator.Validate(price); err != nil {
                valid = false
                break
            }
        }

        if valid {
            prices = append(prices, price)
        }
    }

    if len(prices) < a.config.MinSources {
        return nil, fmt.Errorf("insufficient valid sources: %d/%d", len(prices), a.config.MinSources)
    }

    // Calculate weighted average
    return a.calculateWeightedAverage(prices), nil
}

func (a *PriceAggregator) calculateWeightedAverage(prices []*PriceData) *AggregatedPrice {
    var totalWeight float64
    var weightedSum float64
    var totalConfidence float64

    for _, price := range prices {
        weight := a.weights[price.Source]
        confidence := price.Confidence

        totalWeight += weight
        weightedSum += price.Price * weight * confidence
        totalConfidence += confidence
    }

    return &AggregatedPrice{
        Price:      weightedSum / totalWeight,
        Confidence: totalConfidence / float64(len(prices)),
        Sources:    len(prices),
        Timestamp:  time.Now(),
    }
}
```

### Deliverables
- [ ] Multi-source aggregation engine
- [ ] Outlier detection algorithms
- [ ] Confidence scoring system
- [ ] Historical data analysis
- [ ] Performance metrics
- [ ] Quality assurance framework

---

## 📊 Phase 5: Decentralized Oracle Network
**Duration:** 5-7 days
**Complexity:** Very High
**Status:** Pending

### Objectives
- Create distributed oracle network
- Implement consensus mechanisms
- Build slashing system for bad actors
- Develop reputation system

### Network Architecture
```go
type OracleNode struct {
    ID          string            `json:"id"`
    Address     sdk.AccAddress    `json:"address"`
    Sources     []DataSource      `json:"sources"`
    Reputation  float64           `json:"reputation"`
    Stake       sdk.Int           `json:"stake"`
    LastUpdate  time.Time         `json:"last_update"`
    Performance Metrics           `json:"performance"`
}

type OracleNetwork struct {
    nodes       map[string]*OracleNode
    consensus   ConsensusAlgorithm
    slashing    SlashingModule
    reputation  ReputationSystem
    governance  GovernanceModule
}
```

### Consensus Mechanisms
```go
type ConsensusAlgorithm interface {
    CalculateConsensus(prices []*PriceData) (*ConsensusPrice, error)
    DetectByzantine(prices []*PriceData) []string
    CalculateRewards(prices []*PriceData) map[string]sdk.Int
    UpdateReputation(nodes []*OracleNode, results []*PriceData)
}

// Median Consensus
type MedianConsensus struct {
    maxDeviation float64
    minSources   int
    reputationThreshold float64
}

func (m *MedianConsensus) CalculateConsensus(prices []*PriceData) (*ConsensusPrice, error) {
    if len(prices) < m.minSources {
        return nil, fmt.Errorf("insufficient sources")
    }

    // Sort by price
    sort.Slice(prices, func(i, j int) bool {
        return prices[i].Price < prices[j].Price
    })

    median := prices[len(prices)/2].Price

    // Check deviations
    var validPrices []*PriceData
    for _, price := range prices {
        deviation := math.Abs(price.Price - median) / median
        if deviation <= m.maxDeviation {
            validPrices = append(validPrices, price)
        }
    }

    if len(validPrices) < m.minSources {
        return nil, fmt.Errorf("too many outliers")
    }

    return &ConsensusPrice{
        Price:       median,
        Confidence:  float64(len(validPrices)) / float64(len(prices)),
        Sources:     len(validPrices),
        Timestamp:   time.Now(),
    }, nil
}
```

### Slashing System
```go
type SlashingModule struct {
    slashingRate    sdk.Dec
    jailDuration    time.Duration
    minStake        sdk.Int
    maxSlashPercent sdk.Dec
}

func (s *SlashingModule) SlashNode(nodeID string, reason SlashReason, amount sdk.Int) error {
    // Implement slashing logic
    // - Reduce stake
    // - Update reputation
    // - Jail node if necessary
    // - Emit events
}
```

### Deliverables
- [ ] Distributed oracle network
- [ ] Consensus algorithms (median, weighted, reputation-based)
- [ ] Slashing mechanism
- [ ] Reputation system
- [ ] Governance framework
- [ ] Network monitoring
- [ ] Economic incentives

---

## 🛠️ Technical Architecture

### Project Structure
```
x/oracle/
├── sources/              # Data source implementations
│   ├── scraper.go        # Basic web scraping
│   ├── api.go           # API integrations
│   ├── advanced.go      # Advanced parsing (Playwright)
│   ├── config.go        # Source configurations
│   └── manager.go       # Source management
├── aggregator/          # Data aggregation
│   ├── validator.go     # Price validation
│   ├── consensus.go    # Consensus algorithms
│   ├── weights.go      # Source weighting
│   └── aggregator.go   # Main aggregation logic
├── network/            # Decentralized network
│   ├── node.go        # Oracle nodes
│   ├── consensus.go   # Network consensus
│   ├── slashing.go    # Slashing mechanism
│   ├── reputation.go  # Reputation system
│   └── governance.go  # Governance
├── keeper/             # Keeper with new logic
│   ├── sources.go     # Source management
│   ├── aggregator.go  # Aggregation
│   ├── network.go     # Network operations
│   └── keeper.go      # Main keeper
└── types/             # Type definitions
    ├── price.go       # Price data structures
    ├── consensus.go   # Consensus types
    └── network.go    # Network types
```

### Configuration Files
```yaml
# config/sources.yaml
sources:
  forex:
    - name: "investing.com"
      url: "https://www.investing.com/currencies/{symbol}"
      selector: ".instrument-price_last__KQzyA"
      rate_limit: "1s"
      weight: 0.3
      timeout: "10s"
      retries: 3

    - name: "x-rates.com"
      url: "https://www.x-rates.com/table/?from=USD&amount=1"
      selector: "td.rtRates"
      rate_limit: "2s"
      weight: 0.2

  crypto:
    - name: "coingecko"
      api_url: "https://api.coingecko.com/api/v3/simple/price"
      rate_limit: "1s"
      weight: 0.4
      api_key: "${COINGECKO_API_KEY}"

    - name: "coinmarketcap"
      api_url: "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
      rate_limit: "1s"
      weight: 0.3
      api_key: "${CMC_API_KEY}"

# config/aggregation.yaml
aggregation:
  min_sources: 3
  max_age: "5m"
  outlier_threshold: 0.1
  confidence_threshold: 0.7
  consensus_method: "median"

# config/network.yaml
network:
  min_stake: "1000000unuah"
  slashing_rate: "0.01"
  jail_duration: "24h"
  reputation_decay: 0.95
  consensus_threshold: 0.67
```

---

## 🚀 Implementation Roadmap

### Week 1: Foundation (Phases 1-2)
**Days 1-2: Phase 1 - Basic Web Scraping**
- [ ] Set up HTTP client with rate limiting
- [ ] Implement HTML parsing for 5 sources
- [ ] Create error handling and retry logic
- [ ] Write unit tests

**Days 3-5: Phase 2 - API Integration**
- [ ] Integrate CoinGecko API
- [ ] Add Alpha Vantage integration
- [ ] Implement quota management
- [ ] Create unified data interface

### Week 2: Advanced Parsing (Phase 3)
**Days 6-9: Phase 3 - Advanced Parsing**
- [ ] Set up Playwright for JavaScript rendering
- [ ] Implement proxy rotation
- [ ] Add anti-bot detection
- [ ] Create dynamic content handlers

### Week 3: Aggregation (Phase 4)
**Days 10-14: Phase 4 - Data Aggregation**
- [ ] Implement aggregation algorithms
- [ ] Create validation system
- [ ] Build confidence scoring
- [ ] Add outlier detection

### Week 4: Network (Phase 5)
**Days 15-21: Phase 5 - Decentralized Network**
- [ ] Design oracle network architecture
- [ ] Implement consensus mechanisms
- [ ] Build slashing system
- [ ] Create reputation framework

### Week 5: Testing & Deployment
**Days 22-28: Testing & Deployment**
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Deployment preparation

---

## 📊 Success Metrics

### Technical Metrics
- **Data Accuracy:** >99% price accuracy vs reference sources
- **Uptime:** >99.9% availability
- **Latency:** <5 seconds average response time
- **Coverage:** Support for 50+ trading pairs
- **Sources:** 10+ active data sources

### Economic Metrics
- **Cost Efficiency:** <$100/month operational costs
- **Reliability:** <0.1% false positive rate
- **Scalability:** Support for 1000+ concurrent requests
- **Performance:** <1% CPU usage per oracle node

### Network Metrics
- **Decentralization:** 5+ active oracle nodes
- **Consensus:** >67% agreement rate
- **Reputation:** >0.8 average node reputation
- **Slashing:** <1% slashing rate

---

## 🔧 Dependencies & Requirements

### Technical Dependencies
- **Go 1.21+** for core implementation
- **Playwright** for advanced scraping
- **Cosmos SDK v0.47+** for blockchain integration
- **Docker** for containerization
- **Prometheus** for monitoring

### External Dependencies
- **Proxy Services** for anti-bot measures
- **Cloud Infrastructure** for scaling
- **Monitoring Tools** for observability
- **CI/CD Pipeline** for deployment

### Team Requirements
- **Backend Developer** (Go, web scraping)
- **DevOps Engineer** (infrastructure, monitoring)
- **Data Engineer** (aggregation, validation)
- **Blockchain Developer** (Cosmos SDK, consensus)

---

## 💡 Key Benefits

### Independence
- No reliance on external paid APIs
- Full control over data sources
- Customizable aggregation logic

### Reliability
- Multiple data sources with validation
- Consensus mechanisms for accuracy
- Fault tolerance and redundancy

### Cost Efficiency
- Minimal operational costs
- No per-request API fees
- Scalable infrastructure

### Decentralization
- Distributed oracle network
- Community-driven governance
- Economic incentives for accuracy

---

## 🎯 Next Steps

1. **Review and approve** this technical specification
2. **Set up development environment** with required dependencies
3. **Begin Phase 1** implementation (basic web scraping)
4. **Establish monitoring** and testing frameworks
5. **Plan team allocation** and resource requirements

---

**Document Version:** 1.0
**Last Updated:** 2024-01-XX
**Status:** Ready for Implementation
**Owner:** Oracle Development Team
