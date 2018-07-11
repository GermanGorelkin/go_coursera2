package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"context"
	"strconv"
)


func apiResponse(data interface{}, err error) []byte {
	m := make(map[string]interface{})
	if err != nil {
		m["error"] = err.Error()
	} else
	{
		m["error"] = ""
		m["response"] = data
	}

	b, _ := json.Marshal(m)
	return b
}

func apiParRequired(val, name string) error {
	if val == "" {
		return fmt.Errorf("%s must me not empty", name)
	}
	return nil
}

func apiParMin(val interface{}, name string, num int) error {
	switch v := val.(type) {
	case string:
		{
			if len([]rune(v)) < num {
				return fmt.Errorf("%s len must be >= %d", name, num)
			}
		}
	case int:
		{
			if v < num {
				return fmt.Errorf("%s must be >= %d", name, num)
			}
		}
	}
	return nil
}

func apiParMax(val interface{}, name string, num int) error {
	switch v := val.(type) {
	case string:
		{
			if len([]rune(v)) > num {
				return fmt.Errorf("%s len must be <= %d", name, num)
			}
		}
	case int:
		{
			if v > num {
				return fmt.Errorf("%s must be <= %d", name, num)
			}
		}
	}
	return nil
}


func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	
        case r.URL.Path == "/user/profile":
		srv.wrapperProfile(w, r)
    
        case r.URL.Path == "/user/create":
		if r.Method == http.MethodPost {
			srv.wrapperCreate(w, r)
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write(apiResponse("", fmt.Errorf("bad method")))
		}
    
	default:
		{
			w.WriteHeader(http.StatusNotFound)
			w.Write(apiResponse("", fmt.Errorf("unknown method")))
		}
	}
}


func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	
        case r.URL.Path == "/user/create":
		if r.Method == http.MethodPost {
			srv.wrapperCreate(w, r)
		} else {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write(apiResponse("", fmt.Errorf("bad method")))
		}
    
	default:
		{
			w.WriteHeader(http.StatusNotFound)
			w.Write(apiResponse("", fmt.Errorf("unknown method")))
		}
	}
}

    func (srv *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {

	// auth
	


	// valid
	
		
            login := r.FormValue("login")
        

        

		//Required
		if err := apiParRequired(login, "login"); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Min
		
		//Max
		

    


	// bl

	ctx, _ := context.WithCancel(r.Context())
	
	in := ProfileParams{
		
			Login:  login,
    	
	}

	u, err := srv.Profile(ctx, in)
	
	if err != nil {
		switch ar := err.(type) {
		case ApiError:
			w.WriteHeader(ar.HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	
		w.Write(apiResponse("", err))
		return
	}
	
	w.Write(apiResponse(u, err))
}
    func (srv *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {

	// auth
	
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write(apiResponse("", fmt.Errorf("unauthorized")))

		return
	}
    


	// valid
	
		
            login := r.FormValue("login")
        

        

		//Required
		if err := apiParRequired(login, "login"); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Min
		if err := apiParMin(login, "login", 10); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Max
		

    
		
            full_name := r.FormValue("full_name")
        

        

		//Required
		
		//Min
		
		//Max
		

    
		
            status := r.FormValue("status")
        

        

		//Required
		
		//Min
		
		//Max
		lstatus := map[string]struct{}{
	
		"user":      {},
	
		"moderator":      {},
	
		"admin":      {},
	
	}
	// status := r.FormValue("status")
	if status == "" {
		status = "user"
	}
	_, ok := lstatus[status]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", fmt.Errorf("status must be one of [user, moderator, admin]")))
	
		return
	}

    
		
            age, err := strconv.Atoi(r.FormValue("age"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(apiResponse("", fmt.Errorf("age must be int")))

				return
			}
        

        

		//Required
		
		//Min
		if err := apiParMin(age, "age", 0); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Max
		if err := apiParMax(age, "age", 128); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}

    


	// bl

	ctx, _ := context.WithCancel(r.Context())
	
	in := CreateParams{
		
			Login:  login,
    	
			Name:  full_name,
    	
			Status:  status,
    	
			Age:  age,
    	
	}

	u, err := srv.Create(ctx, in)
	
	if err != nil {
		switch ar := err.(type) {
		case ApiError:
			w.WriteHeader(ar.HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	
		w.Write(apiResponse("", err))
		return
	}
	
	w.Write(apiResponse(u, err))
}
    func (srv *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {

	// auth
	
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write(apiResponse("", fmt.Errorf("unauthorized")))

		return
	}
    


	// valid
	
		
            username := r.FormValue("username")
        

        

		//Required
		if err := apiParRequired(username, "username"); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Min
		if err := apiParMin(username, "username", 3); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Max
		

    
		
            account_name := r.FormValue("account_name")
        

        

		//Required
		
		//Min
		
		//Max
		

    
		
            class := r.FormValue("class")
        

        

		//Required
		
		//Min
		
		//Max
		lclass := map[string]struct{}{
	
		"warrior":      {},
	
		"sorcerer":      {},
	
		"rouge":      {},
	
	}
	// class := r.FormValue("class")
	if class == "" {
		class = "warrior"
	}
	_, ok := lclass[class]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")))
	
		return
	}

    
		
            level, err := strconv.Atoi(r.FormValue("level"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(apiResponse("", fmt.Errorf("level must be int")))

				return
			}
        

        

		//Required
		
		//Min
		if err := apiParMin(level, "level", 1); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}
		//Max
		if err := apiParMax(level, "level", 50); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiResponse("", err))

		return
		}

    


	// bl

	ctx, _ := context.WithCancel(r.Context())
	
	in := OtherCreateParams{
		
			Username:  username,
    	
			Name:  account_name,
    	
			Class:  class,
    	
			Level:  level,
    	
	}

	u, err := srv.Create(ctx, in)
	
	if err != nil {
		switch ar := err.(type) {
		case ApiError:
			w.WriteHeader(ar.HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	
		w.Write(apiResponse("", err))
		return
	}
	
	w.Write(apiResponse(u, err))
}