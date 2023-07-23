package main

import (
	"fmt"
	"recorder/internal/server"
	"github.com/google/uuid"
)

func main() {
	uuid := uuid.NewString()
    fmt.Println(uuid)
	server.Start_server()
}
