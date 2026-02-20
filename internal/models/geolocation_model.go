package models

import (
	"errors"
	"fmt"
)

type GeolocationConfiguration struct {
	Enabled           bool               `json:"enabled" validate:"required"`
	Latitude          float64            `json:"latitude" validate:"required"`
	Longitude         float64            `json:"longitude" validate:"required"`
	Radius            float64            `json:"radius" validate:"required"`
	QRValidationRules []QRValidationRule `json:"qrValidationRules" validate:"omitempty,dive"`
	ErrorAction       string             `json:"errorAction" validate:"required,oneof=reject allow warn"`
}

// Validation errors
var (
	ErrInvalidLatitude    = errors.New("latitude must be between -90 and 90 degrees")
	ErrInvalidLongitude   = errors.New("longitude must be between -180 and 180 degrees")
	ErrInvalidRadius      = errors.New("radius must be greater than 0 meters")
	ErrInvalidErrorAction = errors.New("errorAction must be one of: reject, allow, warn")
	ErrDuplicateQRType    = errors.New("duplicate QR type in validation rules")
	ErrInvalidQRType      = errors.New("invalid QR type")
)

// Validate performs comprehensive validation on GeolocationConfiguration
// This goes beyond struct tag validation to ensure business logic correctness
func (g *GeolocationConfiguration) Validate() error {
	// If disabled, no further validation needed
	if !g.Enabled {
		return nil
	}

	// Validate latitude range (-90 to +90)
	if g.Latitude < -90 || g.Latitude > 90 {
		return fmt.Errorf("%w: got %.6f", ErrInvalidLatitude, g.Latitude)
	}

	// Validate longitude range (-180 to +180)
	if g.Longitude < -180 || g.Longitude > 180 {
		return fmt.Errorf("%w: got %.6f", ErrInvalidLongitude, g.Longitude)
	}

	// Validate radius (must be positive and reasonable)
	if g.Radius <= 0 {
		return fmt.Errorf("%w: got %.2f", ErrInvalidRadius, g.Radius)
	}

	// Warn about unreasonably large radius (> 10km = 10,000 meters)
	// This is a soft limit - still valid but potentially unintended
	const maxReasonableRadius = 10000.0
	if g.Radius > maxReasonableRadius {
		// Note: This doesn't return an error, just validates it's intentional
		// You might want to log a warning here in production
	}

	// Validate error action (already validated by struct tag, but double-check)
	validActions := map[string]bool{
		"reject": true,
		"allow":  true,
		"warn":   true,
	}
	if !validActions[g.ErrorAction] {
		return fmt.Errorf("%w: got %s", ErrInvalidErrorAction, g.ErrorAction)
	}

	// Validate QR validation rules if provided
	if len(g.QRValidationRules) > 0 {
		if err := g.validateQRRules(); err != nil {
			return err
		}
	}

	return nil
}

// validateQRRules validates the QR validation rules for logical consistency
func (g *GeolocationConfiguration) validateQRRules() error {
	qrTypeSeen := make(map[string]bool)
	hasRequiredRule := false

	for i, rule := range g.QRValidationRules {
		// Check for duplicate QR types
		if qrTypeSeen[rule.QRType] {
			return fmt.Errorf("%w: %s", ErrDuplicateQRType, rule.QRType)
		}
		qrTypeSeen[rule.QRType] = true

		// Validate QR type (already done by struct tag, but explicit check)
		validQRTypes := map[string]bool{
			"personal-qr":     true,
			"event-qr":        true,
			"session-qr":      true,
			"registration-qr": true,
		}
		if !validQRTypes[rule.QRType] {
			return fmt.Errorf("%w at index %d: %s", ErrInvalidQRType, i, rule.QRType)
		}

		// Track if any rule is required
		if rule.Required {
			hasRequiredRule = true
		}

		// Logical validation: if required=true, allowOverride should typically be false
		// This is a soft warning - not an error, but potentially inconsistent
		if rule.Required && rule.AllowOverride {
			// This is allowed but might be worth logging as a warning
			// "Required" means location check is mandatory
			// "AllowOverride" means user can bypass it
			// Having both true is contradictory but we'll allow it
		}
	}

	// If geolocation is enabled but no rules are required, it's essentially optional
	// This is valid but might be worth noting
	if !hasRequiredRule {
		// All rules are optional - geolocation check is soft
		// This is a valid configuration
	}

	return nil
}

// IsStrictMode returns true if geolocation is enabled with reject error action
func (g *GeolocationConfiguration) IsStrictMode() bool {
	return g.Enabled && g.ErrorAction == "reject"
}

// IsLenientMode returns true if geolocation is enabled but allows ignoring errors
func (g *GeolocationConfiguration) IsLenientMode() bool {
	return g.Enabled && (g.ErrorAction == "allow" || g.ErrorAction == "warn")
}

// HasRequiredQRType checks if a specific QR type requires geolocation validation
func (g *GeolocationConfiguration) HasRequiredQRType(qrType string) bool {
	for _, rule := range g.QRValidationRules {
		if rule.QRType == qrType && rule.Required {
			return true
		}
	}
	return false
}

// CanOverrideForQRType checks if geolocation can be overridden for a specific QR type
func (g *GeolocationConfiguration) CanOverrideForQRType(qrType string) bool {
	for _, rule := range g.QRValidationRules {
		if rule.QRType == qrType {
			return rule.AllowOverride
		}
	}
	// If no rule exists for this QR type, default to not allowing override
	return false
}

// GetRadiusInKilometers returns the radius in kilometers for display purposes
func (g *GeolocationConfiguration) GetRadiusInKilometers() float64 {
	return g.Radius / 1000.0
}

// String provides a human-readable representation
func (g *GeolocationConfiguration) String() string {
	if !g.Enabled {
		return "Geolocation: Disabled"
	}
	return fmt.Sprintf("Geolocation: Enabled (%.6f, %.6f) radius: %.0fm, action: %s",
		g.Latitude, g.Longitude, g.Radius, g.ErrorAction)
}
