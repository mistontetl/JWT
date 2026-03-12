package worker

import "portal_autofacturacion/data/queue"

type WorkerDataSource interface {
	HandleDelivery(delivery queue.Delivery)

	StartConsuming(consumer queue.Consumer) error
}
