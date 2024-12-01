//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
)

type ErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Event struct {
	EventID      string    `json:"eventId,omitempty"`
	StartAt      time.Time `json:"startAt,omitempty"`
	EndAt        time.Time `json:"endAt,omitempty"`
	Title        string    `json:"title,omitempty"`
	Description  string    `json:"description,omitempty"`
	NotifyBefore uint32    `json:"notifyBefore,omitempty"`
}

func EventFromPb(ev *v1.Event) *Event {
	return &Event{
		EventID:      ev.EventID,
		StartAt:      ev.StartAt.AsTime(),
		EndAt:        ev.EndAt.AsTime(),
		Title:        ev.Title,
		Description:  ev.Description,
		NotifyBefore: ev.NotifyBefore,
	}
}

func (ev *Event) ToPb() *v1.Event {
	return &v1.Event{
		EventID:      ev.EventID,
		StartAt:      timestamppb.New(ev.StartAt),
		EndAt:        timestamppb.New(ev.EndAt),
		Title:        ev.Title,
		Description:  ev.Description,
		NotifyBefore: ev.NotifyBefore,
	}
}

type CreateEventResponse struct {
	Event *Event `json:"event"`
}

type GetEventsResponse struct {
	Events []*Event `json:"events"`
}

type HTTPClient struct {
	Base    string
	Connect string
}

func NewHTTPClient(connect string) (*HTTPClient, error) {
	return &HTTPClient{
		Base:    "/api",
		Connect: connect,
	}, nil
}

func (c *HTTPClient) Done() error {
	return nil
}

func doReq[S any, T any](
	ctx context.Context,
	c *HTTPClient,
	method string,
	paths []string,
	ownerID string,
	req *S,
	resp T,
) (*T, error) {
	client := http.Client{}

	u := url.URL{
		Scheme: "http",
		Host:   c.Connect,
		Path:   path.Join(append([]string{c.Base}, paths...)...),
	}

	var err error
	var body []byte
	if req != nil {
		body, err = json.Marshal(req)
		if err != nil {
			return nil, err
		}
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		method,
		u.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	if ownerID != "" {
		httpReq.Header.Add("X-Owner-ID", ownerID)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode >= 400 {
		var resp ErrorResp
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(resp.Message)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *HTTPClient) CreateEvent(ctx context.Context, ownerID string, ev *v1.Event) (*v1.Event, error) {
	resp, err := doReq(
		ctx,
		c,
		http.MethodPost,
		[]string{"/v1/events"},
		ownerID,
		EventFromPb(ev),
		CreateEventResponse{},
	)
	if err != nil {
		return nil, err
	}

	return resp.Event.ToPb(), nil
}

func (c *HTTPClient) GetDayEvents(
	ctx context.Context,
	ownerID string,
	req *v1.GetDayEventsRequest,
) ([]*v1.Event, error) {
	resp, err := doReq(
		ctx,
		c,
		http.MethodGet,
		[]string{
			"/v1/events/query/day",
			strconv.Itoa(int(req.Day.Year)),
			strconv.Itoa(int(req.Day.Month)),
			strconv.Itoa(int(req.Day.Day)),
		},
		ownerID,
		new(struct{}),
		GetEventsResponse{},
	)
	if err != nil {
		return nil, err
	}

	events := []*v1.Event{}
	for _, ev := range resp.Events {
		events = append(events, ev.ToPb())
	}
	return events, nil
}

func (c *HTTPClient) GetWeekEvents(
	ctx context.Context,
	ownerID string,
	req *v1.GetWeekEventsRequest,
) ([]*v1.Event, error) {
	resp, err := doReq(
		ctx,
		c,
		http.MethodGet,
		[]string{
			"/v1/events/query/week",
			strconv.Itoa(int(req.StartDay.Year)),
			strconv.Itoa(int(req.StartDay.Month)),
			strconv.Itoa(int(req.StartDay.Day)),
		},
		ownerID,
		new(struct{}),
		GetEventsResponse{},
	)
	if err != nil {
		return nil, err
	}

	events := []*v1.Event{}
	for _, ev := range resp.Events {
		events = append(events, ev.ToPb())
	}
	return events, nil
}

func (c *HTTPClient) GetMonthEvents(
	ctx context.Context,
	ownerID string,
	req *v1.GetMonthEventsRequest,
) ([]*v1.Event, error) {
	resp, err := doReq(
		ctx,
		c,
		http.MethodGet,
		[]string{
			"/v1/events/query/month",
			strconv.Itoa(int(req.Month.Year)),
			strconv.Itoa(int(req.Month.Month)),
		},
		ownerID,
		new(struct{}),
		GetEventsResponse{},
	)
	if err != nil {
		return nil, err
	}

	events := []*v1.Event{}
	for _, ev := range resp.Events {
		events = append(events, ev.ToPb())
	}
	return events, nil
}
