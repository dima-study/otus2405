// grpc/auth - пакет отвечающий за "авторизацию",
// По хорошему - нужен отдельный сервис.
// По факту - просто провряет корректность OwnerID, т.к. ДЗ не требует авторизации
//
// OwnerID нужен, чтобы корректно привязывать события в коллекциях событий.
package auth

import (
	"context"
	"errors"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

var ErrMissedOwnerID = errors.New("value of owner id is missed")

type key int

const AuthOwnerKey = key(1)

func ValidateOwner(owner string) (model.OwnerID, error) {
	ownerID, err := model.NewOwnerIDFromString(owner)

	return ownerID, err
}

func OwnerIDFromContext(ctx context.Context) (model.OwnerID, error) {
	ownerID, ok := ctx.Value(AuthOwnerKey).(model.OwnerID)
	if !ok {
		return model.OwnerID(""), ErrMissedOwnerID
	}

	return ownerID, nil
}
