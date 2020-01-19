// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"math/rand"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"
)

type TestEnvironment struct {
	Teams        []*model.Team
	Environments []TeamEnvironment
}

func createTestEnvironmentWithTeams(a *App, client *model.Client4, rangeTeams utils.Range, rangeChannels utils.Range, rangeUsers utils.Range, rangePosts utils.Range, fuzzy bool) (TestEnvironment, bool) {
	rand.Seed(time.Now().UTC().UnixNano())

	teamCreator := newAutoTeamCreator(client)
	teamCreator.Fuzzy = fuzzy
	teams, err := teamCreator.createTestTeams(rangeTeams)
	if !err {
		return TestEnvironment{}, false
	}

	environment := TestEnvironment{teams, make([]TeamEnvironment, len(teams))}

	for i, team := range teams {
		userCreator := newAutoUserCreator(a, client, team)
		userCreator.Fuzzy = fuzzy
		randomUser, err := userCreator.createRandomUser()
		if !err {
			return TestEnvironment{}, false
		}
		client.LoginById(randomUser.Id, USER_PASSWORD)
		teamEnvironment, err := createTestEnvironmentInTeam(a, client, team, rangeChannels, rangeUsers, rangePosts, fuzzy)
		if !err {
			return TestEnvironment{}, false
		}
		environment.Environments[i] = teamEnvironment
	}

	return environment, true
}

func createTestEnvironmentInTeam(a *App, client *model.Client4, team *model.Team, rangeChannels utils.Range, rangeUsers utils.Range, rangePosts utils.Range, fuzzy bool) (TeamEnvironment, bool) {
	rand.Seed(time.Now().UTC().UnixNano())

	// We need to create at least one user
	if rangeUsers.Begin <= 0 {
		rangeUsers.Begin = 1
	}

	userCreator := newAutoUserCreator(a, client, team)
	userCreator.Fuzzy = fuzzy
	users, err := userCreator.createTestUsers(rangeUsers)
	if !err {
		return TeamEnvironment{}, false
	}
	usernames := make([]string, len(users))
	for i, user := range users {
		usernames[i] = user.Username
	}

	channelCreator := newAutoChannelCreator(client, team)
	channelCreator.Fuzzy = fuzzy
	channels, err := channelCreator.createTestChannels(rangeChannels)

	// Have every user join every channel
	for _, user := range users {
		for _, channel := range channels {
			client.LoginById(user.Id, USER_PASSWORD)
			client.AddChannelMember(channel.Id, user.Id)
		}
	}

	if !err {
		return TeamEnvironment{}, false
	}

	numPosts := utils.RandIntFromRange(rangePosts)
	numImages := utils.RandIntFromRange(rangePosts) / 4
	for j := 0; j < numPosts; j++ {
		user := users[utils.RandIntFromRange(utils.Range{Begin: 0, End: len(users) - 1})]
		client.LoginById(user.Id, USER_PASSWORD)
		for i, channel := range channels {
			postCreator := newAutoPostCreator(client, channel.Id)
			postCreator.HasImage = i < numImages
			postCreator.Users = usernames
			postCreator.Fuzzy = fuzzy
			postCreator.createRandomPost()
		}
	}

	return TeamEnvironment{users, channels}, true
}
