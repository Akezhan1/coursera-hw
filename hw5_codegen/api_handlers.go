package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type resp map[string]interface{}


func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
	case "/user/profile":
		if !(r.Method == http.MethodGet || r.Method == http.MethodPost) {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}
		srv.handleMyApiProfile(w, r)
	case "/user/create":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}
		srv.handleMyApiCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resp{"error": "unknown method"})
		w.Write(data)
		return
	}	
}
	
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
	case "/user/create":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}
		srv.handleOtherApiCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resp{"error": "unknown method"})
		w.Write(data)
		return
	}	
}
	
func (srv *MyApi) handleMyApiProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	login := r.FormValue("login")
	if login == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "login must me not empty"})
		w.Write(data)
		return
	} 
	
	
	in := ProfileParams { 
		Login: login,
	}
	user,err := srv.Profile(context.Background(),in)
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}

	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
	
}
func (srv *MyApi) handleMyApiCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if at := r.Header.Get("X-Auth"); at != "100500" {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resp{"error": "unauthorized"})
		w.Write(data)
		return
	}
	
	
	login := r.FormValue("login")
	if login == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "login must me not empty"})
		w.Write(data)
		return
	} 
	
	if len(login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "login len must be >= 10"})
		w.Write(data)
		return
	}
	
	full_name := r.FormValue("full_name")
	status := r.FormValue("status")
	if status == "" {
		status = "user"
	}
	
	if !(status == "user" || status == "moderator" || status == "admin") {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "status must be one of [user, moderator, admin]"})
		w.Write(data)
		return
	}
	age_int := r.FormValue("age")
	age,err := strconv.Atoi(age_int)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "age must be int"})
		w.Write(data)
		return
	}
	if age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "age must be <= 128"})
		w.Write(data)
		return
	}
	if age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "age must be >= 0"})
		w.Write(data)
		return
	}
	
	
	in := CreateParams { 
		Login: login,
		Name: full_name,
		Status: status,
		Age: age,
	}
	user,err := srv.Create(context.Background(),in)
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}

	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
	
}
func (srv *OtherApi) handleOtherApiCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if at := r.Header.Get("X-Auth"); at != "100500" {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resp{"error": "unauthorized"})
		w.Write(data)
		return
	}
	
	
	username := r.FormValue("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "username must me not empty"})
		w.Write(data)
		return
	} 
	
	if len(username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "username len must be >= 3"})
		w.Write(data)
		return
	}
	
	account_name := r.FormValue("account_name")
	class := r.FormValue("class")
	if class == "" {
		class = "warrior"
	}
	
	if !(class == "warrior" || class == "sorcerer" || class == "rouge") {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "class must be one of [warrior, sorcerer, rouge]"})
		w.Write(data)
		return
	}
	level_int := r.FormValue("level")
	level,err := strconv.Atoi(level_int)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "level must be int"})
		w.Write(data)
		return
	}
	if level > 50 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "level must be <= 50"})
		w.Write(data)
		return
	}
	if level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "level must be >= 1"})
		w.Write(data)
		return
	}
	
	
	in := OtherCreateParams { 
		Username: username,
		Name: account_name,
		Class: class,
		Level: level,
	}
	user,err := srv.Create(context.Background(),in)
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}

	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
	
}