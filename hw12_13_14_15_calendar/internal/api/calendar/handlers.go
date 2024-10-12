package calendar

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	proto "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/grpc/auth"
	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

func (a *App) CreateEvent(ctx context.Context, req *proto.CreateEventRequest) (*proto.CreateEventResponse, error) {
	ownerID, err := auth.OwnerIDFromContext(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateEvent", whereAttr("OwnerIDFromContext"))
	}

	event, err := protoToModel(req.Event, ownerID)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateEvent", whereAttr("protoToModel"))
	}

	err = a.business.CreateEvent(ctx, event)
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateEvent", whereAttr("business.CreateEvent"))
	}

	return &proto.CreateEventResponse{}, nil
}

func (a *App) UpdateEvent(ctx context.Context, req *proto.UpdateEventRequest) (*proto.UpdateEventResponse, error) {
	ownerID, err := auth.OwnerIDFromContext(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "UpdateEvent", whereAttr("OwnerIDFromContext"))
	}

	event, err := protoToModel(req.Event, ownerID)
	if err != nil {
		return nil, a.handleError(ctx, err, "UpdateEvent", whereAttr("protoToModel"))
	}

	err = a.business.UpdateEvent(ctx, event)
	if err != nil {
		return nil, a.handleError(ctx, err, "UpdateEvent", whereAttr("business.UpdateEvent"))
	}

	return &proto.UpdateEventResponse{}, nil
}

func (a *App) DeleteEvent(ctx context.Context, req *proto.DeleteEventRequest) (*proto.DeleteEventResponse, error) {
	ownerID, err := auth.OwnerIDFromContext(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "DeleteEvent", whereAttr("OwnerIDFromContext"))
	}

	eventID, err := model.NewIDFromString(req.EventID)
	if err != nil {
		return nil, a.handleError(ctx, err, "DeleteEvent", whereAttr("model.NewIDFromString"))
	}

	err = a.business.DeleteEvent(ctx, ownerID, eventID)
	if err != nil {
		return nil, a.handleError(ctx, err, "DeleteEvent", whereAttr("business.DeleteEvent"))
	}

	return &proto.DeleteEventResponse{}, nil
}

func (a *App) GetDayEvents(ctx context.Context, req *proto.GetDayEventsRequest) (*proto.GetDayEventsResponse, error) {
	ownerID, err := auth.OwnerIDFromContext(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetDayEvents", whereAttr("OwnerIDFromContext"))
	}

	events, err := a.business.GetDayEvents(
		ctx,
		ownerID,
		int(req.Day.Year),
		int(req.Day.Month),
		int(req.Day.Day),
	)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetDayEvents", whereAttr("business.GetDayEvents"))
	}

	return &proto.GetDayEventsResponse{Events: modelsToProto(events)}, nil
}

func (a *App) GetWeekEvents(
	ctx context.Context,
	req *proto.GetWeekEventsRequest,
) (*proto.GetWeekEventsResponse, error) {
	ownerID, err := auth.OwnerIDFromContext(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetWeekEvents", whereAttr("OwnerIDFromContext"))
	}

	events, err := a.business.GetWeekEvents(
		ctx,
		ownerID,
		int(req.StartDay.Year),
		int(req.StartDay.Month),
		int(req.StartDay.Day),
	)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetWeekEvents", whereAttr("business.GetWeekEvents"))
	}

	return &proto.GetWeekEventsResponse{Events: modelsToProto(events)}, nil
}

func (a *App) GetMonthEvents(
	ctx context.Context,
	req *proto.GetMonthEventsRequest,
) (*proto.GetMonthEventsResponse, error) {
	ownerID, err := auth.OwnerIDFromContext(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetMonthEvents", whereAttr("OwnerIDFromContext"))
	}

	events, err := a.business.GetMonthEvents(
		ctx,
		ownerID,
		int(req.Month.Year),
		int(req.Month.Month),
	)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetMonthEvents", whereAttr("business.GetMonthEvents"))
	}

	return &proto.GetMonthEventsResponse{Events: modelsToProto(events)}, nil
}

func protoToModel(p *proto.Event, ownerID model.OwnerID) (model.Event, error) {
	eventID, err := model.NewIDFromString(p.EventID)
	if err != nil {
		return model.Event{}, err
	}

	title, err := model.NewTitle(p.Title)
	if err != nil {
		return model.Event{}, err
	}

	ev, err := model.NewEvent(eventID, ownerID, title, p.StartAt.AsTime(), p.EndAt.AsTime())
	if err != nil {
		return model.Event{}, err
	}

	ev.Description = p.Description
	ev.NotifyBefore = uint(p.NotifyBefore)

	return ev, nil
}

func modelToProto(event model.Event) *proto.Event {
	return &proto.Event{
		EventID:      string(event.EventID()),
		StartAt:      timestamppb.New(event.StartAt()),
		EndAt:        timestamppb.New(event.EndAt()),
		Title:        string(event.Title),
		Description:  event.Description,
		NotifyBefore: uint32(event.NotifyBefore),
	}
}

func modelsToProto(events []model.Event) []*proto.Event {
	protoEvents := make([]*proto.Event, len(events))
	for i := range len(events) {
		protoEvents[i] = modelToProto(events[i])
	}

	return protoEvents
}
