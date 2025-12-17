package featureflag

import (
	"context"
	"log/slog"

	flipt "github.com/open-feature/go-sdk-contrib/providers/flipt/pkg/provider"
	"github.com/open-feature/go-sdk/openfeature"
)

func InitFeatureFlag(fliptURL string) error {
	// Set the provider as the default for OpenFeature
	err := openfeature.SetProvider(flipt.NewProvider(flipt.WithAddress(fliptURL)))
	if err != nil {
		return err
	}

	// Create a new client
	client := openfeature.NewClient("golang-core-template")

	// Test the connection
	_, err = client.BooleanValue(context.Background(), "ping-pong", false, openfeature.EvaluationContext{})
	if err != nil {
		return err
	}

	slog.Info("Feature flag system initialized successfully")
	return nil
}
