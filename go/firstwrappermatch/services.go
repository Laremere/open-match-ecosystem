// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"math/rand"

	"open-match.dev/open-match-ecosystem/go/demoui"
	"open-match.dev/open-match-ecosystem/go/wrapper"
	"open-match.dev/open-match/pkg/pb"
)

func runServices(update demoui.SetFunc) {
	//   "ticketGenerator": runTicketGenerator,
	// "mmf":             runMmf,
	// // Default evaluator
	// "profilesProvider": runProfilesProvider,
	// "allocater":        runAllocater,

}

type services struct {
}

func (s *services) GenerateTicket(context.Context, *wrapper.GenerateTicketRequest) (*wrapper.GenerateTicketResponse, error) {
	return &wrapper.GenerateTicketResponse{
		Ticket: &pb.Ticket{},
	}, nil
}

func (s *services) GetProfiles(context.Context, *wrapper.GetProfilesRequest) (*wrapper.GetProfilesResponse, error) {
	return &wrapper.GetProfilesResponse{
		Profiles: []*pb.MatchProfile{
			{
				Name: "1v1",
				Pools: []*pb.Pool{
					{
						Name: "Everyone",
					},
				},
			},
		},
	}, nil
}

func (s *services) AllocateMatch(ctx context.Context, req *wrapper.AllocateMatchRequest) (*wrapper.AllocateMatchResponse, error) {
	assignment := &pb.Assignment{
		Connection: fmt.Sprintf("%d.%d.%d.%d:2222", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256)),
	}

	resp := &wrapper.AllocateMatchResponse{
		IdsToAssignments: make(map[string]*pb.Assignment),
	}

	for _, t := range req.Match.Tickets {
		resp.IdsToAssignments[t.Id] = assignment
	}

	return resp, nil
}
