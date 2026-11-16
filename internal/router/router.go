package router

import (
	"clothingStore/internal/handler"
	mw "clothingStore/internal/middleware"
	"clothingStore/internal/models"
	"clothingStore/internal/web"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	Product   *handler.ProductHandler
	Photo     *handler.PhotoHandler
	Role      *handler.RoleHandler
	User      *handler.UserHandler
	Cart      *handler.CartHandler
	Order     *handler.OrderHandler
	OrderItem *handler.OrderItemHandler
	Auth      *handler.AuthHandler
	JWTSecret string
	LogMw     func(http.Handler) http.Handler
	UploadDir string
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(d.LogMw)

	r.Handle("/uploads/*", http.StripPrefix("/uploads/",
		http.FileServer(http.Dir(d.UploadDir))))

	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic("embed static: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(staticFS))
	r.Handle("/style.css", fileServer)
	r.Handle("/app.js", fileServer)
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			http.Error(w, "index not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})

	//Функции, доступные гостю
	r.Group(func(pub chi.Router) {
		pub.Post("/login", d.Auth.Login)
		pub.Post("/register", d.Auth.Register)
		pub.Get("/product", d.Product.List)
		pub.Get("/product/{id}", d.Product.Get)
		pub.Get("/photo/{id}", d.Photo.Get)
		pub.Get("/photo", d.Photo.ListByProduct)
	})

	// Функции для которыз нужен вход
	r.Group(func(auth chi.Router) {
		auth.Use(mw.Auth(d.JWTSecret))
		//баланс для любого авторизованого юзера
		auth.Get("/me", d.User.Me)
		auth.Post("/balance/topup", d.User.TopUp)
		auth.Post("/order/{id}/cancel", d.Order.Cancel)

		// Клиент
		auth.Post("/cart", d.Cart.Create)
		auth.Get("/cart/{user_id}", d.Cart.GetByUser)
		auth.Put("/cart/{id}", d.Cart.Update)
		auth.Delete("/cart/{id}", d.Cart.Delete)

		auth.Post("/order", d.Order.Create)
		auth.Get("/my/orders", d.Order.ListMine)
		auth.Get("/order/{id}", d.Order.Get)
		auth.Get("/order_item/{order_id}", d.OrderItem.ListByOrder)

		// Менеджер
		auth.Group(func(mgr chi.Router) {
			mgr.Use(mw.RequireRole(models.RoleManager, models.RoleAdmin))
			mgr.Get("/order", d.Order.List)
			mgr.Put("/order/{id}", d.Order.Update)
			mgr.Post("/product", d.Product.Create)
			mgr.Put("/product/{id}", d.Product.Update)
			mgr.Delete("/product/{id}", d.Product.Delete)
			mgr.Post("/product/import", d.Product.Import)
			mgr.Post("/photo", d.Photo.Create)
			mgr.Delete("/photo/{id}", d.Photo.Delete)
		})

		// Только администратор
		auth.Group(func(adm chi.Router) {
			adm.Use(mw.RequireRole(models.RoleAdmin))
			adm.Post("/role", d.Role.Create)
			adm.Get("/role", d.Role.List)
			adm.Put("/role/{id}", d.Role.Update)
			adm.Delete("/role/{id}", d.Role.Delete)
			adm.Post("/user", d.User.Create)
			adm.Get("/user", d.User.List)
			adm.Get("/user/role/{id}", d.User.ListByRole)
			adm.Put("/user/{id}", d.User.Update)
			adm.Put("/user/{id}/role", d.User.UpdateRole)
			adm.Delete("/user/{id}", d.User.Delete)
			adm.Post("/order_item", d.OrderItem.Create)
			adm.Post("/user/{id}/balance", d.User.TopUp)
		})
	})
	return r
}
