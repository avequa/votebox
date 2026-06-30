package main

import (
	"context"
	"log"
)

// "проголосовать"
type VoteJob struct {
	PollID string
	OptionID string
}

// читает голоса из канала и применяет их к хранилищу по одному
type Worker struct {
	store *Store
	votes chan VoteJob
}

//конструктор
func NewWorker(store *Store) *Worker {
	return &Worker{
		store: store,
		votes: make(chan VoteJob, 100),
	}
}

// submit вызывается из http обработчика
func (w *Worker) Submit(job VoteJob) bool {
	select {
		case w.votes <- job:
			return true
		default:
			return false
	}
}

//цикл воркера , слушает :
// - новый голос из канала votes
// - сигнал остановки 
func (w *Worker) Run(ctx context.Context) {
	log.Println("worker start")

	for {
		select {
		case job := <-w.votes:
			if err := w.store.ApplyVote(job.PollID, job.OptionID); err != nil {
				log.Printf("apply vote method failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("worker end")
			return
		}
	}
}