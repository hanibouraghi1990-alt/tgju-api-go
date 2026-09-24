package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type daraeimanPrices struct {
	Gold18    float64 `json:"gold18"`
	Silver925 float64 `json:"silver925"`
	USD       float64 `json:"usd"`
	Bitcoin   float64 `json:"bitcoin"`
	UpdatedAt string  `json:"updatedAt"`
}

type binanceTicker struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (s *Server) handleDaraeimanPrices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// طلا ۱۸ عیار
	gold, err := s.client.Item(ctx, "geram18")
	if err != nil {
		http.Error(w, "gold price unavailable", http.StatusServiceUnavailable)
		return
	}

	// نقره ۹۲۵
	silver, err := s.client.Item(ctx, "silver_925")
	if err != nil {
		http.Error(w, "silver price unavailable", http.StatusServiceUnavailable)
		return
	}

	// دلار
	usd, err := s.client.Item(ctx, "price_dollar_rl")
	if err != nil {
		http.Error(w, "usd price unavailable", http.StatusServiceUnavailable)
		return
	}

	// بیت‌کوین به دلار
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://data-api.binance.vision/api/v3/ticker/price?symbol=BTCUSDT",
		nil,
	)
	if err != nil {
		http.Error(w, "bitcoin request failed", http.StatusServiceUnavailable)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "bitcoin price unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "bitcoin price unavailable", http.StatusServiceUnavailable)
		return
	}

	var btcTicker binanceTicker

	if err := json.NewDecoder(resp.Body).Decode(&btcTicker); err != nil {
		http.Error(w, "bitcoin response invalid", http.StatusServiceUnavailable)
		return
	}

	btcUSD, err := strconv.ParseFloat(btcTicker.Price, 64)
	if err != nil {
		http.Error(w, "bitcoin price invalid", http.StatusServiceUnavailable)
		return
	}

	// دلار و طلا و نقره از TGJU به تومان هستند.
	// BTC ابتدا به USD است؛ سپس در قیمت دلارِ تومانی ضرب می‌شود.
	bitcoinToman := btcUSD * usd.Price.Toman()

	result := daraeimanPrices{
		Gold18:    gold.Price.Toman(),
		Silver925: silver.Price.Toman(),
		USD:       usd.Price.Toman(),
		Bitcoin:   bitcoinToman,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	s.cacheControl(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		return
	}
}
