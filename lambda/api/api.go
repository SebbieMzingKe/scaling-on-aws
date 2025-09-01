package api

import (
	"encoding/json"
	"fmt"
	"lambda-func/database"
	"lambda-func/types"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type ApiHandler struct {
	dbStore database.UserStore
}

func NewApiHandler(dbStore database.UserStore) ApiHandler {
	return ApiHandler{
		dbStore: dbStore,
	}
}

func (api ApiHandler) RegisterUserHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var registerUser types.RegisterUser

	err := json.Unmarshal([]byte(request.Body), &registerUser)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "invalid request",
			StatusCode: http.StatusBadRequest,
		}, err
	}
	if registerUser.Username == "" || registerUser.Password == "" {
		return events.APIGatewayProxyResponse{
			Body:       "invalid request - field empty",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	userExists, err := api.dbStore.DoesUSerExist(registerUser.Username)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "internals server error",
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	if userExists {
		return events.APIGatewayProxyResponse{
			Body:       "user already exists",
			StatusCode: http.StatusConflict,
		}, nil
	}

	user, err := types.NewUser(registerUser)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "internals server error",
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("could not create user - %v", err)
	}

	err = api.dbStore.InsertUser(user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "internals server error",
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("error inserting user - %v", err)
	}

	return events.APIGatewayProxyResponse{
		Body:       "succesfully registered user",
		StatusCode: http.StatusOK,
	}, nil
	// return nil
}

func (api ApiHandler) LoginUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var loginRequest LoginRequest

	err := json.Unmarshal([]byte(request.Body), &loginRequest)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "invalid request",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	user, err := api.dbStore.GetUser(loginRequest.Username)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "internal server errror",
			StatusCode: http.StatusInternalServerError,
		}, err

	}

	if !types.ValidatePassword(user.PasswordHash, loginRequest.Password) {
		return events.APIGatewayProxyResponse{
			Body:       "invalid user credentials",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	return events.APIGatewayProxyResponse{
		Body: "successfully logged in",
		StatusCode: http.StatusOK,
	}, nil
}
