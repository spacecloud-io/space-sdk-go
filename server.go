package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/jsonschema"
)

type Server struct {
	serverConfig *ServerConfig
	router       *Router
}

type ServerConfig struct {
	name string
	port int
}

type RegistrationFunctionSchema struct {
	InputSchema  *jsonschema.Schema `json:"input"`
	OutputSchema *jsonschema.Schema `json:"output"`
	FnName       string             `json:"functionName"`
	EndPoint     string             `json:"endpoint"`
}

// Creates and returns an empty Server, with a basic serverConfig
func Create(name string) *Server {

	return &Server{
		serverConfig: &ServerConfig{
			name: name,
			port: 3000,
		},
		router: &Router{
			routes: []*Route{},
		},
	}
}

func (s *Server) Router() *Router {
	return s.router
}

func (server *Server) Start() {
	// todo: check if router was nil
	// todo: check if the routes was empty
	// todo: any additional checks needed ?
	// todo: check if the server was already started ?

	// create a new internal go-chi router
	chi_router := chi.NewRouter()

	// read all the routes from the router
	routes := server.router.routes
	for _, route := range routes {
		newFunc := func(w http.ResponseWriter, r *http.Request) {
			// Decode the JSON request body into the struct instance
			structPtr := reflect.New(route.handler.inpType).Interface()
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			err := decoder.Decode(structPtr)
			if err != nil {
				http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
				return
			}

			decodedValue := reflect.ValueOf(structPtr).Elem()
			decodedType := decodedValue.Type()

			if decodedType != route.handler.inpType {
				// Compare the structure of the decoded request with the expected type
				if !reflect.DeepEqual(decodedType, route.handler.inpType) {
					http.Error(w, "JSON body does not match expected structure", http.StatusBadRequest)
					return
				}
			}

			// Fetching the function from the saved Handler
			fn := route.handler.function

			function := reflect.ValueOf(fn)
			args := []reflect.Value{
				reflect.ValueOf(context.TODO()),
				reflect.ValueOf(structPtr).Elem(),
			}
			result := function.Call(args)
			if len(result) == 1 {
				err := result[0].Interface()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "text/plain")
					fmt.Fprint(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprint(w)
				return
			} else if len(result) == 2 {
				res := result[0].Interface()
				err := result[1].Interface()
				fmt.Println(res)
				fmt.Println(err)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "text/plain")
					fmt.Fprint(w, err)
					return
				}

				resValue := reflect.ValueOf(res)
				if resValue.Type().AssignableTo(route.handler.outType) {
					responseJSON, err := json.Marshal(res)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						fmt.Fprintf(w, "Error marshaling response: %v", err)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(responseJSON)
					return

				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "text/plain")
					fmt.Fprintf(w, "Response of function is not matching the output type: %v", err)
					return
				}
			}
		}
		if route.routeConfig.method == http.MethodGet {
			chi_router.Get(route.routeConfig.url, newFunc)
		} else if route.routeConfig.method == http.MethodPost {
			chi_router.Post(route.routeConfig.url, newFunc)
		}
	}

	// Add an endpoint for fetching the JSON Schema
	endPointForJSONSchemaFunction := func(w http.ResponseWriter, r *http.Request) {
		jsonOpenAPI, err := generateOpenAPIDocument(routes)
		if err != nil {
			// Handle the error, e.g., log it or return an error response
			http.Error(w, "Failed to generate OpenAPI JSON", http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to "application/json"
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON file as the response body
		_, err = w.Write(jsonOpenAPI)
		if err != nil {
			// Handle the error, e.g., log it or return an error response
			http.Error(w, "Failed to send OpenAPI JSON response", http.StatusInternalServerError)
			return
		}
	}
	chi_router.Get("/openapi.json", endPointForJSONSchemaFunction)

	http.ListenAndServe(":8080", chi_router)
}

// This function generates the Open API for all the registered Routes
func generateOpenAPIDocument(routes []*Route) ([]byte, error) {

	openAPI := &openapi3.T{
		OpenAPI: "3.0.0.",
		Info: &openapi3.Info{
			Title:       "Go Server",
			Description: "List of all the registered endpoints",
		},
	}

	for _, route := range routes {

		jsonInpSchema := route.routeConfig.jsonSchema.input

		jsonObject, err := json.Marshal(jsonInpSchema)
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
		fmt.Println(string(jsonObject))

		schema := openapi3.NewSchema()
		if err := schema.UnmarshalJSON(jsonObject); err != nil {
			panic(err)
		}

		endpoint := &openapi3.Operation{
			RequestBody: &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: schema,
							},
						},
					},
				},
			},
		}

		// Add the response object too
		openAPI.AddOperation(route.routeConfig.url, route.routeConfig.method, endpoint)
	}
	return openAPI.MarshalJSON()
}
