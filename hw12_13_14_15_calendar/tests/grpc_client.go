//go:build integration

package tests

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	v1 "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
)

type GRPCClient struct {
	conn *grpc.ClientConn
}

func NewGRPCClient(connect string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(connect, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCClient{conn: conn}, nil
}

func (c *GRPCClient) Done() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *GRPCClient) client() v1.EventServiceClient {
	return v1.NewEventServiceClient(c.conn)
}

func (c *GRPCClient) CreateEvent(ctx context.Context, ownerID string, ev *v1.Event) (*v1.Event, error) {
	client := c.client()

	md := metadata.New(map[string]string{"X-Owner-Id": ownerID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.CreateEvent(
		ctx,
		&v1.CreateEventRequest{
			Event: ev,
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Event, nil
}

func (c *GRPCClient) GetDayEvents(
	ctx context.Context,
	ownerID string,
	req *v1.GetDayEventsRequest,
) ([]*v1.Event, error) {
	client := c.client()

	md := metadata.New(map[string]string{"X-Owner-Id": ownerID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.GetDayEvents(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}

func (c *GRPCClient) GetWeekEvents(
	ctx context.Context,
	ownerID string,
	req *v1.GetWeekEventsRequest,
) ([]*v1.Event, error) {
	client := c.client()

	md := metadata.New(map[string]string{"X-Owner-Id": ownerID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.GetWeekEvents(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}

func (c *GRPCClient) GetMonthEvents(
	ctx context.Context,
	ownerID string,
	req *v1.GetMonthEventsRequest,
) ([]*v1.Event, error) {
	client := c.client()

	md := metadata.New(map[string]string{"X-Owner-Id": ownerID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.GetMonthEvents(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}
