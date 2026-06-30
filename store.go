package main

import (
	"errors"
	"sync"
	"time"
	"crypto/rand"
	"encoding/hex"
)

 // доменные ошибки
var (
	ErrPollNotFound = errors.New("poll not found")
	ErrOptionNotFound = errors.New("option not found")
)

// модели

// вариант ответа
type Option struct {
	ID string `json:"id"`
	Text string `json:"text"`
	Votes int `json:"votes"`
}

// опрос 
type Poll struct {
	ID string `json:"id"`
	Question string `json:"question"`
	Options []*Option `json:"options"`
	CreatedAt time.Time `json:"created_at"`
}


// хранилище опросов

type Store struct {
	mu sync.RWMutex
	polls map[string]*Poll
}

//конструктор
func NewStore() *Store {
	return &Store{
		polls: make(map[string]*Poll),
	}
}


// методы

// создание опроса - create poll
func (s *Store) CreatePoll(question string, optionTexts []string) *Poll {

	options := make([]*Option, 0, len(optionTexts))

	for _, text := range optionTexts {
		options = append(options, &Option{
			ID: newID(),
			Text: text,
		})
	}

	poll := &Poll{
		ID: newID(),
		Question: question,
		Options: options,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.polls[poll.ID] = poll
	return poll
}

// возвращает опрос - get poll
func (s *Store) GetPoll(id string) (*Poll, error) {
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.polls[id]
	if !ok {
		return nil, ErrPollNotFound
	}

	return p, nil
}

// возвращаем все опросы - list poll
func (s *Store) ListPolls() []*Poll {

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Poll, 0, len(s.polls))

	for _, value := range s.polls {
		out = append(out, value)
	}
	return out
}

// применяем один голос
func (s *Store) ApplyVote(pollID, optionID string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	poll, ok := s.polls[pollID]
	if !ok {
		return ErrPollNotFound
	}

	for _, value := range poll.Options {
		if value.ID == optionID {
			value.Votes++
			return nil
		}
	}
	return ErrOptionNotFound

}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}