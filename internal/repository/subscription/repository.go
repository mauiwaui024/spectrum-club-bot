package subscription

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type subscriptionRepository struct {
	db *sqlx.DB
}

func NewSubscriptionRepository(db *sqlx.DB) repository.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

//todo:implement
func (r *subscriptionRepository) Create(subscription *models.Subscription) error {
	query := `
        INSERT INTO spectrum.subscriptions 
        (student_id, start_date, end_date, total_lessons, remaining_lessons, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	return r.db.QueryRow(
		query,
		subscription.StudentID,
		subscription.StartDate,
		subscription.EndDate,
		subscription.TotalLessons,
		subscription.RemainingLessons,
		subscription.CreatedAt,
	).Scan(&subscription.ID)
}

// В SubscriptionRepository
func (r *subscriptionRepository) Delete(id int64) error {
	query := `DELETE FROM spectrum.subscriptions WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("абонемент с ID %d не найден", id)
	}

	return nil
}

func (r *subscriptionRepository) GetByStudentID(studentID int64) ([]*models.Subscription, error) {
	query := `
        SELECT id, student_id, start_date, end_date, total_lessons, remaining_lessons, created_at
        FROM spectrum.subscriptions
        WHERE student_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription
	for rows.Next() {
		var subscription models.Subscription
		err := rows.Scan(
			&subscription.ID,
			&subscription.StudentID,
			&subscription.StartDate,
			&subscription.EndDate,
			&subscription.TotalLessons,
			&subscription.RemainingLessons,
			&subscription.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) GetAll() ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	query := `SELECT * FROM spectrum.subscriptions ORDER BY start_date`

	err := r.db.Select(&subscriptions, query)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// func (r *subscriptionRepository) GetByStudentID(studentID int64) (*models.Subscription, error) {
// 	var subscription models.Subscription
// 	query := `SELECT * FROM spectrum.subscriptions WHERE student_id = $1 ORDER BY created_at DESC LIMIT 1`
// 	err := r.db.Get(&subscription, query, studentID)
// 	return &subscription, err
// }

func (r *subscriptionRepository) GetByID(id int64) (*models.Subscription, error) {
	var subscription models.Subscription
	query := `SELECT * FROM spectrum.subscriptions WHERE id = $1`
	err := r.db.Get(&subscription, query, id)
	return &subscription, err
}

//	func (r *subscriptionRepository) GetActiveByStudentID(studentID int64) (*models.Subscription, error) {
//		var subscription models.Subscription
//		query := `SELECT * FROM spectrum.subscriptions WHERE student_id = $1 AND is_active = true ORDER BY created_at DESC LIMIT 1`
//		err := r.db.Get(&subscription, query, studentID)
//		return &subscription, err
//	}
func (r *subscriptionRepository) GetActiveByStudentID(studentID int64) (*models.Subscription, error) {
	var subscription models.Subscription
	query := `
		SELECT * FROM spectrum.subscriptions 
		WHERE student_id = $1 
		AND remaining_lessons > 0 
		AND (end_date IS NULL OR end_date > CURRENT_TIMESTAMP)
		ORDER BY created_at DESC 
		LIMIT 1`

	err := r.db.Get(&subscription, query, studentID)
	return &subscription, err
}

func (r *subscriptionRepository) GetHistoryByStudentID(studentID int64) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	query := `SELECT * FROM spectrum.subscriptions WHERE student_id = $1 ORDER BY created_at DESC`
	err := r.db.Select(&subscriptions, query, studentID)
	return subscriptions, err
}

func (r *subscriptionRepository) Update(subscription *models.Subscription) error {
	// query := `
	// 	UPDATE spectrum.subscriptions
	// 	SET used_lessons = $1, is_active = $2, end_date = $3
	// 	WHERE id = $4
	// `
	// _, err := r.db.Exec(
	// 	query,
	// 	subscription.UsedLessons,
	// 	subscription.IsActive,
	// 	subscription.EndDate,
	// 	subscription.ID,
	// )
	return nil
}
