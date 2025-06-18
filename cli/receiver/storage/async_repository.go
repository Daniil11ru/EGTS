package storage

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
)

type AsyncRepository struct {
	repo   *Repository
	ch     chan interface{ ToBytes() ([]byte, error) }
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewAsyncRepository(repo *Repository, buffer, workers int) *AsyncRepository {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	ctx, cancel := context.WithCancel(context.Background())
	ar := &AsyncRepository{
		repo:   repo,
		ch:     make(chan interface{ ToBytes() ([]byte, error) }, buffer),
		ctx:    ctx,
		cancel: cancel,
	}
	for i := 0; i < workers; i++ {
		ar.wg.Add(1)
		go ar.worker()
	}
	return ar
}

func (a *AsyncRepository) worker() {
	defer a.wg.Done()
	for {
		select {
		case msg, ok := <-a.ch:
			if !ok {
				return
			}
			if err := a.repo.Save(msg); err != nil {
				log.WithField("err", err).Error("Ошибка сохранения телеметрии")
			}
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *AsyncRepository) Save(m interface{ ToBytes() ([]byte, error) }) error {
	select {
	case a.ch <- m:
		return nil
	case <-a.ctx.Done():
		return fmt.Errorf("асинхронный репозиторий был закрыт")
	}
}

func (a *AsyncRepository) Close() {
	a.cancel()
	close(a.ch)
	a.wg.Wait()
}
