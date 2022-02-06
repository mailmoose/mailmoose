package handlers

import (
	"github.com/gopistolet/smtp/smtp"
)

/**
 * Handler is an interface for handler mechanisms.
 *
 * A handler is a struct on which the 'Handle' method can be called with an SMTP state
 */
type Handler interface {
	Handle(state *smtp.State) error
}

/**
 * HandlerMechanism contains a list of all handlers and executes the chain
 * it is meant to be passed to the MTA as mta.Handler interface
 */
type HandlerMachanism struct {
	Handlers []Handler
}

func (h *HandlerMachanism) Handle(state *smtp.State) error {
	for _, handler := range h.Handlers {
		err := handler.Handle(state)
		if err != nil {
			return err
		}
	}
	return nil
}
