package rest

import (
	"Akso/meta"
	"Akso/store"
	"encoding/json"
	"net/http"
	"strings"
)

type foodManager struct {
	foodStore store.FoodStore
}

type serviceRequestBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewFoodManager() (*foodManager, error) {
	s, err := store.NewStore()
	if err != nil {
		return nil, err
	}
	return &foodManager{foodStore: s}, nil
}

func parseRequestURI(r *http.Request) (string, string) {
	segs := strings.Split(r.RequestURI, "/")
	segLength := len(segs)
	if segLength > 4 {
		return segs[4], ""
	}
	return "", ""
}

func decodeServiceRequest(r *http.Request) (*serviceRequestBody, error) {
	decoder := json.NewDecoder(r.Body)
	var request serviceRequestBody
	err := decoder.Decode(&request)
	if err != nil {
		return nil, err
	}

	// TODO: Verify if context is good
	return &request, nil
}

func decodeRequestBody(r *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	return nil
}

func sendBadRequestResponse(w http.ResponseWriter, errorMessage string) {
	payload, err := json.Marshal(errorResponse{
		Error: errorMessage,
	})
	if err != nil {
		// Log a error mesage
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(payload)
}

func sendPageNotFoundResponse(w http.ResponseWriter) {
	http.Error(w, "", http.StatusNotFound)
}

func sendEmptyListResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[]"))
}

func sendCreatedResponse(w http.ResponseWriter, object interface{}) {
	var payload []byte
	if object != nil {
		var err error
		payload, err = json.Marshal(object)
		if err != nil {
			sendBadRequestResponse(w, err.Error())
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if object != nil {
		w.Write(payload)
	}
}

func writeResponse(w http.ResponseWriter, object interface{}) {
	payload, err := json.Marshal(object)
	if err != nil {
		sendBadRequestResponse(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func (mgr *foodManager) createMaterial(w http.ResponseWriter, r *http.Request) {
	var material meta.Material
	err := decodeRequestBody(r, &material)
	if err != nil {
		sendBadRequestResponse(w, err.Error())
		return
	}

	if _, err := mgr.foodStore.GetMaterial(material.Name); err == nil {
		// material already exists.
		sendBadRequestResponse(w, "material already exists.")
		return
	}
	if err := mgr.foodStore.CreateMaterial(&material); err != nil {
		sendBadRequestResponse(w, err.Error())
		return
	}

	sendCreatedResponse(w, &material)
}

func (mgr *foodManager) getMaterial(w http.ResponseWriter, r *http.Request) {
	materialName, _ := parseRequestURI(r)
	if len(materialName) == 0 {
		sendBadRequestResponse(w, "Invalid material name.")
		return
	}

	material, err := mgr.foodStore.GetMaterial(materialName)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not Found") {
			sendPageNotFoundResponse(w)
		} else {
			sendBadRequestResponse(w, err.Error())
		}

		return
	}
	writeResponse(w, &material)
}
