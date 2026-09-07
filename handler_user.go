package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Nicokamo/blog_aggregator/internal/config"
	"github.com/Nicokamo/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db         *database.Queries
	CfgPointer *config.Config
}

type command struct {
	Name         string
	CommandSlice []string
}

type commands struct {
	MapCommandToHandler map[string]func(*state, command) error
}




func handlerLogin(s *state, cmd command) error {
	if len(cmd.CommandSlice) == 0 {
		return errors.New("no commandslice")
	}
	userName := cmd.CommandSlice[0]
	con := context.Background()
	_, err := s.db.GetUser(con, userName)
	if err != nil {
		fmt.Println("user does not exist")
		os.Exit(1)
	}
	err = s.CfgPointer.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Println("User has been set", userName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.CommandSlice) == 0 {
		return errors.New("No name provided")
	}
	userName := cmd.CommandSlice[0]
	con := context.Background()
	_, err := s.db.GetUser(con, userName)
	if err == nil {
		fmt.Println("user already exists")
		os.Exit(1)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	newID := uuid.New()
	current_time := time.Now()
	newUserPara := database.CreateUserParams{
		ID:        newID,
		CreatedAt: current_time,
		UpdatedAt: current_time,
		Name:      userName,
	}
	dbUser, err := s.db.CreateUser(con, newUserPara)
	if err != nil {
		return err
	}
	err = s.CfgPointer.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Println("New user created:", dbUser)
	return nil
}

func handlerReset(s *state, cmd command) error {
	con := context.Background()
	err := s.db.DeleteAllUsers(con)
	if err != nil {
		fmt.Println("Users could not be deleted")
		os.Exit(1)
	}
	fmt.Println("All users successfully deleted")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	con := context.Background()
	users, err := s.db.GetUsers(con)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == s.CfgPointer.CurrentUserName {
			fmt.Println(user, "(current)")
		} else {
			fmt.Println(user)
		}
	}
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	function, ok := c.MapCommandToHandler[cmd.Name]
	if !ok {
		return errors.New("command does not exist")
	}
	err := function(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.MapCommandToHandler[name] = f
}
