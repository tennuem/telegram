package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Handler interface {
	Serve(*ResponseWriter, *Request)
}

type ResponseWriter struct {
	tgbotapi.MessageConfig
}

type Request struct {
	tgbotapi.Update
}

func ListenAndServe(token string, handler Handler) error {
	svr, close, err := NewServer(token)
	if err != nil {
		return err
	}
	defer close()

	return svr.Serve(handler)
}

func NewServer(token string) (*Server, func() error, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, nil, err
	}
	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = 60
	updates, err := api.GetUpdatesChan(updateCfg)
	if err != nil {
		return nil, nil, err
	}
	svr := &Server{
		api:     api,
		updates: updates,
		close:   make(chan struct{}, 1),
	}
	close := func() error {
		svr.close <- struct{}{}
		return nil
	}
	return svr, close, nil
}

type Server struct {
	api     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	close   chan struct{}
}

func (s *Server) Serve(handler Handler) error {
	if handler == nil {
		handler = DefaultServeMux
	}
	for {
		select {
		case <-s.close:
			return fmt.Errorf("shutdown server")
		case u := <-s.updates:
			if u.Message == nil { // ignore non-Message updates
				continue
			}
			resp := &ResponseWriter{tgbotapi.NewMessage(u.Message.Chat.ID, "")}
			req := &Request{u}
			handler.Serve(resp, req)
			s.api.Send(resp)
		}
	}
}
