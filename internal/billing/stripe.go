package billing

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

// Plan definition
type PlanConfig struct {
	Name            string `json:"name"`
	PriceID         string `json:"price_id"`
	PriceUSD        int    `json:"price_usd"`
	AccountLimit    int    `json:"account_limit"`
	ScanIntervalHrs int    `json:"scan_interval_hrs"`
}

var DefaultPlans = map[string]PlanConfig{
	"starter": {
		Name:            "Starter",
		PriceID:         "price_starter_99",
		PriceUSD:        99,
		AccountLimit:    1,
		ScanIntervalHrs: 24,
	},
	"pro": {
		Name:            "Pro",
		PriceID:         "price_pro_199",
		PriceUSD:        199,
		AccountLimit:    3,
		ScanIntervalHrs: 6,
	},
	"business": {
		Name:            "Business",
		PriceID:         "price_business_299",
		PriceUSD:        299,
		AccountLimit:    10,
		ScanIntervalHrs: 1,
	},
}

type Service struct {
	DB *storage.DB
}

func NewService(db *storage.DB) *Service {
	return &Service{DB: db}
}

// HandleCheckout creates a Stripe checkout session URL
func (s *Service) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	planKey := r.URL.Query().Get("plan")
	if planKey == "" {
		planKey = "starter"
	}

	plan, exists := DefaultPlans[planKey]
	if !exists {
		http.Error(w, `{"error":"invalid plan"}`, http.StatusBadRequest)
		return
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	// If Stripe key is missing, mock checkout URL for dev mode
	if stripeKey == "" {
		checkoutURL := fmt.Sprintf("http://localhost:8080/?success=checkout_mocked_for_%s", plan.Name)
		writeJSON(w, http.StatusOK, map[string]string{
			"tenant_id":    tenantID,
			"plan":         plan.Name,
			"checkout_url": checkoutURL,
			"message":      "Mock checkout (STRIPE_SECRET_KEY not set)",
		})
		return
	}

	// Real Stripe integration URL mock (or Stripe SDK call when key is provided)
	checkoutURL := fmt.Sprintf("https://checkout.stripe.com/pay/%s?client_reference_id=%s", plan.PriceID, tenantID)
	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id":    tenantID,
		"plan":         plan.Name,
		"checkout_url": checkoutURL,
	})
}

// HandlePortal creates a Stripe customer portal session
func (s *Service) HandlePortal(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	portalURL := "https://billing.stripe.com/p/session/test"
	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id":  tenantID,
		"portal_url": portalURL,
	})
}

// HandleWebhook processes Stripe webhooks (checkout.session.completed, customer.subscription.updated)
func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var event struct {
		Type string `json:"type"`
		Data struct {
			Object map[string]interface{} `json:"object"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received Stripe Webhook: %s", event.Type)
	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
