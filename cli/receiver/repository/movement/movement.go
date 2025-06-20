package movement

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/daniil11ru/egts/cli/receiver/repository/movement/source"
	log "github.com/sirupsen/logrus"
)

type SaveRequest struct {
	message   interface{ ToBytes() ([]byte, error) }
	vehicleID int
}

type VehicleMovementRepository struct {
	repo   *source.Repository
	ch     chan SaveRequest
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewVehicleMovementRepository(repo *source.Repository, buffer, workers int) *VehicleMovementRepository {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	ctx, cancel := context.WithCancel(context.Background())
	ar := &VehicleMovementRepository{
		repo:   repo,
		ch:     make(chan SaveRequest, buffer),
		ctx:    ctx,
		cancel: cancel,
	}
	for i := 0; i < workers; i++ {
		ar.wg.Add(1)
		go ar.worker()
	}
	return ar
}

func (a *VehicleMovementRepository) worker() {
	defer a.wg.Done()
	for {
		select {
		case req, ok := <-a.ch:
			if !ok {
				return
			}
			if err := a.repo.Save(req.message, req.vehicleID); err != nil {
				log.WithField("err", err).Error("Ошибка сохранения телеметрии")
			}
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *VehicleMovementRepository) Save(message interface{ ToBytes() ([]byte, error) }, vehicleID int) error {
	select {
	case a.ch <- SaveRequest{message: message, vehicleID: vehicleID}:
		return nil
	case <-a.ctx.Done():
		return fmt.Errorf("асинхронный репозиторий был закрыт")
	}
}

func (a *VehicleMovementRepository) Close() {
	a.cancel()
	close(a.ch)
	a.wg.Wait()
}
