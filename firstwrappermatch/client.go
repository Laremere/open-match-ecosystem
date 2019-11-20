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
	"io"
	"time"

	"google.golang.org/grpc"
	"open-match.dev/open-match-ecosystem/demoui"
	"open-match.dev/open-match-ecosystem/wrapper"
)

func runClients(update demoui.SetFunc) {
	update("Main menu (sleeping)")
	time.Sleep(time.Second * 3)
	update("Finding match")
	connInfo, err := findMatch()
	if err != nil {
		update("Error finding match:" + err.Error())
	} else {
		update("Playing match on " + connInfo + " (sleeping)")
	}
	time.Sleep(time.Second * 5)
}

func findMatch() (string, error) {
	//////////////////////////////////////////////////////////////////////////////
	conn, err := grpc.Dial("om-front-door.open-match.svc.cluster.local:50520", grpc.WithInsecure())
	if err != nil {
		return "", fmt.Errorf("Error connecting to front door: %w", err)
	}
	defer conn.Close()
	fd := wrapper.NewFrontDoorClient(conn)

	//////////////////////////////////////////////////////////////////////////////
	stream, err := fd.FindMatch(context.Background())
	if err != nil {
		return "", fmt.Errorf("Error starting fd.FindMatch: %w", err)
	}

	err = stream.Send(&wrapper.FindMatchRequest{})
	if err != nil {
		return "", fmt.Errorf("Error sending fd.FindMatch: %w", err)
	}

	var resp *wrapper.FindMatchResponse
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("Error recieved from fd.FindMatch: %w", err)
		}
		resp = r
	}

	//////////////////////////////////////////////////////////////////////////////
	if resp.State != wrapper.FindMatchResponse_ASSIGNED {
		return "", fmt.Errorf("Unexpected state from findMatch: %v", resp.State.String())
	}
	a := resp.Assignment
	if a == nil {
		return "", fmt.Errorf("Missing assignment in response from findMatch.")
	}

	if a.Error != nil {
		return "", fmt.Errorf("Assignment from findMatch has error code %d, message: %s", a.Error.Code, a.Error.Message)
	}

	return a.Connection, nil
}
