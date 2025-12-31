package app

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolManager struct {
	Pools map[uuid.UUID]*pgxpool.Pool
	mu    sync.RWMutex
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		Pools: make(map[uuid.UUID]*pgxpool.Pool),
	}
}

func (pm *PoolManager) AddPool(id uuid.UUID, c *pgx.ConnConfig) (*pgxpool.Pool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	poolConfig := &pgxpool.Config{
		ConnConfig: c,
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	pm.Pools[id] = pool
	return pool, nil
}

func (pm *PoolManager) GetPool(id uuid.UUID) (*pgxpool.Pool, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	pool, exists := pm.Pools[id]
	return pool, exists
}

func (pm *PoolManager) DeletePool(id uuid.UUID) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, exists := pm.Pools[id]
	if !exists {
		return errors.New("connection does not exist")
	}

	// Close the pool before deleting it
	pool.Close()

	// Remove the pool from the map
	delete(pm.Pools, id)

	return nil
}
