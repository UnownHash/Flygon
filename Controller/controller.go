package Controller

import log "github.com/sirupsen/logrus"

type ControllerBody struct {
	Type     string `json:"type" binding:"required"`
	Uuid     string `json:"uuid" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func Controller(body ControllerBody) (interface{}, error) {
	switch body.Type {
	case "init":
		return handleInit(body)
	case "get_task":
		return handleGetTask(body)
	default:
		return nil, nil
	}
}

func handleInit(body ControllerBody) (interface{}, error) {
	log.Printf("HandleInit")
	return nil, nil
}

func handleGetTask(body ControllerBody) (interface{}, error) {
	log.Printf("HandleGetTask")
	return nil, nil
}
