package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	storageCfg "github.com/pvxdv/self_improver/internal/config/storage"
	"github.com/pvxdv/self_improver/internal/model"
	"github.com/pvxdv/self_improver/internal/storage"
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

func (s *Storage) AddDebt(ctx context.Context, debt *model.Debt) (int64, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var debtID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO debt (description, amount, return_date)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		debt.Description,
		debt.Amount,
		debt.ReturnDate,
	).Scan(&debtID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Errorf("database error: %v", pgErr)
		}
		return -1, fmt.Errorf("failed to insert debt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Infof("successfully added debt (ID: %d)", debtID)
	return debtID, nil
}

func (s *Storage) GetDebt(ctx context.Context, id int64) (model.Debt, error) {
	var debt model.Debt

	err := s.db.QueryRow(ctx,
		`SELECT id, description, amount
		 FROM debt WHERE id = $1`,
		id,
	).Scan(&debt.ID, &debt.Description, &debt.Amount)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warnf("debt not found (ID: %d)", id)
			return model.Debt{}, storage.ErrDebtNotFound
		}
		s.logger.Errorf("failed to get debt: %v", err)
		return model.Debt{}, fmt.Errorf("failed to get debt: %w", err)
	}

	return debt, nil
}

func (s *Storage) UpdateDebt(ctx context.Context, debt model.Debt) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx,
		`UPDATE debt 
		 SET description = $1, amount = $2, return_date= $3, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $4`,
		debt.Description,
		debt.Amount,
		debt.ReturnDate,
		debt.ID,
	)

	if err != nil {
		s.logger.Errorf("failed to update debt: %v", err)
		return fmt.Errorf("failed to update debt: %w", err)
	}

	if result.RowsAffected() == 0 {
		s.logger.Warnf("debt not found for update (ID: %d)", debt.ID)
		return storage.ErrDebtNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Infof("successfully updated debt (ID: %d)", debt.ID)
	return nil
}

func (s *Storage) DeleteDebt(ctx context.Context, id int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx,
		"DELETE FROM debt WHERE id = $1",
		id,
	)

	if err != nil {
		s.logger.Errorf("failed to delete debt: %v", err)
		return fmt.Errorf("failed to delete debt: %w", err)
	}

	if result.RowsAffected() == 0 {
		s.logger.Warnf("debt not found for deletion (ID: %d)", id)
		return storage.ErrDebtNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Infof("successfully deleted debt (ID: %d)", id)
	return nil
}

func (s *Storage) GetAmountDebt(ctx context.Context) (*model.AmountDebt, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Errorf("failed to begin transaction: %v", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var totalAmount int64
	err = tx.QueryRow(ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM debt",
	).Scan(&totalAmount)

	if err != nil {
		s.logger.Errorf("failed to get total debt amount: %v", err)
		return nil, fmt.Errorf("failed to get total debt amount: %w", err)
	}

	rows, err := tx.Query(ctx,
		"SELECT id, description, amount, return_date FROM debt",
	)
	if err != nil {
		s.logger.Errorf("failed to get debt details: %v", err)
		return nil, fmt.Errorf("failed to get debt details: %w", err)
	}
	defer rows.Close()

	debts := make([]*model.Debt, 0)
	for rows.Next() {
		var d model.Debt
		var date sql.NullTime

		if err := rows.Scan(&d.ID, &d.Description, &d.Amount, &date); err != nil {
			s.logger.Warnf("failed to scan debt row: %v", err)
			continue
		}

		if date.Valid {
			d.ReturnDate = &date.Time
		}

		debts = append(debts, &d)
	}

	if err := rows.Err(); err != nil {
		s.logger.Errorf("rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Errorf("failed to commit transaction: %v", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &model.AmountDebt{
		Amount:  totalAmount,
		Details: debts,
	}, nil
}
