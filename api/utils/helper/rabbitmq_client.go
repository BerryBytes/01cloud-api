package helper

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type RabbitmqRoutine struct {
	conn            *amqp.Connection
	client          *amqp.Channel
	mu              sync.Mutex
	subscriptions   map[string]func(interface{}) error
	stackForPublish map[string]interface{}
	stackMu         sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewRabbitmqRoutine() (*RabbitmqRoutine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &RabbitmqRoutine{
		ctx:    ctx,
		cancel: cancel,
	}
	err := r.connect()
	if err != nil {
		return nil, err
	}
	go r.reconnectLoop()
	return r, nil
}

func (r *RabbitmqRoutine) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	r.conn = conn
	r.client = ch
	return nil
}

func (r *RabbitmqRoutine) reconnectLoop() {
	for {
		reason, ok := <-r.conn.NotifyClose(make(chan *amqp.Error))
		if ok {
			log.Info("RabbitMQ connection is OK")
			time.Sleep(1 * time.Minute)
			continue
		}
		log.Warnf("RabbitMQ connection closed, reason: %v", reason)
		for {
			err := r.connect()
			if err == nil {
				log.Info("Reconnected to RabbitMQ")
				time.Sleep(5 * time.Second)
				r.reSubscribe()
				if r.stackForPublish != nil {
					r.stackMu.Lock()
					for name, req := range r.stackForPublish {
						go func(name string, req interface{}) {
							err := r.Publish(name, req)
							if err != nil {
								log.Errorf("Failed to publish message in queue %s: %v", name, err)
							}
						}(name, req)
					}
					r.stackForPublish = nil
					r.stackMu.Unlock()
				}
				break
			}
			log.Warnf("Failed to reconnect to RabbitMQ: %v", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func (r *RabbitmqRoutine) reSubscribe() {
	log.Infof("Re-subscribing to queues")
	for name, method := range r.subscriptions {
		if method == nil {
			log.Warnf("Skipping re-subscription for queue %s: method is nil", name)
			continue
		}
		go func(name string, method func(interface{}) error) {
			select {
			case <-r.ctx.Done():
				log.Infof("Stopping re-subscription for queue %s", name)
				return
			default:
				err := r.Subscribe(name, method)
				if err != nil {
					log.Errorf("Failed to re-subscribe to queue %s: %v", name, err)
				} else {
					log.Infof("Re-subscribed to queue %s", name)
				}
			}
		}(name, method)
	}
}

func (r *RabbitmqRoutine) Subscribe(name string, method func(interface{}) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.subscriptions == nil {
		r.subscriptions = make(map[string]func(interface{}) error)
	}
	if r.subscriptions[name] == nil {
		r.subscriptions[name] = method
	}

	_, err := r.client.QueueDeclare(name, false, false, false, false, nil)
	if err != nil {
		return err
	}
	messages, err := r.client.Consume(name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range messages {
			log.Infof("Got message in queue :: %s", name)
			err := method(d.Body)
			if err != nil {
				log.Errorf("Error processing message: %v", err)
			}
		}
	}()

	return nil
}

func (r *RabbitmqRoutine) Publish(name string, req interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.client.QueueDeclare(name, false, false, false, false, nil)
	if err != nil {
		if err == amqp.ErrClosed {
			log.Warn("RabbitMQ connection closed")
			r.stackMu.Lock()
			if r.stackForPublish == nil {
				r.stackForPublish = make(map[string]interface{})
			}
			r.stackForPublish[name] = req
			log.Infof("stacking for publish %v", r.stackForPublish)
			r.stackMu.Unlock()
			return nil
		}
		return err
	}
	msg, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = r.client.Publish("", name, false, false, amqp.Publishing{Body: msg})
	if err != nil {
		if err == amqp.ErrClosed {
			log.Warn("RabbitMQ connection closed")
			r.stackMu.Lock()
			if r.stackForPublish == nil {
				r.stackForPublish = make(map[string]interface{})
			}
			r.stackForPublish[name] = req
			r.stackMu.Unlock()
			log.Infof("stacking for publish %v", r.stackForPublish)
			return nil
		}
		return err
	}
	log.Infof("Published a message in queue :: %s", name)
	log.Debug(string(msg))
	return nil
}

func (r *RabbitmqRoutine) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancel()

	if r.client != nil {
		r.client.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
