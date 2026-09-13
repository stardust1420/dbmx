package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgconn/ctxwatch"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolManager struct {
	mu          sync.RWMutex
	Pools       map[uuid.UUID]*pgxpool.Pool
	ActiveConns map[int64]int64
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		Pools:       make(map[uuid.UUID]*pgxpool.Pool),
		ActiveConns: make(map[int64]int64),
	}
}

func (pm *PoolManager) AddPool(id uuid.UUID, c *pgx.ConnConfig, connID int64) (*pgxpool.Pool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	poolConfig, err := pgxpool.ParseConfig(c.ConnString())
	if err != nil {
		return nil, err
	}
	poolConfig.ConnConfig = c

	// Cancelling a query has to reach the server. pgx defaults to
	// DeadlineContextWatcherHandler, which only sets a deadline on the socket:
	// the client returns immediately with a cancellation error while postgres
	// keeps executing the query to completion, holding the connection and
	// burning CPU. A backend doing real work never looks at its socket, so
	// nothing interrupts it. CancelRequestContextWatcherHandler instead sends a
	// genuine cancel request, which stops the query and leaves the connection
	// reusable.
	//
	// The delays are pgx's own: cancelling is inherently racy, so give a query
	// that is about to finish on its own a moment to do so, and fall back to the
	// socket deadline if the server never answers the cancel request.
	poolConfig.ConnConfig.BuildContextWatcherHandler = func(pgConn *pgconn.PgConn) ctxwatch.Handler {
		return &pgconn.CancelRequestContextWatcherHandler{
			Conn:               pgConn,
			CancelRequestDelay: 250 * time.Millisecond,
			DeadlineDelay:      5 * time.Second,
		}
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	pm.Pools[id] = pool

	// Add the connection ID to the map
	// This is used to track the count of the number of active pools for a connection
	pm.ActiveConns[connID] += 1

	return pool, nil
}

func (pm *PoolManager) GetPool(id uuid.UUID) (*pgxpool.Pool, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	pool, exists := pm.Pools[id]
	return pool, exists
}

func (pm *PoolManager) DeletePool(id uuid.UUID, connID int64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if the connection ID exists
	if _, exists := pm.ActiveConns[connID]; !exists {
		return errors.New("connection does not exist")
	}

	pool, exists := pm.Pools[id]
	if !exists {
		return errors.New("connection does not exist")
	}

	// Close the pool before deleting it
	pool.Close()

	// Remove the pool from the pools map
	delete(pm.Pools, id)

	// Get the number of active pools for the connection
	activeConns := pm.ActiveConns[connID]

	// If the number of active pools is 0, remove the connection ID from the map
	if activeConns-1 == 0 {
		delete(pm.ActiveConns, connID)
	} else {
		pm.ActiveConns[connID] -= 1
	}

	return nil
}
