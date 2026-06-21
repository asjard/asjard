package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestDefaultHandler(t *testing.T) {
	api := &DefaultHandlersAPI{}
	resp, err := api.Favicon(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, api.RestServiceDesc())
}

func TestHealthDescriptors(t *testing.T) {
	health := Health{}
	require.NotNil(t, health.RestServiceDesc())
	require.NotNil(t, health.GrpcServiceDesc())
}
