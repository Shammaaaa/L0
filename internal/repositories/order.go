package repositories

import (
	"context"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"

	"order/domain"
	"order/internal/models"
)

// OrderRepository отвечает за работу с сущностью Order в базе данных
type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(ctx context.Context, db *sqlx.DB) (*OrderRepository, error) {
	o := &OrderRepository{db: db}
	if err := o.migrate(ctx); err != nil {
		return nil, err
	}
	return o, nil
}

// migrate имитирует миграции в приложение, обычно это делается через файлы миграции
// в директории ./migrations
func (o *OrderRepository) migrate(ctx context.Context) error {
	const query = `create table if not exists public.order
(
    order_uid varchar(19) primary key not null,
    data      jsonb default '{}'::jsonb not null
);

create unique index if not exists uq_order_uid
    on public.order (order_uid);`

	if _, err := o.db.ExecContext(ctx, query); err != nil {
		return err
	}
	return nil
}

func (o *OrderRepository) Create(ctx context.Context, order *domain.Order) (int64, error) {
	const query = `INSERT INTO public.order (order_uid, data) VALUES (:id,:data)`
	result, err := o.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":   order.OrderUID,
		"data": order,
	})

	if err != nil {
		return 0, err
	}

	// получаем сколько строк изменилось/добавилось
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (o *OrderRepository) Get(ctx context.Context, id string) (domain.Order, error) {
	order := models.Order{}
	// выбираем только jsob поле data, т.к. идентификатор уже содержится внутри этой структуры
	err := o.db.GetContext(ctx, &order, "SELECT data FROM public.order WHERE order_uid=$1", id)
	if err != nil {
		return domain.Order{}, err
	}
	return order.Data, nil
}

func (o *OrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order

	// выбираем только jsob поле data, т.к. идентификатор уже содержится внутри этой структуры
	rows, err := o.db.QueryxContext(ctx, "SELECT data FROM public.order")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		order := models.Order{}
		err = rows.StructScan(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order.Data)
	}

	return orders, nil
}
