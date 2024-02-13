package nats

import (
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// Client небольшая обертка над пакетом Nats для удобства пользования
// здесь мы можем создавать подписчиков и удалять их
type Client struct {
	conn          *nats.Conn
	mutex         *sync.RWMutex
	subscriptions map[string]*nats.Subscription
}

func (c *Client) Publish(subj string, data []byte) error {
	return c.conn.Publish(subj, data)
}

func (c *Client) Subscribe(topic string, cb nats.MsgHandler) error {
	s, err := c.conn.Subscribe(topic, cb)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.subscriptions[topic] = s
	c.mutex.Unlock()

	return nil
}

func (c *Client) Unsubscribe(topic string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.subscriptions[topic]; ok {
		err := c.subscriptions[topic].Unsubscribe()
		if err != nil {
			return err
		}
		delete(c.subscriptions, topic)
	}
	return nil
}

func (c *Client) Close() error {
	c.mutex.Lock()
	for key := range c.subscriptions {
		if err := c.subscriptions[key].Unsubscribe(); err != nil {
			c.mutex.Unlock()
			return err
		}
	}
	c.mutex.Unlock()

	if err := c.conn.Drain(); err != nil {
		return err
	}
	c.conn.Close()
	return nil
}

func New(url string) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			fmt.Printf("Got disconnected! Reason: %q\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:          nc,
		mutex:         &sync.RWMutex{},
		subscriptions: map[string]*nats.Subscription{},
	}, nil
}
