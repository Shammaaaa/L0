package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"

	"order/domain"
	"order/pkg/nats"
)

type Cache interface {
	Set(ctx context.Context, key string, value domain.Order, ttl time.Duration) error
	Get(ctx context.Context, key string) (domain.Order, bool, error)
	Has(_ context.Context, key string) bool
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (int64, error)
	Get(ctx context.Context, id string) (domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
}

// Handler является некоторой оберткой над http интерфейсом, инкапсулирующий
// внутри себя логику приема запросов от внешнего мира.
// В целом можно сделать, используя, CQRS, но у нас только одна ручка создания записи Create.
//
// Вот как бы это выглядело в CQRS:
//  - handler принимал бы querieries и commands
//  - listQuerier(ctx, args listArgs)
//  - getQuerier(ctx, args getArgs)
//  - createCommandHandler(ctx, cmd createCommand)
type Handler struct {
	orderRepository OrderRepository
	cache           Cache

	nats *nats.Client
}

func NewHandler(
	orderRepository OrderRepository,
	cache Cache,
	nats *nats.Client,
) *Handler {
	h := &Handler{
		orderRepository: orderRepository,
		cache:           cache,
		nats:            nats,
	}
	return h
}

func (h *Handler) MountRoutes(app *fiber.App) {
	app.Get("/list", h.list)

	// example routes:
	// http://localhost:3000/api/v1/all
	// http://localhost:3000/api/v1/get/341
	v1 := app.Group("/api/v1")
	v1.Get("/list", h.listJSON)
	v1.Get("/get/:id", h.get)
	v1.Post("/create", h.create)
	v1.Post("/publish", h.publish)
}

// list ручка отображает страничку с записами
func (h *Handler) list(ctx *fiber.Ctx) error {
	// сюда также можно добавить пагинацию в виде limit & offset
	all, err := h.orderRepository.List(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.Render("index", fiber.Map{
		"orders": all,
	})
}

// listJSON ручка получения списка записей в формате JSON
func (h *Handler) listJSON(ctx *fiber.Ctx) error {
	all, err := h.orderRepository.List(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(all)
}

// create ручка создания записи
func (h *Handler) create(ctx *fiber.Ctx) error {
	request := &domain.Order{}
	// пробуем считать тело http запроса, который отправил нам клиент
	// в данном случае сервер сам понимает, какой тип данных нам пришел
	// на основе залоговка Content-Type, в нашем случае это application/json
	if err := ctx.BodyParser(request); err != nil {
		return err
	}

	// далее сохраняем запись в базу данных
	affected, err := h.orderRepository.Create(ctx.Context(), request)
	if err != nil {
		return err
	}

	// возвращаем клиенту ответ, сколько строк было сохранено
	return ctx.JSON(map[string]interface{}{
		"rows_affected": affected,
	})
}

// get ручка получения записи по идентификатору
func (h *Handler) get(ctx *fiber.Ctx) error {
	key := ctx.Params("id")
	if key == "" {
		return ctx.JSON(map[string]string{
			"error": "empty key",
		})
	}

	// если в кеше есть значение, то сразу же возвращаем его,
	// без необходимости читать из базы
	if h.cache.Has(ctx.Context(), key) {
		order, _, _ := h.cache.Get(ctx.Context(), key)
		return ctx.JSON(map[string]interface{}{
			"order": order,
		})
	}

	// иначе, если данных в кеше не оказалось, получаем запись из базы по идентификатору
	order, err := h.orderRepository.Get(ctx.Context(), key)
	if err != nil {
		return err
	}

	// далее сохраняем в кеше на один час (можно настроить)
	_ = h.cache.Set(ctx.Context(), key, order, time.Hour)

	// и возвращаем, только что сформированный ответ клиенту
	return ctx.JSON(map[string]interface{}{
		"order": order,
	})
}

// publish данная ручка позволяет сохранять запись в базе не на прямую, а через очередь сообщений
// в нашем случае это Nats-Streaming
// эта ручка просто иммитация обычного shell скрипта, который бы отправлял данные
// на прямую, непосредственно сервер Nats
func (h *Handler) publish(ctx *fiber.Ctx) error {
	request := &domain.Order{}
	if err := ctx.BodyParser(request); err != nil {
		return err
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if err = h.nats.Publish("test_topic", bytes); err != nil {
		return err
	}

	return ctx.JSON(map[string]interface{}{
		"error": "",
	})
}
