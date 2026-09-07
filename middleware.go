package main

import (
	"context"

	"github.com/Nicokamo/blog_aggregator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		currentUserName := s.CfgPointer.CurrentUserName
		currentUser, err := s.db.GetUser(ctx, currentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, currentUser)
	}
}
