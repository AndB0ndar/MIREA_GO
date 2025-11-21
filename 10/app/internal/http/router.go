package router

import (
    "app/internal/core"
    "app/internal/http/middleware"
    "app/internal/platform/config"
    "app/internal/platform/jwt"
    "app/internal/repo"
    "github.com/go-chi/chi/v5"
    "net/http"
)

func Build(cfg config.Config) http.Handler {
    r := chi.NewRouter()

    // Инициализация зависимостей
    userRepo := repo.NewUserMem()
    blacklist := repo.NewBlacklistMem()
    jwtv := jwt.NewHS256(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
    svc := core.NewService(userRepo, blacklist, jwtv)

    // Публичные маршруты
    r.Post("/login", svc.LoginHandler)
    r.Post("/refresh", svc.RefreshHandler)
    r.Post("/logout", svc.LogoutHandler)

    // Защищённые маршруты для всех аутентифицированных пользователей
    r.Group(func(priv chi.Router) {
        priv.Use(middleware.AuthN(jwtv))
        priv.Use(middleware.AuthZRoles("admin", "user"))

        priv.Get("/me", svc.MeHandler)

        // ABAC защита - пользователи могут получать только свои данные
        priv.With(middleware.ABACMiddleware).Get("/users/{id}", svc.GetUserHandler)
    })

    // Маршруты только для админов
    r.Group(func(admin chi.Router) {
        admin.Use(middleware.AuthN(jwtv))
        admin.Use(middleware.AuthZRoles("admin"))

        admin.Get("/admin/stats", svc.AdminStats)
        // Админы могут получать любые пользовательские данные без ABAC ограничений
        admin.Get("/admin/users/{id}", svc.GetUserHandler)
    })

    return r
}
