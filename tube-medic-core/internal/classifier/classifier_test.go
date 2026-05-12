package classifier_test

import (
	"testing"

	"github.com/aquaflamingo/tubemedicmvp/internal/classifier"
	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

func TestClassify_urlPatterns(t *testing.T) {
	tests := []struct {
		name string
		link youtube.ScrapedLink
		want classifier.Priority
	}{
		// Revenue-critical URL patterns
		{name: "product path", link: link("https://example.com/products/widget"), want: classifier.PriorityRevenue},
		{name: "course path", link: link("https://example.com/courses/python"), want: classifier.PriorityRevenue},
		{name: "checkout", link: link("https://example.com/checkout"), want: classifier.PriorityRevenue},
		{name: "cart", link: link("https://example.com/cart"), want: classifier.PriorityRevenue},
		{name: "purchase", link: link("https://example.com/purchase"), want: classifier.PriorityRevenue},
		{name: "pricing", link: link("https://example.com/pricing"), want: classifier.PriorityRevenue},
		{name: "membership", link: link("https://example.com/membership"), want: classifier.PriorityRevenue},
		{name: "subscribe", link: link("https://example.com/subscribe"), want: classifier.PriorityRevenue},
		{name: "donate", link: link("https://example.com/donate"), want: classifier.PriorityRevenue},
		{name: "enroll", link: link("https://example.com/enroll"), want: classifier.PriorityRevenue},
		{name: "download", link: link("https://example.com/download"), want: classifier.PriorityRevenue},
		{name: "shop", link: link("https://example.com/shop"), want: classifier.PriorityRevenue},
		{name: "bundle", link: link("https://example.com/bundle"), want: classifier.PriorityRevenue},
		{name: "amazon dp", link: link("https://amazon.com/dp/B08ABC123"), want: classifier.PriorityRevenue},
		{name: "amazon gp product", link: link("https://amazon.com/gp/product/B08ABC123"), want: classifier.PriorityRevenue},
		{name: "affiliate tag", link: link("https://example.com/?tag=myaffiliate-20"), want: classifier.PriorityRevenue},
		{name: "affiliate ref", link: link("https://example.com/?ref=partner"), want: classifier.PriorityRevenue},
		{name: "affiliate word in url", link: link("https://example.com/affiliate"), want: classifier.PriorityRevenue},
		{name: "referral in url", link: link("https://example.com/referral"), want: classifier.PriorityRevenue},
		{name: "premium", link: link("https://example.com/premium"), want: classifier.PriorityRevenue},
		{name: "gumroad domain", link: link("https://gumroad.com/l/mycourse"), want: classifier.PriorityRevenue},
		{name: "teachable domain", link: link("https://mycourse.teachable.com/"), want: classifier.PriorityRevenue},
		{name: "patreon domain", link: link("https://patreon.com/creator"), want: classifier.PriorityRevenue},
		{name: "ko-fi domain", link: link("https://ko-fi.com/creator"), want: classifier.PriorityRevenue},
		{name: "buymeacoffee", link: link("https://buymeacoffee.com/creator"), want: classifier.PriorityRevenue},
		{name: "shopify domain", link: link("https://myshop.shopify.com/product"), want: classifier.PriorityRevenue},

		// Non-revenue URLs
		{name: "plain url", link: link("https://example.com/about"), want: classifier.PriorityNormal},
		{name: "social media", link: link("https://twitter.com/user"), want: classifier.PriorityNormal},
		{name: "website home", link: link("https://mysite.com"), want: classifier.PriorityNormal},
		{name: "blog post", link: link("https://example.com/blog/post-title"), want: classifier.PriorityNormal},
		{name: "youtube video", link: link("https://youtube.com/watch?v=abc123"), want: classifier.PriorityNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.Classify(tt.link)
			if got != tt.want {
				t.Errorf("Classify(%q) = %d, want %d", tt.link.URL, got, tt.want)
			}
		})
	}
}

func TestClassify_contextKeywords(t *testing.T) {
	tests := []struct {
		name string
		link youtube.ScrapedLink
		want classifier.Priority
	}{
		{
			name: "buy keyword in context",
			link: youtube.ScrapedLink{
				URL:     "https://example.com/widget",
				Context: "Check out my new course! Buy it now at the link below",
			},
			want: classifier.PriorityRevenue,
		},
		{
			name: "course keyword",
			link: youtube.ScrapedLink{
				URL:     "https://example.com/widget",
				Context: "Enroll in my Python course today",
			},
			want: classifier.PriorityRevenue,
		},
		{
			name: "digital download",
			link: youtube.ScrapedLink{
				URL:     "https://example.com/widget",
				Context: "Get the digital download of my template pack",
			},
			want: classifier.PriorityRevenue,
		},
		{
			name: "affiliate mention",
			link: youtube.ScrapedLink{
				URL:     "https://amazon.com/widget",
				Context: "This is an affiliate link that supports the channel",
			},
			want: classifier.PriorityRevenue,
		},
		{
			name: "patreon mention",
			link: youtube.ScrapedLink{
				URL:     "https://example.com/support",
				Context: "Support me on Patreon for exclusive content",
			},
			want: classifier.PriorityRevenue,
		},
		{
			name: "donation context",
			link: youtube.ScrapedLink{
				URL:     "https://paypal.me/user",
				Context: "If you'd like to make a donation to support the channel",
			},
			want: classifier.PriorityRevenue,
		},
		{
			name: "no revenue context",
			link: youtube.ScrapedLink{
				URL:     "https://example.com/about",
				Context: "Check out my website for more information about me",
			},
			want: classifier.PriorityNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.Classify(tt.link)
			if got != tt.want {
				t.Errorf("Classify(%q) = %d, want %d", tt.link.URL, got, tt.want)
			}
		})
	}
}

func link(url string) youtube.ScrapedLink {
	return youtube.ScrapedLink{URL: url}
}
