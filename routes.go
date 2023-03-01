package main

import (
	"Flygon/db"
	"github.com/gin-gonic/gin"
)

var dbDetails *db.DbDetails

func ConnectDatabase(dbd *db.DbDetails) {
	dbDetails = dbd
}

func Controller(c *gin.Context) {
	print("Got here")
}
