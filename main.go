package main

import (
	"log"
	"mockgitea/internal/server"
	"github.com/gofiber/fiber/v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	mockServer := server.NewMockServer()

	app := fiber.New()                                                                              
                                                                                                  
  // 路由注册                                                                                     
  app.Get("/api/v1/user", mockServer.HandleUser)
  app.Get("/api/v1/repos/search", mockServer.HandleRepoSearch)                                    
  app.Get("/api/v1/repos/:owner/:repo/branches", mockServer.HandleBranches)                       
  app.Get("/api/v1/repos/:owner/:repo/commits", mockServer.HandleCommits)                         
  app.Get("/api/v1/repos/:owner/:repo/git/commits/:sha", mockServer.HandleSingleCommit)           
                                                                                                  
  app.Listen(":3333")      
}
