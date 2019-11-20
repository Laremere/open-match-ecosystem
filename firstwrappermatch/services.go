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
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"open-match.dev/open-match-ecosystem/demoui"
	"open-match.dev/open-match-ecosystem/wrapper"
	"open-match.dev/open-match/pkg/matchfunction"
	"open-match.dev/open-match/pkg/pb"
)

const port = 50502

var (
	logger = logrus.WithFields(logrus.Fields{
		"app":       "services",
		"component": "services.server",
	})
)

func runServices(update demoui.SetFunc) {
	mmlogic, err := grpc.Dial("om-mmlogic.open-match.svc.cluster.local:50503", grpc.WithInsecure())
	if err != nil {
		update(err.Error())
		return
	}
	defer mmlogic.Close()

	services := &services{
		update:  update,
		mmlogic: pb.NewMmLogicClient(mmlogic),
	}
	server := grpc.NewServer()
	wrapper.RegisterTicketGeneratorServer(server, services)
	wrapper.RegisterProfilesProviderServer(server, services)
	wrapper.RegisterAllocaterServer(server, services)
	pb.RegisterEvaluatorServer(server, services)
	pb.RegisterMatchFunctionServer(server, services)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"port":  port,
		}).Fatal("net.Listen() error")
	}

	logger.WithFields(logrus.Fields{
		"port": port,
	}).Info("TCP net listener initialized")

	err = server.Serve(ln)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("gRPC serve() error")
	}
}

type services struct {
	update  demoui.SetFunc
	mmlogic pb.MmLogicClient
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

func (s *services) Run(req *pb.RunRequest, stream pb.MatchFunction_RunServer) error {
	poolTickets, err := matchfunction.QueryPools(stream.Context(), s.mmlogic, req.GetProfile().GetPools())
	if err != nil {
		return err
	}

	tickets, ok := poolTickets["Everyone"]
	if !ok {
		return errors.New("Expected pool named Everyone.")
	}

	t := time.Now().Format("2006-01-02T15:04:05.00")

	matchesFound := 0
	for i := 0; i+1 < len(tickets); i += 2 {
		proposal := &pb.Match{
			MatchId:       fmt.Sprintf("profile-%s-time-%s-num-%d", req.Profile.Name, t, i/2),
			MatchProfile:  req.Profile.Name,
			MatchFunction: "first-match-mmf",
			Tickets: []*pb.Ticket{
				tickets[i], tickets[i+1],
			},
		}
		matchesFound++

		err := stream.Send(&pb.RunResponse{Proposal: proposal})
		if err != nil {
			return err
		}
	}

	s.update(fmt.Sprintf("Last run created %d matches", matchesFound))

	return nil
}

func (s *services) Evaluate(stream pb.Evaluator_EvaluateServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = stream.Send(&pb.EvaluateResponse{Match: req.GetMatch()})
		if err != nil {
			return err
		}
	}

	return nil
}
