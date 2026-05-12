package classifier

import (
	"regexp"
	"strings"

	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

// Priority indicates the revenue importance of a link.
type Priority int

const (
	PriorityNormal  Priority = 0
	PriorityRevenue Priority = 1
)

// revenueDomains are platforms commonly used to sell products, courses, etc.
var revenueDomains = []string{
	"gumroad.com",
	"teachable.com",
	"thinkific.com",
	"shopify.com",
	"etsy.com",
	"patreon.com",
	"ko-fi.com",
	"buymeacoffee.com",
	"memberful.com",
	"podia.com",
	"clickfunnels.com",
	"convertkit.com",
	"cart", // cart.co, shop.app/cart, etc.
}

var revenueURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`/(product|products|course|courses)/`),
	regexp.MustCompile(`/(checkout|cart|purchase|buy|order)`),
	regexp.MustCompile(`/(pricing|membership|subscribe|subscription)`),
	regexp.MustCompile(`/(donate|donation|support|tip)`),
	regexp.MustCompile(`/(enroll|enrol|register|signup|apply)`),
	regexp.MustCompile(`/(download|downloads)`),
	regexp.MustCompile(`/(shop|store|storefront)`),
	regexp.MustCompile(`/(bundle|bundles|pack|packs|template|templates|ebook|ebooks)`),
	regexp.MustCompile(`/(sale|deals|offer|discount|coupon)`),
	regexp.MustCompile(`/dp/`),           // Amazon product detail page
	regexp.MustCompile(`/gp/product/`),   // Amazon product page
	regexp.MustCompile(`aff_id=`),        // affiliate ID parameter
	regexp.MustCompile(`[?&]tag=`),       // Amazon affiliate tag
	regexp.MustCompile(`[?&]ref=`),       // referral/affiliate ref
	regexp.MustCompile(`affiliate`),
	regexp.MustCompile(`referral`),
	regexp.MustCompile(`/premium`),
	regexp.MustCompile(`/pro/`),
}

// revenueKeywords are matched against the surrounding context text (case-insensitive).
var revenueKeywords = []string{
	"buy", "purchase", "checkout", "cart",
	"course", "enroll", "enrollment", "enrol", "class", "curriculum",
	"product", "digital", "download", "bundle", "pack",
	"shop", "store", "order",
	"membership", "premium", "pro", "exclusive",
	"sale", "offer", "deal", "discount", "coupon",
	"affiliate", "referral", "commission",
	"patreon", "ko-fi", "donation", "donate", "support", "tip",
	"template", "ebook", "pdf", "software", "tool", "app",
	"subscribe", "subscription",
	"pricing", "price",
}

// Classify determines whether a link is revenue-critical based on
// its URL structure and surrounding context text.
func Classify(link youtube.ScrapedLink) Priority {
	if matchURL(link.URL) {
		return PriorityRevenue
	}
	if matchContext(link.Context) {
		return PriorityRevenue
	}
	return PriorityNormal
}

func matchURL(raw string) bool {
	lower := strings.ToLower(raw)
	for _, domain := range revenueDomains {
		if strings.Contains(lower, domain) {
			return true
		}
	}
	for _, pat := range revenueURLPatterns {
		if pat.MatchString(lower) {
			return true
		}
	}
	return false
}

func matchContext(ctx string) bool {
	lower := strings.ToLower(ctx)
	for _, kw := range revenueKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
