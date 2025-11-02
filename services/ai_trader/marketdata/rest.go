package marketdata

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// REST exposes minimal HTTP endpoints for market data.
type REST struct {
	svc          *Service
	repo         Repository
	authzGrantor AuthzGrantor
	feeGrantor   FeeGrantor
	limiter      *RateLimiter
	corsOrigins  []string
}

type RESTOption func(*REST)

func NewREST(svc *Service, opts ...RESTOption) *REST {
	var repo Repository
	if svc != nil {
		repo = svc.Repository()
	}
	r := &REST{svc: svc, repo: repo}
	if grantor := newRemoteAuthzGrantor(os.Getenv("AI_TRADER_NODE_GRPC")); grantor != nil {
		r.authzGrantor = grantor
	}
	if fee := newRemoteFeeGrantor(os.Getenv("AI_TRADER_NODE_GRPC")); fee != nil {
		r.feeGrantor = fee
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.limiter == nil {
		r.limiter = newLimiterFromEnv()
	}
	return r
}

// WithAuthzGrantor overrides the default grantor; primarily used in tests.
func (r *REST) WithAuthzGrantor(grantor AuthzGrantor) *REST {
	r.authzGrantor = grantor
	return r
}

// WithFeeGrantor overrides the default feegrant client (tests).
func (r *REST) WithFeeGrantor(grantor FeeGrantor) *REST {
	r.feeGrantor = grantor
	return r
}

// WithRateLimiter allows callers to provide custom limiter.
func WithRateLimiter(l *RateLimiter) RESTOption {
	return func(r *REST) { r.limiter = l }
}

// WithCORS configures allowed CORS origins.
func WithCORS(origins []string) RESTOption {
	return func(r *REST) { r.corsOrigins = origins }
}

type ctxKey string

const apiKeyRecordKey ctxKey = "apiKeyRecord"

func apiKeyRecordFromContext(ctx context.Context) (*APIKeyRecord, bool) {
	rec, ok := ctx.Value(apiKeyRecordKey).(*APIKeyRecord)
	return rec, ok
}

func requireScope(rec *APIKeyRecord, allowed ...string) bool {
	for _, scope := range allowed {
		if rec.Scope == scope {
			return true
		}
	}
	return false
}

func (r *REST) logAudit(userID *int64, botID *int64, apiKeyID *int64, action string, meta map[string]any) error {
	if r.repo == nil {
		return nil
	}
	action = strings.TrimSpace(action)
	if action == "" {
		return errors.New("audit action required")
	}
	if meta == nil {
		meta = map[string]any{}
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return r.repo.LogAudit(userID, botID, apiKeyID, action, string(payload))
}

type auditEntry struct {
	ID       int64           `json:"id"`
	UserID   *int64          `json:"user_id,omitempty"`
	BotID    *int64          `json:"bot_id,omitempty"`
	APIKeyID *int64          `json:"api_key_id,omitempty"`
	Action   string          `json:"action"`
	Meta     json.RawMessage `json:"meta"`
	Created  time.Time       `json:"created_at"`
}

func (r *REST) Register(mux *http.ServeMux) {
	// public endpoint for user registration
	mux.HandleFunc("/users/register", r.handleRegisterUser)
	// protected endpoints
	mux.Handle("/spot", r.withAuth(http.HandlerFunc(r.handleSpot)))
	mux.Handle("/ohlcv", r.withAuth(http.HandlerFunc(r.handleOHLCV)))
	mux.Handle("/indicators", r.withAuth(http.HandlerFunc(r.handleIndicators)))
	mux.Handle("/bots", r.withAuth(http.HandlerFunc(r.handleCreateBot)))
	mux.Handle("/bots/perspective", r.withAuth(http.HandlerFunc(r.handleBotPerspective)))
	mux.Handle("/grants", r.withAuth(http.HandlerFunc(r.handleCreateGrant)))
	mux.Handle("/wallets/create", r.withAuth(http.HandlerFunc(r.handleCreateWallet)))
	mux.Handle("/apikeys/issue", r.withAuth(http.HandlerFunc(r.handleIssueAPIKey)))
	mux.Handle("/apikeys/rotate", r.withAuth(http.HandlerFunc(r.handleRotateAPIKey)))
	mux.Handle("/apikeys/revoke", r.withAuth(http.HandlerFunc(r.handleRevokeAPIKey)))
	mux.Handle("/audits", r.withAuth(http.HandlerFunc(r.handleListAudits)))
}

func (r *REST) handleSpot(w http.ResponseWriter, req *http.Request) {
	sym := req.URL.Query().Get("symbol")
	if sym == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	p, err := r.svc.Latest(req.Context(), sym)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"symbol": p.Symbol, "price": p.Value, "source": p.Source, "ts": p.Timestamp.Unix()})
}

func (r *REST) handleOHLCV(w http.ResponseWriter, req *http.Request) {
	sym := req.URL.Query().Get("symbol")
	tf := Timeframe(req.URL.Query().Get("tf"))
	if tf == "" {
		tf = TF1m
	}
	if sym == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	bars, err := r.svc.OHLCV(req.Context(), sym, tf, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	_ = json.NewEncoder(w).Encode(bars)
}

func (r *REST) handleIndicators(w http.ResponseWriter, req *http.Request) {
	sym := req.URL.Query().Get("symbol")
	tf := Timeframe(req.URL.Query().Get("tf"))
	if tf == "" {
		tf = TF1m
	}
	var maWins, emaWins []int
	if ma := strings.TrimSpace(req.URL.Query().Get("ma")); ma != "" {
		for _, p := range strings.Split(ma, ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				maWins = append(maWins, n)
			}
		}
	}
	if ema := strings.TrimSpace(req.URL.Query().Get("ema")); ema != "" {
		for _, p := range strings.Split(ema, ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				emaWins = append(emaWins, n)
			}
		}
	}
	out, err := r.svc.Indicators(req.Context(), sym, tf, maWins, emaWins)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

// withAuth enforces X-API-Key header unless repository is nil (dev mode)
func (r *REST) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !r.checkCORS(w, req) {
			return
		}
		if r.repo == nil {
			next.ServeHTTP(w, req)
			return
		}
		apiKey := strings.TrimSpace(req.Header.Get("X-API-Key"))
		if apiKey == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}
		if r.limiter != nil && !r.limiter.Allow(apiKey) {
			r.logAudit(nil, nil, nil, "rate.limit", map[string]any{"key": apiKey, "path": req.URL.Path})
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		rec, err := r.repo.VerifyAPIKey(apiKey)
		if err != nil {
			http.Error(w, "invalid api key", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(req.Context(), apiKeyRecordKey, rec)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *REST) checkCORS(w http.ResponseWriter, req *http.Request) bool {
	if len(r.corsOrigins) == 0 {
		return true
	}
	origin := strings.TrimSpace(req.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	allowed := false
	for _, allowedOrigin := range r.corsOrigins {
		if origin == allowedOrigin {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return false
	}
	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-API-Key")
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	return true
}

type registerReq struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type registerResp struct {
	APIKey string `json:"api_key"`
}

func (r *REST) handleRegisterUser(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		if r.svc != nil {
			repoFromSvc := r.svc.Repository()
			if repoFromSvc != nil {
				r.repo = repoFromSvc
			}
		}
	}
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in registerReq
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil || strings.TrimSpace(in.Email) == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	uid, err := r.repo.CreateUser(strings.TrimSpace(in.Email), strings.TrimSpace(in.Name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := r.repo.CreateAPIKey(uid, ScopeUserAdmin, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	keyRec, err := r.repo.VerifyAPIKey(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.logAudit(&uid, keyRec.BotID, &keyRec.ID, "user.register", map[string]any{
		"email":          strings.TrimSpace(in.Email),
		"name":           strings.TrimSpace(in.Name),
		"api_key_prefix": keyRec.Prefix,
		"api_key_scope":  keyRec.Scope,
		"api_key_id":     keyRec.ID,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(registerResp{APIKey: key})
}

type createBotReq struct {
	Name       string `json:"name"`
	ConfigJSON string `json:"config_json"`
}

type perspectiveReq struct {
	BotID       int64  `json:"bot_id"`
	PreTF       string `json:"pre_tf"`
	PreLimit    int    `json:"pre_limit"`
	TargetTF    string `json:"target_tf"`
	TargetLimit int    `json:"target_limit"`
	PostTF      string `json:"post_tf"`
	PostLimit   int    `json:"post_limit"`
}

type perspectiveResp struct {
	BotID       int64     `json:"bot_id"`
	PreTF       Timeframe `json:"pre_tf"`
	PreLimit    int       `json:"pre_limit"`
	TargetTF    Timeframe `json:"target_tf"`
	TargetLimit int       `json:"target_limit"`
	PostTF      Timeframe `json:"post_tf"`
	PostLimit   int       `json:"post_limit"`
}

func (r *REST) handleCreateBot(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var in createBotReq
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil || strings.TrimSpace(in.Name) == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	configPayload := strings.TrimSpace(in.ConfigJSON)
	if configPayload != "" {
		if cfg, err := ParsePerspective(configPayload); err == nil {
			if normalized, err := cfg.Marshal(); err == nil {
				configPayload = normalized
			}
		} else if defJSON, derr := DefaultPerspectiveConfig().Marshal(); derr == nil {
			configPayload = defJSON
		}
	} else if defJSON, err := DefaultPerspectiveConfig().Marshal(); err == nil {
		configPayload = defJSON
	}
	id, err := r.repo.CreateBot(rec.UserID, strings.TrimSpace(in.Name), configPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.logAudit(&rec.UserID, &id, &rec.ID, "bot.create", map[string]any{
		"bot_id":     id,
		"bot_name":   strings.TrimSpace(in.Name),
		"has_config": strings.TrimSpace(in.ConfigJSON) != "",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"bot_id": id})
}

func (r *REST) handleBotPerspective(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch req.Method {
	case http.MethodGet:
		botIDStr := strings.TrimSpace(req.URL.Query().Get("bot_id"))
		if botIDStr == "" {
			http.Error(w, "bot_id required", http.StatusBadRequest)
			return
		}
		botID, err := strconv.ParseInt(botIDStr, 10, 64)
		if err != nil || botID <= 0 {
			http.Error(w, "invalid bot_id", http.StatusBadRequest)
			return
		}
		bot, err := r.repo.GetBot(botID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "bot not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if bot.UserID != rec.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		cfg, err := ParsePerspective(bot.ConfigJSON)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(perspectiveResp{
			BotID:       bot.ID,
			PreTF:       cfg.PreTF,
			PreLimit:    cfg.PreLimit,
			TargetTF:    cfg.TargetTF,
			TargetLimit: cfg.TargetLimit,
			PostTF:      cfg.PostTF,
			PostLimit:   cfg.PostLimit,
		})
	case http.MethodPut:
		var in perspectiveReq
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if in.BotID <= 0 {
			http.Error(w, "bot_id required", http.StatusBadRequest)
			return
		}
		bot, err := r.repo.GetBot(in.BotID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "bot not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if bot.UserID != rec.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		cfg := PerspectiveConfig{
			PreTF:       Timeframe(strings.TrimSpace(in.PreTF)),
			PreLimit:    in.PreLimit,
			TargetTF:    Timeframe(strings.TrimSpace(in.TargetTF)),
			TargetLimit: in.TargetLimit,
			PostTF:      Timeframe(strings.TrimSpace(in.PostTF)),
			PostLimit:   in.PostLimit,
		}
		if err := cfg.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		payload, err := cfg.Marshal()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.repo.UpdateBotConfig(in.BotID, payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		botID := in.BotID
		if err := r.logAudit(&rec.UserID, &botID, &rec.ID, "bot.perspective.update", map[string]any{
			"bot_id": in.BotID,
			"config": cfg,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(perspectiveResp{
			BotID:       in.BotID,
			PreTF:       cfg.PreTF,
			PreLimit:    cfg.PreLimit,
			TargetTF:    cfg.TargetTF,
			TargetLimit: cfg.TargetLimit,
			PostTF:      cfg.PostTF,
			PostLimit:   cfg.PostLimit,
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type createGrantReq struct {
	BotID          int64    `json:"bot_id"`
	GrantType      string   `json:"grant_type"`
	ParamsJSON     string   `json:"params_json"`
	GranteeAddress string   `json:"grantee_address"`
	ExpiresInHours int      `json:"expires_in_hours"`
	IncludeBonding bool     `json:"include_bonding"`
	CreateFeegrant bool     `json:"create_feegrant"`
	FeeSpendLimit  string   `json:"fee_spend_limit"`
	FeeAllowedMsgs []string `json:"fee_allowed_msgs"`
}

func (r *REST) handleCreateGrant(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var in createGrantReq
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil || in.BotID <= 0 || strings.TrimSpace(in.GrantType) == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	bot, err := r.repo.GetBot(in.BotID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "bot not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if bot.UserID != rec.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	grantType := strings.TrimSpace(in.GrantType)
	switch {
	case strings.EqualFold(grantType, GrantTypeAuthzTrade):
		if r.authzGrantor == nil {
			http.Error(w, "authz client not configured", http.StatusServiceUnavailable)
			return
		}
		granteeAddr := strings.TrimSpace(in.GranteeAddress)
		if granteeAddr == "" {
			http.Error(w, "grantee_address required", http.StatusBadRequest)
			return
		}
		granterAddr, _, err := r.repo.LoadWallet(rec.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "wallet not found", http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		expires := in.ExpiresInHours
		if expires <= 0 {
			expires = 24 * 30
		}
		expiration := time.Now().UTC().Add(time.Duration(expires) * time.Hour)
		messages, err := r.authzGrantor.GrantTradingAuthz(req.Context(), granterAddr, granteeAddr, expiration, in.IncludeBonding)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		params := map[string]any{
			"grantee_address": granteeAddr,
			"expires_at":      expiration.UTC().Format(time.RFC3339),
			"include_bonding": in.IncludeBonding,
			"messages":        messages,
		}
		grantType = GrantTypeAuthzTrade
		var feeInfo map[string]any
		if (in.CreateFeegrant || strings.TrimSpace(in.FeeSpendLimit) != "" || len(in.FeeAllowedMsgs) > 0) && r.feeGrantor != nil {
			feeRes, err := r.feeGrantor.GrantFeeAllowance(req.Context(), granterAddr, granteeAddr, strings.TrimSpace(in.FeeSpendLimit), expiration, in.FeeAllowedMsgs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			feeInfo = map[string]any{
				"spend_limit":      feeRes.SpendLimit,
				"allowed_messages": feeRes.AllowedMessages,
				"expires_at":       feeRes.Expiration.UTC().Format(time.RFC3339),
			}
		} else if in.CreateFeegrant && r.feeGrantor == nil {
			http.Error(w, "feegrant client not configured", http.StatusServiceUnavailable)
			return
		}
		if feeInfo != nil {
			params["fee_grant"] = feeInfo
		}
		payload, err := json.Marshal(params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, err := r.repo.CreateGrant(in.BotID, grantType, string(payload), &expiration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		botID := in.BotID
		auditMeta := map[string]any{
			"grant_id":        id,
			"bot_id":          in.BotID,
			"grant_type":      grantType,
			"grantee_address": granteeAddr,
			"messages":        messages,
			"expires_at":      expiration.UTC().Format(time.RFC3339),
		}
		if feeInfo != nil {
			auditMeta["fee_grant"] = feeInfo
		}
		if err := r.logAudit(&rec.UserID, &botID, &rec.ID, "grant.create", auditMeta); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := map[string]any{
			"grant_id":   id,
			"messages":   messages,
			"expires_at": expiration.UTC().Format(time.RFC3339),
		}
		if feeInfo != nil {
			resp["fee_grant"] = feeInfo
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	default:
		id, err := r.repo.CreateGrant(in.BotID, grantType, strings.TrimSpace(in.ParamsJSON), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		botID := in.BotID
		if err := r.logAudit(&rec.UserID, &botID, &rec.ID, "grant.create", map[string]any{
			"grant_id":   id,
			"bot_id":     in.BotID,
			"grant_type": grantType,
			"params":     strings.TrimSpace(in.ParamsJSON),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"grant_id": id})
	}
}

type createWalletReq struct {
	Address string `json:"address"`
	PrivKey string `json:"priv_base64"`
}

func (r *REST) handleCreateWallet(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var in createWalletReq
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil || strings.TrimSpace(in.Address) == "" || strings.TrimSpace(in.PrivKey) == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	pk, err := base64.StdEncoding.DecodeString(in.PrivKey)
	if err != nil {
		http.Error(w, "invalid priv_base64", http.StatusBadRequest)
		return
	}
	addr := strings.TrimSpace(in.Address)
	if err := r.repo.StoreWallet(rec.UserID, addr, pk); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.logAudit(&rec.UserID, nil, &rec.ID, "wallet.store", map[string]any{
		"address": addr,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (r *REST) handleIssueAPIKey(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var in issueKeyReq
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	scope := strings.TrimSpace(in.Scope)
	if scope == "" {
		scope = ScopeUserAdmin
	}
	var botPtr *int64
	if in.BotID != nil {
		if *in.BotID <= 0 {
			http.Error(w, "bot_id must be > 0", http.StatusBadRequest)
			return
		}
		botPtr = in.BotID
	}
	key, err := r.repo.CreateAPIKey(rec.UserID, scope, botPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	issuedRec, err := r.repo.VerifyAPIKey(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prefix := key
	if len(prefix) > 8 {
		prefix = key[:8]
	}
	if err := r.logAudit(&rec.UserID, issuedRec.BotID, &rec.ID, "api_key.issue", map[string]any{
		"scope":              issuedRec.Scope,
		"new_api_key_id":     issuedRec.ID,
		"new_api_key_prefix": issuedRec.Prefix,
		"requested_scope":    scope,
		"bot_id":             issuedRec.BotID,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(issueKeyResp{APIKey: key, Prefix: prefix, Scope: scope})
}

type rotateKeyReq struct {
	OldPrefix string `json:"old_prefix"`
}

type issueKeyReq struct {
	Scope string `json:"scope"`
	BotID *int64 `json:"bot_id"`
}

type issueKeyResp struct {
	APIKey string `json:"api_key"`
	Prefix string `json:"prefix"`
	Scope  string `json:"scope"`
}

func (r *REST) handleRotateAPIKey(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var in rotateKeyReq
	_ = json.NewDecoder(req.Body).Decode(&in)
	oldPrefix := strings.TrimSpace(in.OldPrefix)
	newKey, err := r.repo.RotateAPIKey(rec.UserID, oldPrefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rotRec, err := r.repo.VerifyAPIKey(newKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.logAudit(&rec.UserID, rotRec.BotID, &rec.ID, "api_key.rotate", map[string]any{
		"old_prefix":         oldPrefix,
		"new_api_key_id":     rotRec.ID,
		"new_api_key_prefix": rotRec.Prefix,
		"scope":              rotRec.Scope,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"api_key": newKey})
}

type revokeKeyReq struct {
	Prefix string `json:"prefix"`
}

func (r *REST) handleRevokeAPIKey(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var in revokeKeyReq
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil || strings.TrimSpace(in.Prefix) == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	prefix := strings.TrimSpace(in.Prefix)
	target, err := r.repo.GetAPIKeyByPrefix(rec.UserID, prefix)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "api key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := r.repo.RevokeAPIKey(rec.UserID, prefix); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "api key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	if err := r.logAudit(&rec.UserID, target.BotID, &rec.ID, "api_key.revoke", map[string]any{
		"prefix":             prefix,
		"revoked_api_key_id": target.ID,
		"scope":              target.Scope,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "revoked"})
}

func (r *REST) handleListAudits(w http.ResponseWriter, req *http.Request) {
	if r.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rec, ok := apiKeyRecordFromContext(req.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !requireScope(rec, ScopeUserAdmin, ScopeUserRead) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	limit := 50
	if q := strings.TrimSpace(req.URL.Query().Get("limit")); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil || v <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if v > 500 {
			v = 500
		}
		limit = v
	}
	rows, err := r.repo.ListAudits(rec.UserID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]auditEntry, 0, len(rows))
	for _, row := range rows {
		meta := strings.TrimSpace(row.MetaJSON)
		if meta == "" {
			meta = "{}"
		}
		raw := json.RawMessage([]byte(meta))
		resp = append(resp, auditEntry{
			ID:       row.ID,
			UserID:   row.UserID,
			BotID:    row.BotID,
			APIKeyID: row.APIKeyID,
			Action:   row.Action,
			Meta:     raw,
			Created:  row.CreatedAt,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}
