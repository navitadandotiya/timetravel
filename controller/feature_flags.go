package controller

import (
	"github.com/rainbowmga/timetravel/service"
	"github.com/rainbowmga/timetravel/common"
	"context"
	"github.com/rainbowmga/timetravel/observability"
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
func (c *FeatureFlagController) IsEnabled(ctx context.Context, flag string) bool {
	userID, ok := common.GetUserID(ctx)
	if !ok {
		observability.DefaultLogger.Error("IsEnabled failed to fetch context error")
		return false
	}
	return c.service.IsEnabled(flag,userID);
}

// Refresh reloads flags from the DB at runtime
func (c *FeatureFlagController) Refresh() error {
	return c.service.Refresh()
}
