package main

import (
	"fmt"
	"log"
	"mockgitea/internal/config"
	"mockgitea/internal/server"
	"mockgitea/internal/utils"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	port := config.DefaultPort
	mockServer := server.NewMockServer()

	mux := http.NewServeMux()
	// 用户相关接口 - 获取当前用户信息
	mux.HandleFunc("/api/v1/user", mockServer.HandleUser)
	// 仓库搜索接口 - 搜索仓库列表
	mux.HandleFunc("/api/v1/repos/search", mockServer.HandleReposearch)
	// 仓库相关接口 - 处理单个仓库的CRUD操作
	mux.HandleFunc("/api/v1/repos/", mockServer.HandleRepoRoutes)
	// 404 未找到处理 - 处理未匹配的路由
	mux.HandleFunc("/", mockServer.HandleNotFound)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[mock-gitea] mock data ready: Users=%d Repos=%d anchor=%s", len(mockServer.Users), len(mockServer.Repos), mockServer.AnchorTime.Format(time.RFC3339))
	log.Printf("[mock-gitea] listening on %s", addr)
	log.Printf("[mock-gitea] try: curl http://localhost:%d/api/v1/user", port)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           utils.LoggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[mock-gitea] server stopped: %v", err)
	}
}
