package app

import (
	"fmt"
	"time"

	"github.com/pluhe7/gophermart/internal/model"
	"github.com/pluhe7/gophermart/internal/util"
)

const tokenLength = 64

func (a *App) CreateSession(userID int) (*model.Session, error) {
	session := &model.Session{
		Token:          generateToken(),
		UserID:         userID,
		CreatedAt:      time.Now(),
		ExpireAt:       time.Now().Add(time.Duration(model.SessionExpireDays) * 24 * time.Hour),
		LastActivityAt: time.Now(),
	}

	err := a.Server.Storage.SessionStorage.Create(session)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return session, nil
}

func generateToken() string {
	return util.GenerateRandomString(tokenLength)
}

func (a *App) GetSession(token string) (*model.Session, error) {
	session, err := a.Server.Storage.SessionStorage.Get(token)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return session, nil
}
