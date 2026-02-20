package validator_test

import (
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"testing"
)

func TestBannerRequiresImages(t *testing.T) {
	tests := []struct {
		name        string
		imageLinks  []string
		bannerLink  *string
		shouldPass  bool
		description string
	}{
		{
			name:        "Empty images, empty banner - VALID",
			imageLinks:  []string{},
			bannerLink:  nil,
			shouldPass:  true,
			description: "Both empty should be valid",
		},
		{
			name:        "Empty images, empty string banner - VALID",
			imageLinks:  []string{},
			bannerLink:  stringPtr(""),
			shouldPass:  true,
			description: "Empty string banner with no images should be valid",
		},
		{
			name:        "Empty images, has banner - INVALID",
			imageLinks:  []string{},
			bannerLink:  stringPtr("https://example.com/banner.jpg"),
			shouldPass:  false,
			description: "Banner without images should fail validation",
		},
		{
			name:        "Has images, empty banner - VALID",
			imageLinks:  []string{"https://example.com/image1.jpg"},
			bannerLink:  nil,
			shouldPass:  true,
			description: "Images without banner should be valid",
		},
		{
			name:        "Has images, has banner - VALID",
			imageLinks:  []string{"https://example.com/image1.jpg"},
			bannerLink:  stringPtr("https://example.com/banner.jpg"),
			shouldPass:  true,
			description: "Both images and banner should be valid",
		},
		{
			name:        "Multiple images, has banner - VALID",
			imageLinks:  []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
			bannerLink:  stringPtr("https://example.com/banner.jpg"),
			shouldPass:  true,
			description: "Multiple images with banner should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventImages := models.EventImages{
				ImageLinks: tt.imageLinks,
				BannerLink: tt.bannerLink,
			}

			err := validator.Validate(eventImages)

			if tt.shouldPass && err != nil {
				t.Errorf("%s: Expected validation to pass, but got error: %v", tt.description, err)
			}

			if !tt.shouldPass && err == nil {
				t.Errorf("%s: Expected validation to fail, but it passed", tt.description)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
