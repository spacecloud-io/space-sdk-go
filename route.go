package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
)

type Route struct {
	handler     *Handler
	routeConfig *RouteConfig
}

// (opType: "query" or  opType: "mutation")
type RouteConfig struct {
	opId       string
	opType     string
	method     string
	url        string
	jsonSchema *JSONSchemaForInputAndOutput
}
type JSONSchemaForInputAndOutput struct {
	input  *jsonschema.Schema
	output *jsonschema.Schema
}

type Handler struct {
	inpType  reflect.Type
	outType  reflect.Type
	function interface{}
}

func (r *Route) method(method string) *Route {
	// todo - check if the method is valid string
	r.routeConfig.method = method
	return r
}
func (r *Route) url(url string) *Route {
	// todo - check for any invalid URL
	r.routeConfig.url = url
	return r
}

func (r *Route) fn(function interface{}) *Route {
	inpType, outType, err := validateAndRetrieveHandlerParamType(function)
	if err != nil {
		panic(err)
	}
	handler := &Handler{inpType, outType, function}
	r.handler = handler
	return r
}

func validateAndRetrieveHandlerParamType(fn interface{}) (reflect.Type, reflect.Type, error) {

	fnValue := reflect.ValueOf(fn)

	// check if the type if Func
	if fnValue.Kind() != reflect.Func {
		return nil, nil, errors.New("Provided handler is not a function")
	}

	// get the function's signature
	fnType := fnValue.Type()

	inpType, err := validateAndRetrieveInputParamType(fnType)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	outType, err := validateAndRetrieveOutputParamType(fnType)
	if err != nil {
		return nil, nil, err
	}
	return inpType, outType, nil
}
func validateAndRetrieveInputParamType(fnType reflect.Type) (reflect.Type, error) {

	noOfInputParam := fnType.NumIn()

	if noOfInputParam == 2 {
		if reflect.TypeOf((*context.Context)(nil)).Elem() == fnType.In(0) {
			return fnType.In(1), nil
		} else {
			return nil, errors.New("Second input parameter of handler function should be of type context")
		}
	} else {
		return nil, errors.New(fmt.Sprintf("Invalid number of input arguments - Expected : 2, Actual : %v", noOfInputParam))
	}

}

func validateAndRetrieveOutputParamType(fnType reflect.Type) (reflect.Type, error) {

	noOfOutputParam := fnType.NumOut()

	if noOfOutputParam == 1 {
		if reflect.TypeOf((*error)(nil)).Elem() == fnType.Out(0) {
			return fnType.Out(0), nil
		}
		return nil, errors.New("If the output of the function has 1 paramteres, it should of type error")
	} else if noOfOutputParam == 2 {
		if reflect.TypeOf((*error)(nil)).Elem() == fnType.Out(1) {
			return fnType.Out(0), nil
		}
		return nil, errors.New("Second output parameter of handler function should be of type error")
	}
	return nil, errors.New(fmt.Sprintf("Invalid number of output arguments of handler function - Expected : 2, Actual : %v", noOfOutputParam))
}
