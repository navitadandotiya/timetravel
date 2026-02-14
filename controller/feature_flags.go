package controller

import (
	"github.com/rainbowmga/timetravel/service"
)

// FeatureFlagController handles runtime feature flags
type FeatureFlagController struct {
	service *service.FeatureFlagService
}

// NewFeatureFlagController initializes and loads all flags
func NewFeatureFlagController(dbPath string) (*FeatureFlagController, error) {

	svc, err := service.NewFeatureFlagService(dbPath)
	if err != nil {
		return nil, err
	}

	return &FeatureFlagController{service: svc}, nil
}



// IsEnabled returns true if the flag is enabled
func (c *FeatureFlagController) IsEnabled(flag string) bool {
	return c.service.IsEnabled(flag);
}

// Refresh reloads flags from the DB at runtime
func (c *FeatureFlagController) Refresh() error {
	return c.service.Refresh()
}
