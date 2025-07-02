package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	storageCfg "github.com/pvxdv/self_improver/internal/config/storage"
	"github.com/pvxdv/self_improver/internal/model"
	"github.com/pvxdv/self_improver/internal/storage"
)

const (
	createTableTrendSQL = `CREATE TABLE IF NOT EXISTS trend (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE)`

	createTableChallengeSQL = `CREATE TABLE IF NOT EXISTS challenge (
		id SERIAL PRIMARY KEY,
		trend_id INTEGER NOT NULL REFERENCES trend(id) ON DELETE CASCADE,
		description VARCHAR(1000) NOT NULL,
		start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        end_date TIMESTAMP)`
)

type Storage struct {
	db     *pgx.Conn
	logger *zap.SugaredLogger
}

func New(ctx context.Context, cfg *storageCfg.Config, logger *zap.SugaredLogger) (*Storage, error) {
	conn, err := pgx.Connect(ctx, fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err = conn.Exec(ctx, createTableTrendSQL); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to create cluster table: %w", err)
	}

	if _, err = conn.Exec(ctx, createTableChallengeSQL); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to create container table: %w", err)
	}

	return &Storage{db: conn, logger: logger}, nil
}

func (s *Storage) Close(ctx context.Context) error {
	if err := s.db.Close(ctx); err != nil {
		s.logger.Errorf("failed to close database connection: %v", err)
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

func (s *Storage) SaveTrend(ctx context.Context, trend *model.Trend) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return 0, err
	}
	defer tx.Rollback(ctx)

	var trendID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO trend (name) 
         VALUES ($1) 
         RETURNING id`,
		trend.Name,
	).Scan(&trendID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			s.logger.Warnf("trend already exists: %s", trend.Name)
			return 0, storage.ErrTrendExists
		}
		s.logger.Errorf("failed to create trend: %v", err)
		return -1, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return 0, err
	}

	s.logger.Infof("successfully added trend %s (ID: %d)", trend.Name, trendID)

	return trendID, nil
}

func (s *Storage) DeleteTrend(ctx context.Context, name string) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return -1, err
	}
	defer tx.Rollback(ctx)

	var trendID int64
	err = tx.QueryRow(ctx,
		"SELECT id FROM trend WHERE name = $1", name).
		Scan(&trendID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warnf("trend not found: %s", name)
			return -1, storage.ErrTrendNotFound
		}
		s.logger.Errorf("failed to get trend ID: %v", err)
		return -1, err
	}

	_, err = tx.Exec(ctx,
		"DELETE FROM trend WHERE id = $1", trendID)
	if err != nil {
		s.logger.Errorf("failed to delete trend: %v", err)
		return -1, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return -1, err
	}

	s.logger.Infof("successfully deleted trend %s (ID: %d)", name, trendID)
	return trendID, nil
}

func (s *Storage) SaveChallenge(ctx context.Context, challenge *model.Challenge) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return 0, err
	}
	defer tx.Rollback(ctx)

	var trendID int64
	err = tx.QueryRow(ctx,
		"SELECT id FROM trend WHERE name = $1", challenge.TrendName).
		Scan(&trendID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warnf("trend not found: %s", challenge.TrendName)
			return -1, storage.ErrTrendNotFound
		}
		s.logger.Errorf("failed to get trend ID: %v", err)
		return -1, err
	}

	var challengeID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO challenge (trend_id, description, start_date) 
         VALUES ($1, $2, $3) 
         RETURNING id`,
		trendID,
		challenge.Description,
		challenge.StartDate,
	).Scan(&challengeID)

	if err != nil {
		s.logger.Errorf("failed to create chellenge: %v", err)
		return -1, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return 0, err
	}

	s.logger.Infof("successfully created challenge (ID: %d)", challengeID)

	return challengeID, nil
}

func (s *Storage) GetTrend(ctx context.Context, name string) (*model.Trend, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var trend model.Trend
	err = tx.QueryRow(ctx,
		`SELECT id, name
		FROM trend WHERE name = $1`, name,
	).Scan(&trend.ID, &trend.Name)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warnf("trend not found: %s", name)
			return nil, storage.ErrTrendNotFound
		}
		return nil, fmt.Errorf("failed to get trend: %w", err)
	}

	rows, err := tx.Query(ctx,
		`SELECT id, trend_id, description, start_date, end_date 
		FROM challenge WHERE trend_id = $1`, trend.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c model.Challenge
		if err = rows.Scan(
			&c.ID,
			&c.TrendID,
			&c.Description,
			&c.StartDate,
			&c.EndDate,
		); err != nil {
			s.logger.Warnf("failed to scan challenge: %v", err)
			continue
		}
		trend.Challenges = append(trend.Challenges, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("challenge rows error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &trend, nil
}
