package quotes

import "context"

type Quote interface {
	Get(context.Context) ([]byte, error)
}
