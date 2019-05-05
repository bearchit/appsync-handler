package appsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Handler interface {
	Handle(context.Context, json.RawMessage) (interface{}, error)
}

type Resolver interface{}

type Resolvers map[string]Resolver

type handler struct {
	resolvers Resolvers
}

func (h *handler) AddResolver(field string, r Resolver) {
	h.resolvers[field] = r
}

func NewHandler() *handler {
	return &handler{
		resolvers: make(Resolvers),
	}
}

type Payload struct {
	Resolve   string          `json:"resolve"`
	Arguments json.RawMessage `json:"arguments"`
}

func (h handler) Handle(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	payload := new(Payload)
	if err := json.Unmarshal(raw, payload); err != nil {
		return nil, err
	}

	resolverFunc, ok := h.resolvers[payload.Resolve]
	if !ok {
		return nil, errors.New("no matched resolver")
	}

	resolver := reflect.ValueOf(resolverFunc)
	resolverType := reflect.TypeOf(resolverFunc)
	if resolver.Kind() != reflect.Func {
		return nil, fmt.Errorf("resolver kind %s is not %s", resolverType.Kind(), reflect.Func)
	}
	hasContext, err := validateArguments(resolverType)
	if err != nil {
		return nil, err
	}
	if err := validateReturns(resolverType); err != nil {
		return nil, err
	}
	args := make([]reflect.Value, 0)
	if hasContext {
		args = append(args, reflect.ValueOf(ctx))
	}
	if (resolverType.NumIn() == 1 && !hasContext) || resolverType.NumIn() == 2 {
		t := resolverType.In(resolverType.NumIn() - 1)
		v := reflect.New(t)
		if err := json.Unmarshal(payload.Arguments, v.Interface()); err != nil {
			return nil, err
		}
		args = append(args, v.Elem())
	}

	return call(resolver, args)
}

func call(resolver reflect.Value, args []reflect.Value) (interface{}, error) {
	response := resolver.Call(args)
	var err error
	if len(response) > 0 {
		if errVal, ok := response[len(response)-1].Interface().(error); ok {
			err = errVal
		}
	}
	var val interface{}
	if len(response) > 1 {
		val = response[0].Interface()

	}
	return val, err
}

func validateArguments(resolver reflect.Type) (bool, error) {
	hasContext := false
	if resolver.NumIn() > 2 {
		return false, fmt.Errorf("resolvers may not take more than two arguments, but resolver takes %d", resolver.NumIn())
	} else if resolver.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := resolver.In(0)
		hasContext = argumentType.Implements(contextType)
		if resolver.NumIn() > 1 && !hasContext {
			return false, fmt.Errorf("resolver takes two arguments, but the first is not Context. got %s", argumentType)
		}
	}

	return hasContext, nil
}

func validateReturns(resolver reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if resolver.NumOut() > 2 {
		return fmt.Errorf("resolver may not return more than two values")
	} else if resolver.NumOut() > 1 {
		if !resolver.Out(1).Implements(errorType) {
			return fmt.Errorf("resolver returnes two values, but the second does not implement error")
		}
	} else if resolver.NumOut() == 1 {
		if !resolver.Out(0).Implements(errorType) {
			return fmt.Errorf("resolver returns a single value, but it doest not implement error")
		}
	}
	return nil
}
