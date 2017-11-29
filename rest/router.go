package rest

import (
	"github.com/gorilla/mux"
	"net/http"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

func initRouters() (*routes, error) {
	manager, err := NewFoodManager()
	if err != nil {
		return nil, err
	}

	return &routes{
		route{
			"CreateMaterial",
			"POST",
			FoodPath + "material",
			manager.createMaterial,
		},

		route{
			"GetMaterial",
			"GET",
			FoodPath + "material/{materialName}",
			manager.getMaterial,
		},
	}, nil
}

func NewRouter() (*mux.Router, error) {
	routes, err := initRouters()
	if err != nil {
		return nil, err
	}
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range *routes {
		var handler http.Handler
		handler = route.HandlerFunc
		//handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router, nil
}
