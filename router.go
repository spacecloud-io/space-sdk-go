package main

import (
	"net/http"
)

type Router struct {
	name   string
	routes []*Route
}

// func (r *Router) Query(opId string, fn interface{}) *Route {
// 	// todo: should I return error?
// 	inpType, outType, err := validateAndRetrieveHandlerParamType(fn)
// 	if err != nil {
// 		panic(err)
// 	}

// 	handler := &Handler{inpType, outType, fn}
// 	route := &Route{
// 		handler: handler,
// 		routeConfig: &RouteConfig{
// 			opId:   opId,
// 			opType: "query",
// 			method: http.MethodGet,
// 			url:    "/v1/" + opId,
// 			jsonSchema: &JSONSchemaForInputAndOutput{
// 				input:  jsonschema.ReflectFromType(inpType),
// 				output: jsonschema.ReflectFromType(outType),
// 			},
// 		}}
// 	r.routes = append(r.routes, route)
// 	return route
// }

func (r *Router) Query(items ...interface{}) *Route {
	var operationId string
	var function interface{}
	var route *Route
	// if the Query function is called with operationID
	if len(items) >= 1 {
		operationId = items[0].(string)

		// prepare the config.
		// todo: can be refactored into a prepareConfig func
		var config RouteConfig
		config.opId = operationId
		config.opType = "query"
		config.method = http.MethodGet
		config.url = "/v1/" + operationId
		route = NewRoute(&config)
	}

	// if the function is called alongwith the Handler fn
	if len(items) == 2 {
		function = items[1]
		route = route.Fn(function)

	} else if len(items) < 1 || len(items) > 2 {
		panic("Illegal number of arguments provided to Query")
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Mutation(items ...interface{}) *Route {
	var operationId string
	var function interface{}
	var route *Route
	// if the Mutation function is called with operationID
	if len(items) >= 1 {
		operationId = items[0].(string)
		var config RouteConfig
		config.opId = operationId
		config.opType = "mutation"
		config.method = http.MethodPost
		config.url = "/v1/" + operationId
		route = NewRoute(&config)
	}

	// if the function is called alongwith the Handler fn
	if len(items) == 2 {
		function = items[1]
		route = route.Fn(function)
	} else if len(items) < 1 || len(items) > 2 {
		panic("Illegal number of arguments provided to Mutation")
	}
	r.routes = append(r.routes, route)
	return route
}

// For now, this wrapper is not needed
// func wrapper(fn interface{}) func(context.Context, interface{}) []interface{} {
// 	fnValue := reflect.ValueOf(fn)

// 	return func(ctx context.Context, i interface{}) []interface{} {
// 		args := []reflect.Value{
// 			reflect.ValueOf(ctx),
// 			reflect.ValueOf(i),
// 		}

// 		result := fnValue.Call(args)

// 		if len(result) == 1 {
// 			err := result[0].Interface()
// 			return []interface{}{
// 				err,
// 			}
// 		} else if len(result) == 2 {
// 			res := result[0].Interface()
// 			err := result[1].Interface()

// 			if err != nil {
// 				return []interface{}{
// 					nil,
// 					err,
// 				}
// 			}

// 			return []interface{}{
// 				res,
// 				nil,
// 			}
// 		}

// 		return []interface{}{
// 			fmt.Errorf("Invalid function signature: expected (context.Context, struct) (struct, error)"),
// 		}
// 	}
// }
