package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"

	"order/domain"
	"order/internal/repositories"
	"order/internal/server"
	"order/pkg/cache"
	natsLocal "order/pkg/nats"
)

// OrderRepository чтобы не завязываться на конкретной реализации
// объявляем интерфейс по работе с заказами тут
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (int64, error)
}

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}

// Main небольшая функция обертка для точки входа main, чтобы было удобнее обрабатывать ошибки
func Main() error {
	// создаем основной контекст жизненного цила приложеия
	ctx, cancel := context.WithCancel(context.Background())

	// как только мы получим команду Ctrl+C мы завершим контекст и все зависимые от данного контекста комоненты также
	// автоматически отменятся или завершатся.
	// также у нас есть функции defer с Close() методами, которые закрывают все активные ресурсы
	defer cancel()

	// полуаем данные из переменных сред/окржения
	// примеры можно посомтреть в .env файле проекта
	o := opt{
		host: os.Getenv("PG_HOST"),
		user: os.Getenv("PG_USER"),
		pass: os.Getenv("PG_PASS"),
		port: os.Getenv("PG_PORT"),
		name: os.Getenv("PG_NAME"),
	}

	// подключаемся к базе данных Postgres
	db, err := sqlx.Open("postgres", o.ConnectionString())
	if err != nil {
		return err
	}

	defer func() {
		_ = db.Close()
	}()

	// тут можно настроить параметры подключения к базе
	db.SetMaxOpenConns(10)
	db.SetMaxOpenConns(12)

	// тут уже подключаем саму реализацию репозитория
	repo, err := repositories.NewOrderRepository(ctx, db)
	if err != nil {
		return err
	}

	// инициализиурем наш клиент Nats
	natsClient, err := natsLocal.New(os.Getenv("NATS_URL"))
	if err != nil {
		return err
	}

	defer func() {
		// тут мы очищаем ненужные нам данные, подписчиков и т.п., обрываем соединение с Nats
		_ = natsClient.Close()
	}()

	// нужно, чтобы можно было выйти из приложения по команде
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// подписываемся на топик в Nats-Streaming в отдельной горутине, чтобы нчиего не блокировать
	go func() {
		// подписываемся на токи test_topic
		err = natsClient.Subscribe("test_topic", func(msg *nats.Msg) {
			if err = handleEvent(ctx, repo, msg.Data); err != nil {
				log.Println(err)
			}
		})
		if err != nil {
			log.Printf("failed to subscribe Nats-Streaming: %s\n", err)

			// завершаем работу приложения принудительно, т.к. без компонента Nats наше приложение не сможет
			// выполнять всю возложенную на него работу
			sig <- syscall.SIGINT
		}
	}()

	// случаем tcp интерфейс
	ln, err := net.Listen(fiber.NetworkTCP4, os.Getenv("HTTP_ADDRESS"))
	if err != nil {
		return fmt.Errorf("failed to get http listener: %w", err)
	}

	// запускаем сервер в отдельной горутине
	go func() {
		handler := server.NewHandler(repo, cache.NewInMemory(), natsClient)

		app := fiber.New(fiber.Config{
			Views:        html.New("./templates", ".html"),
			ServerHeader: "Order Server",
		})

		handler.MountRoutes(app)

		if err = app.Listener(ln); err != nil {
			log.Printf("failed to start http server: %s\n", err)
			sig <- syscall.SIGINT
		}
	}()

	// на данном этапе у нас main горутина не блокируется и мы спокойно дожидаемся пользовательских команд,
	// сожидаемся пользовательского завершения приложения через Stop, Ctrl+C
	// или когда может возникнуть ошибка выше тогда мы сами посылаем сигнал на завершение
	<-sig

	return nil
}

func handleEvent(ctx context.Context, repo OrderRepository, data []byte) error {
	request := &domain.Order{}
	if err := json.Unmarshal(data, request); err != nil {
		return fmt.Errorf("failed to unmarshal input json: %w", err)
	}

	_, err := repo.Create(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to save data to database: %w", err)
	}
	return nil
}
