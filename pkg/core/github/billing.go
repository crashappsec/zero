// Package github provides GitHub API interactions
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ============================================================================
// Billing API Types
// ============================================================================

// BillingActions represents GitHub Actions billing data
type BillingActions struct {
	TotalMinutesUsed     int            `json:"total_minutes_used"`
	TotalPaidMinutesUsed int            `json:"total_paid_minutes_used"`
	IncludedMinutes      int            `json:"included_minutes"`
	MinutesBreakdown     map[string]int `json:"minutes_used_breakdown"`
}

// BillingPackages represents GitHub Packages billing data
type BillingPackages struct {
	TotalGigabytesBandwidthUsed     int `json:"total_gigabytes_bandwidth_used"`
	TotalPaidGigabytesBandwidthUsed int `json:"total_paid_gigabytes_bandwidth_used"`
	IncludedGigabytesBandwidth      int `json:"included_gigabytes_bandwidth"`
}

// BillingStorage represents shared storage billing data
type BillingStorage struct {
	DaysLeftInBillingCycle       int `json:"days_left_in_billing_cycle"`
	EstimatedPaidStorageForMonth int `json:"estimated_paid_storage_for_month"`
	EstimatedStorageForMonth     int `json:"estimated_storage_for_month"`
}

// BillingSummary combines all billing data for an organization
type BillingSummary struct {
	Owner    string          `json:"owner"`
	Actions  *BillingActions `json:"actions,omitempty"`
	Packages *BillingPackages `json:"packages,omitempty"`
	Storage  *BillingStorage `json:"storage,omitempty"`
	// Computed cost estimates based on GitHub pricing
	EstimatedMonthlyCost *CostEstimate `json:"estimated_monthly_cost,omitempty"`
}

// CostEstimate provides cost breakdown based on GitHub pricing
type CostEstimate struct {
	ActionsCost  float64 `json:"actions_cost"`
	StorageCost  float64 `json:"storage_cost"`
	BandwidthCost float64 `json:"bandwidth_cost"`
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
	Note         string  `json:"note"`
}

// GitHub Actions pricing per minute (as of 2024)
// https://docs.github.com/en/billing/managing-billing-for-github-actions/about-billing-for-github-actions
var actionsPricing = map[string]float64{
	"UBUNTU":        0.008,  // Linux
	"MACOS":         0.08,   // macOS
	"WINDOWS":       0.016,  // Windows
	"ubuntu":        0.008,
	"macos":         0.08,
	"windows":       0.016,
	"UBUNTU_2_CORE": 0.008,
	"UBUNTU_4_CORE": 0.016,
	"MACOS_LARGE":   0.12,
}

// ============================================================================
// Billing API Methods
// ============================================================================

// GetOrgActionsBilling fetches GitHub Actions billing for an organization
// Requires: admin:org scope (read:org is NOT sufficient)
func (c *Client) GetOrgActionsBilling(org string) (*BillingActions, error) {
	if !c.HasToken() {
		return nil, fmt.Errorf("no GitHub token available - set GITHUB_TOKEN with admin:org scope")
	}

	url := fmt.Sprintf("https://api.github.com/orgs/%s/settings/billing/actions", org)
	return c.fetchBillingActions(url)
}

// GetUserActionsBilling fetches GitHub Actions billing for a user
// Requires: user scope
func (c *Client) GetUserActionsBilling(username string) (*BillingActions, error) {
	if !c.HasToken() {
		return nil, fmt.Errorf("no GitHub token available")
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/settings/billing/actions", username)
	return c.fetchBillingActions(url)
}

// GetOrgPackagesBilling fetches GitHub Packages billing for an organization
// Requires: admin:org scope (read:org is NOT sufficient)
func (c *Client) GetOrgPackagesBilling(org string) (*BillingPackages, error) {
	if !c.HasToken() {
		return nil, fmt.Errorf("no GitHub token available - set GITHUB_TOKEN with admin:org scope")
	}

	url := fmt.Sprintf("https://api.github.com/orgs/%s/settings/billing/packages", org)
	return c.fetchBillingPackages(url)
}

// GetOrgStorageBilling fetches shared storage billing for an organization
// Requires: admin:org scope (read:org is NOT sufficient)
func (c *Client) GetOrgStorageBilling(org string) (*BillingStorage, error) {
	if !c.HasToken() {
		return nil, fmt.Errorf("no GitHub token available - set GITHUB_TOKEN with admin:org scope")
	}

	url := fmt.Sprintf("https://api.github.com/orgs/%s/settings/billing/shared-storage", org)
	return c.fetchBillingStorage(url)
}

// GetOrgBillingSummary fetches all billing data for an organization
func (c *Client) GetOrgBillingSummary(org string) (*BillingSummary, error) {
	summary := &BillingSummary{
		Owner: org,
	}

	// Fetch Actions billing
	actions, err := c.GetOrgActionsBilling(org)
	if err == nil {
		summary.Actions = actions
	}

	// Fetch Packages billing
	packages, err := c.GetOrgPackagesBilling(org)
	if err == nil {
		summary.Packages = packages
	}

	// Fetch Storage billing
	storage, err := c.GetOrgStorageBilling(org)
	if err == nil {
		summary.Storage = storage
	}

	// Calculate estimated costs
	summary.EstimatedMonthlyCost = calculateCostEstimate(summary)

	return summary, nil
}

// ============================================================================
// Internal Helper Methods
// ============================================================================

func (c *Client) fetchBillingActions(url string) (*BillingActions, error) {
	body, err := c.doGitHubRequest(url)
	if err != nil {
		return nil, err
	}

	var billing BillingActions
	if err := json.Unmarshal(body, &billing); err != nil {
		return nil, fmt.Errorf("parsing billing response: %w", err)
	}

	return &billing, nil
}

func (c *Client) fetchBillingPackages(url string) (*BillingPackages, error) {
	body, err := c.doGitHubRequest(url)
	if err != nil {
		return nil, err
	}

	var billing BillingPackages
	if err := json.Unmarshal(body, &billing); err != nil {
		return nil, fmt.Errorf("parsing billing response: %w", err)
	}

	return &billing, nil
}

func (c *Client) fetchBillingStorage(url string) (*BillingStorage, error) {
	body, err := c.doGitHubRequest(url)
	if err != nil {
		return nil, err
	}

	var billing BillingStorage
	if err := json.Unmarshal(body, &billing); err != nil {
		return nil, fmt.Errorf("parsing billing response: %w", err)
	}

	return &billing, nil
}

func (c *Client) doGitHubRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authentication failed - token may be expired or invalid")
	}
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("access forbidden - token needs admin:org scope for billing data")
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("billing data not found - requires admin:org scope. Add it with: gh auth refresh -s admin:org")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// calculateCostEstimate computes estimated costs based on GitHub pricing
func calculateCostEstimate(summary *BillingSummary) *CostEstimate {
	estimate := &CostEstimate{
		Currency: "USD",
		Note:     "Estimated based on GitHub's published pricing. Actual costs may vary.",
	}

	// Calculate Actions cost from minutes breakdown
	if summary.Actions != nil && summary.Actions.MinutesBreakdown != nil {
		for runner, minutes := range summary.Actions.MinutesBreakdown {
			if price, ok := actionsPricing[runner]; ok {
				estimate.ActionsCost += float64(minutes) * price
			} else {
				// Default to Linux pricing for unknown runners
				estimate.ActionsCost += float64(minutes) * 0.008
			}
		}
	}

	// Storage cost: $0.25 per GB per month
	if summary.Storage != nil {
		estimate.StorageCost = float64(summary.Storage.EstimatedStorageForMonth) * 0.25
	}

	// Bandwidth cost: $0.50 per GB (data transfer out)
	if summary.Packages != nil {
		// Only paid bandwidth counts
		estimate.BandwidthCost = float64(summary.Packages.TotalPaidGigabytesBandwidthUsed) * 0.50
	}

	estimate.TotalCost = estimate.ActionsCost + estimate.StorageCost + estimate.BandwidthCost

	return estimate
}
