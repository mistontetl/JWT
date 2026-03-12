package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go" // RabbitMQ
)

// TODO BASE LOCAL (REMOVE)!!!
var UUIDBase = make(map[string]string)

const (
	reconnectDelay = 5 * time.Second
	amqpURI        = "amqp://guest:guest@localhost:5672/"

	QueueName = "billing_queue"
	DLXName   = "billing_dlx" // Exchange for failed messages
	DLQName   = "billing_dlq" // Failed message queue

	RabbitMQPrefetchLimit = 100

	retryExchangeName = "billing_retry_exchange"
	retryQueueName    = "billing_wait_queue"
	retry_key         = "retry_key"
)

type RabbitMQDelivery struct {
	delivery amqp.Delivery
}

func NewRabbitMQDelivery(d amqp.Delivery) Delivery {
	return &RabbitMQDelivery{
		delivery: d,
	}
}

func (r *RabbitMQDelivery) Body() []byte {
	return r.delivery.Body
}

func (r *RabbitMQDelivery) Ack() error {
	return r.delivery.Ack(false)
}

// TRUE -- Reencolar!
func (r *RabbitMQDelivery) Nack(requeue bool) error {
	return r.delivery.Nack(false, requeue)
}

type ClientQueue struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	uri  string
	//closeChan chan *amqp.Error
}

func NewClient() (RSGQueue, error) {

	ctx, cancel := context.WithTimeout(context.Background(), reconnectDelay)
	defer cancel()

	// Conexion
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("The connection to RabbitMQ failed due to a timeout: %w", ctx.Err())
	}

	//Chanel

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	client := &ClientQueue{conn: conn, ch: ch}

	if err := client.SetupQueues(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func (c *ClientQueue) SetupQueues() error {
	// (DLX)
	if err := c.ch.ExchangeDeclare(DLXName, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// (DLQ)
	if _, err := c.ch.QueueDeclare(DLQName, true, false, false, false, nil); err != nil {
		return err
	}

	// Bindear DLQ al DLX
	if err := c.ch.QueueBind(DLQName, DLQName, DLXName, false, nil); err != nil {
		return err
	}

	// Waiting room
	c.ch.ExchangeDeclare(retryExchangeName, "direct", true, false, false, false, nil)

	//Redirect
	retryArgs := amqp.Table{
		"x-dead-letter-exchange":    "",        // Exchange por defecto
		"x-dead-letter-routing-key": QueueName, // Return Main Queue
	}

	c.ch.QueueDeclare(retryQueueName, true, false, false, false, retryArgs)
	c.ch.QueueBind(retryQueueName, retry_key, retryExchangeName, false, nil)

	// Main Queue (DLX)
	/*args := amqp.Table{
		"x-dead-letter-exchange": DLXName,
	}
	if _, err := c.ch.QueueDeclare(QueueName, true, false, false, false, args); err != nil {
		return err
	}*/

	// 3. Main Queue
	args := amqp.Table{"x-dead-letter-exchange": DLXName}
	c.ch.QueueDeclare(QueueName, true, false, false, false, args)

	return nil
}

func (c *ClientQueue) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *ClientQueue) Publish(ctx context.Context, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	publishCtx := ctx
	var cancel context.CancelFunc

	if _, deadlineSet := publishCtx.Deadline(); !deadlineSet {
		publishCtx, cancel = context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
	}

	err = c.ch.PublishWithContext(
		publishCtx,
		"",
		QueueName, // routing key (nombre de la cola)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})

	return err
}

func (c *ClientQueue) Consume(handler func([]byte) error) {
	// Qos(1) prevents a Worker from monopolizing messages.
	c.ch.Qos(1, 0, false)

	msgs, err := c.ch.Consume(
		QueueName,
		"",
		false,
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Failure to register consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				// NACK: Falló. No lo regresamos al final de la cola (requeue=false)
				// DLX lo atrape después de los reintentos.
				d.Nack(false, false)
			} else {
				// OK Éxito. Borrar el mensaje de la cola.
				d.Ack(false)
			}
		}
	}()

	log.Println(" [*] Worker escuchando en la cola.")

	<-forever
}

func (c *ClientQueue) ConsumeAsync(handler func(delivery Delivery)) error {
	// Qos(1) prevents a Worker from monopolizing messages.
	if err := c.ch.Qos(RabbitMQPrefetchLimit, 0, false); err != nil {
		return fmt.Errorf("fallo al establecer QoS/Prefetch: %w", err)
	}

	msgs, err := c.ch.Consume(
		QueueName,
		"",
		false, // autoAck: false,
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Bucle!!!
	forever := make(chan bool)

	go func() {
		// This goroutine is fast, it only reads from the RabbitMQ (msgs) channel
		for d := range msgs {
			handler(NewRabbitMQDelivery(d))
		}
	}()

	<-forever
	return nil
}

func (c *ClientQueue) PublishWithRetry(ctx context.Context, payload interface{}, delaySeconds int) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Convertimos segundos a milisegundos para RabbitMQ
	expirationMs := strconv.Itoa(delaySeconds * 1000)

	return c.ch.PublishWithContext(
		ctx,
		retryExchangeName, //
		retry_key,         // Routing key to "Wait Room"
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Expiration:   expirationMs, // <-- El mensaje "morirá" después de este tiempo
		},
	)
}
