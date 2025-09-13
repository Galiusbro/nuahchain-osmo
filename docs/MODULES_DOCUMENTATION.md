# Osmosis Blockchain Modules Documentation

This document provides comprehensive documentation for the **PegKeeper** and **USDOracle** modules integrated into the Osmosis blockchain.

## Table of Contents

1. [USD Oracle Module](#usd-oracle-module)
2. [PegKeeper Module](#pegkeeper-module)
3. [Integration Guide](#integration-guide)
4. [API Reference](#api-reference)

---

## USD Oracle Module

### Overview

The USD Oracle module provides a decentralized price oracle system that tracks USD prices from multiple sources, maintaining historical price data and enabling accurate price feeds for stablecoins and other financial applications within the Osmosis ecosystem.

### Key Features

- **Multi-Source Price Aggregation**: Supports multiple price feed sources with configurable weights
- **Real-Time Price Updates**: Continuous USD price tracking with configurable update intervals
- **Historical Data Storage**: Maintains comprehensive price history for analysis and TWAP calculations
- **Deviation Monitoring**: Automatic detection of price deviations with configurable thresholds
- **Governance Integration**: Parameter updates through governance proposals
- **Security Controls**: Authority-based access control for price updates and configuration

### Architecture

#### Core Components

1. **Keeper**: Core business logic for price management and validation
2. **Types**: Message and query definitions with protobuf serialization
3. **Client**: CLI commands and gRPC interfaces for interaction
4. **Genesis**: Module initialization and state export/import

#### Data Structures

##### USDPrice
```protobuf
message USDPrice {
  string price = 1;                    // USD price as decimal
  google.protobuf.Timestamp timestamp = 2;  // Recording timestamp
  string source = 3;                   // Price source identifier
  int64 block_height = 4;             // Block height when recorded
}
```

##### PriceSource
```protobuf
message PriceSource {
  string name = 1;      // Source name (e.g., "coingecko", "binance")
  string weight = 2;    // Weight in price aggregation
  bool enabled = 3;     // Whether source is active
  string url = 4;       // API endpoint URL
}
```

##### Parameters
```protobuf
message Params {
  bool enabled = 1;                           // Module enabled status
  string admin = 2;                          // Admin account address
  uint64 update_interval = 3;                // Update interval in seconds
  string price_deviation_threshold = 4;      // Deviation threshold (5% default)
}
```

### Functionality

#### Price Management
- **SetCurrentPrice**: Updates the current USD price with validation
- **GetCurrentPrice**: Retrieves the latest USD price data
- **AddPriceHistory**: Stores historical price entries
- **GetPriceHistoryList**: Retrieves price history with pagination

#### Price Source Management
- **SetPriceSource**: Configures price source parameters
- **GetAllPriceSources**: Lists all configured price sources
- **Price Aggregation**: Weighted average calculation from multiple sources

#### Deviation Detection
- **CalculatePriceDeviation**: Computes current price deviation from target
- **IsWithinThreshold**: Validates price changes against deviation limits

### CLI Commands

```bash
# Query current USD price
osmosis query usdoracle usd-price

# Query price history (with optional limit)
osmosis query usdoracle price-history [limit]

# Query price deviation
osmosis query usdoracle price-deviation

# Query module parameters
osmosis query usdoracle params
```

### gRPC Endpoints

- `GET /osmosis/usdoracle/params` - Module parameters
- `GET /osmosis/usdoracle/usd_price` - Current USD price
- `GET /osmosis/usdoracle/price_history` - Price history
- `GET /osmosis/usdoracle/price_deviation` - Price deviation metrics

---

## PegKeeper Module

### Overview

The PegKeeper module implements an automated stabilization mechanism that maintains price pegs for stablecoins by dynamically adjusting token supply based on price deviations from target values. It integrates with the USD Oracle module to monitor price data and execute supply adjustments.

### Key Features

- **Automated Peg Maintenance**: Continuous monitoring and adjustment of token prices
- **Supply Management**: Dynamic minting and burning of tokens to maintain price stability
- **Configurable Parameters**: Adjustable deviation thresholds, adjustment factors, and intervals
- **Historical Tracking**: Complete audit trail of all supply adjustments
- **Oracle Integration**: Seamless integration with USD Oracle for price data
- **Safety Controls**: Maximum supply change limits and minimum adjustment intervals

### Architecture

#### Core Components

1. **Keeper**: Implements stabilization logic and supply adjustment mechanisms
2. **Types**: Defines peg state, adjustment parameters, and historical records
3. **Integration**: Interfaces with Bank, Mint, and USD Oracle modules
4. **Governance**: Parameter updates through governance proposals

#### Data Structures

##### Stabilization Parameters
```protobuf
message Params {
  string max_deviation_threshold = 1;        // Maximum allowed price deviation
  string adjustment_factor = 2;              // Aggressiveness of adjustments
  int64 min_adjustment_interval = 3;         // Minimum time between adjustments
  string max_supply_change_per_adjustment = 4; // Maximum supply change per adjustment
  string oracle_module = 5;                  // Oracle module identifier
  bool enabled = 6;                          // Stabilization mechanism status
  string target_denom = 7;                   // Target denomination (e.g., "nuah")
  string reference_denom = 8;                // Reference denomination (e.g., "usd")
  string target_price = 9;                   // Desired price (usually "1.0")
}
```

##### Peg State
```protobuf
message PegState {
  string target_denom = 1;                   // Pegged denomination
  string reference_denom = 2;                // Reference denomination
  string current_price = 3;                  // Current market price
  string target_price = 4;                   // Target price
  string deviation = 5;                      // Current deviation percentage
  bool is_active = 6;                        // Peg mechanism status
  google.protobuf.Timestamp last_adjustment_time = 7; // Last adjustment timestamp
}
```

##### Supply Adjustment Record
```protobuf
message SupplyAdjustment {
  google.protobuf.Timestamp timestamp = 1;   // Adjustment timestamp
  string price_before = 2;                   // Price before adjustment
  string price_after = 3;                    // Expected price after adjustment
  string supply_change_amount = 4;           // Amount of supply change
  string supply_change_denom = 5;            // Denomination of supply change
  string deviation = 6;                      // Price deviation that triggered adjustment
  string adjustment_type = 7;                // Type: "mint" or "burn"
}
```

### Stabilization Algorithm

#### Price Monitoring
1. **Continuous Monitoring**: Regular price checks via USD Oracle integration
2. **Deviation Calculation**: Comparison of current price against target price
3. **Threshold Evaluation**: Triggering adjustments when deviation exceeds limits

#### Supply Adjustment Logic
1. **Deviation Assessment**: Calculate percentage deviation from target price
2. **Adjustment Calculation**: Determine required supply change using adjustment factor
3. **Safety Checks**: Validate against maximum change limits and time intervals
4. **Execution**: Mint or burn tokens through Bank/Mint module integration
5. **Recording**: Store adjustment details for historical tracking

#### Adjustment Types
- **Mint Tokens**: When price is above target (reduce supply to increase price)
- **Burn Tokens**: When price is below target (increase supply to decrease price)

### Expected Keepers Integration

#### Bank Keeper
- Supply management and token transfers
- Balance queries and validation
- Minting and burning operations

#### Mint Keeper
- Minter parameter management
- Inflation control integration

#### USD Oracle Keeper
- Current price retrieval
- Price history access
- Deviation calculations
- Threshold validation

### CLI Commands

```bash
# Query peg state
osmosis query pegkeeper peg-state

# Query adjustment history
osmosis query pegkeeper adjustment-history [limit] [offset]

# Query module parameters
osmosis query pegkeeper params
```

### gRPC Endpoints

- `GET /osmosis/pegkeeper/params` - Module parameters
- `GET /osmosis/pegkeeper/peg_state` - Current peg state
- `GET /osmosis/pegkeeper/adjustment_history` - Supply adjustment history

---

## Integration Guide

### Module Dependencies

#### USD Oracle Module
- **Dependencies**: None (standalone module)
- **Provides**: Price data services to other modules

#### PegKeeper Module
- **Dependencies**:
  - Bank Module (supply management)
  - Mint Module (token minting/burning)
  - USD Oracle Module (price data)
- **Provides**: Automated price stabilization

### Integration Steps

1. **Module Registration**: Both modules are registered in `app/modules.go`
2. **Keeper Initialization**: Keepers initialized in `app/keepers/keepers.go`
3. **Store Keys**: KV store keys added to application store
4. **Genesis Integration**: Genesis state handling for both modules
5. **API Integration**: REST and gRPC endpoints exposed via API server

### Configuration

#### USD Oracle Configuration
```go
// Default parameters
Params{
    Enabled: true,
    Admin: "", // Set via governance
    UpdateInterval: 60, // 60 seconds
    PriceDeviationThreshold: "0.05", // 5%
}
```

#### PegKeeper Configuration
```go
// Default parameters
Params{
    MaxDeviationThreshold: "0.02", // 2%
    AdjustmentFactor: "0.1", // 10% adjustment factor
    MinAdjustmentInterval: 300, // 5 minutes
    MaxSupplyChangePerAdjustment: "0.05", // 5% max change
    OracleModule: "usdoracle",
    Enabled: true,
    TargetDenom: "nuah",
    ReferenceDenom: "usd",
    TargetPrice: "1.0",
}
```

---

## API Reference

### USD Oracle API Endpoints

#### GET /api/v1/usd/price
Returns current USD price data

**Response:**
```json
{
  "price": "1.0000",
  "timestamp": "2024-01-15T10:30:00Z",
  "source": "aggregated",
  "block_height": 12345678
}
```

### PegKeeper API Endpoints

#### GET /api/v1/pegkeeper/status
Returns current peg status and recent adjustments

**Response:**
```json
{
  "peg_state": {
    "target_denom": "nuah",
    "reference_denom": "usd",
    "current_price": "1.0150",
    "target_price": "1.0000",
    "deviation": "0.015",
    "is_active": true,
    "last_adjustment_time": "2024-01-15T10:25:00Z"
  },
  "recent_adjustments": [
    {
      "timestamp": "2024-01-15T10:25:00Z",
      "adjustment_type": "burn",
      "supply_change_amount": "1000000",
      "price_before": "1.0150",
      "price_after": "1.0000",
      "deviation": "0.015"
    }
  ]
}
```

### Security Considerations

1. **Authority Control**: Only authorized addresses can update prices and parameters
2. **Deviation Limits**: Automatic rejection of prices outside acceptable ranges
3. **Supply Change Limits**: Maximum supply changes per adjustment to prevent manipulation
4. **Time Intervals**: Minimum intervals between adjustments to prevent rapid oscillations
5. **Governance**: All parameter changes require governance approval
6. **Multi-Source Validation**: Price aggregation from multiple sources for reliability

### Testing

Both modules include comprehensive test suites:

```bash
# Run USD Oracle tests
go test ./x/usdoracle/...

# Run PegKeeper tests
go test ./x/pegkeeper/...

# Run integration tests
go test ./app/...
```

### Contributing

When contributing to these modules:

1. Ensure all tests pass
2. Follow Go coding conventions
3. Update documentation for any API changes
4. Consider security implications of changes
5. Test integration with dependent modules

### License

These modules are part of the Osmosis project and follow the same license terms as the main repository.
