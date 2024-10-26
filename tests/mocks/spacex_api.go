package mocks

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockSpaceXClient struct {
	mock.Mock
}

func (m *MockSpaceXClient) CheckLaunchConflict(ctx context.Context, launchpadID string, ts time.Time) (bool, error) {
	args := m.Called(ctx, launchpadID, ts)
	return args.Bool(0), args.Error(1)
}

type MockSpaceXClientAvailable struct{}

func (m *MockSpaceXClientAvailable) CheckLaunchConflict(ctx context.Context, launchpadID string, date time.Time) (bool, error) {
	return true, nil
}

type MockSpaceXClientUnavailable struct{}

func (m *MockSpaceXClientUnavailable) CheckLaunchConflict(ctx context.Context, launchpadID string, date time.Time) (bool, error) {
	return false, nil
}

type MockSpaceXClientError struct{}

func (m *MockSpaceXClientError) CheckLaunchConflict(ctx context.Context, launchpadID string, date time.Time) (bool, error) {
	return false, errors.New("spaceX api error")
}
