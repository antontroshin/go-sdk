/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
)

// AddBindingInvocationHandler appends provided binding invocation handler with its name to the service.
func (s *Server) AddBindingInvocationHandler(name string, fn common.BindingInvocationHandler) error {
	if name == "" {
		return errors.New("binding name required")
	}
	if fn == nil {
		return errors.New("binding handler required")
	}
	s.bindingHandlers[name] = fn
	return nil
}

// ListInputBindings is called by Dapr to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage.
func (s *Server) ListInputBindings(ctx context.Context, in *emptypb.Empty) (*pb.ListInputBindingsResponse, error) {
	list := make([]string, 0)
	for k := range s.bindingHandlers {
		list = append(list, k)
	}

	return &pb.ListInputBindingsResponse{
		Bindings: list,
	}, nil
}

// OnBindingEvent gets invoked every time a new event is fired from a registered binding. The message carries the binding name, a payload and optional metadata.
func (s *Server) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	if in == nil {
		return nil, errors.New("nil binding event request")
	}
	if fn, ok := s.bindingHandlers[in.GetName()]; ok {
		e := &common.BindingEvent{
			Data:     in.GetData(),
			Metadata: in.GetMetadata(),
		}
		data, err := fn(ctx, e)
		if err != nil {
			return nil, fmt.Errorf("error executing %s binding: %w", in.GetName(), err)
		}
		return &pb.BindingEventResponse{
			Data: data,
		}, nil
	}

	return nil, fmt.Errorf("binding not implemented: %s", in.GetName())
}
