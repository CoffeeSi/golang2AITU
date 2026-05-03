package repository

import (
	"context"
	"errors"

	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentRepository struct {
	db *pgxpool.Pool
}

func NewAppointmentRepository(db *pgxpool.Pool) AppointmentRepository {
	return AppointmentRepository{db: db}
}

func (r AppointmentRepository) CreateAppointment(ctx context.Context, appointment *model.Appointment) error {
	var id string
	err := r.db.QueryRow(
		ctx,
		"INSERT INTO appointments (id, title, description, doctor_id, status, created_at, updated_at) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		appointment.ID,
		appointment.Title,
		appointment.Description,
		appointment.DoctorID,
		appointment.Status,
		appointment.CreatedAt,
		appointment.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func (r AppointmentRepository) GetAppointment(ctx context.Context, id string) (*model.Appointment, error) {
	var appointment model.Appointment
	err := r.db.QueryRow(ctx, "SELECT id, title, description, doctor_id, status, created_at, updated_at FROM appointments WHERE id = $1", id).Scan(
		&appointment.ID,
		&appointment.Title,
		&appointment.Description,
		&appointment.DoctorID,
		&appointment.Status,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.AppointmentNotFoundError
		}
		return nil, err
	}
	return &appointment, nil
}

func (r AppointmentRepository) ListAppointments(ctx context.Context) ([]*model.Appointment, error) {
	var appointments []*model.Appointment
	rows, err := r.db.Query(ctx, "SELECT id, title, description, doctor_id, status, created_at, updated_at FROM appointments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var appointment model.Appointment
		err := rows.Scan(
			&appointment.ID,
			&appointment.Title,
			&appointment.Description,
			&appointment.DoctorID,
			&appointment.Status,
			&appointment.CreatedAt,
			&appointment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, &appointment)
	}
	return appointments, nil
}

func (r AppointmentRepository) UpdateAppointmentStatus(ctx context.Context, id string, status string) (*model.Appointment, error) {
	command_tag, err := r.db.Exec(ctx, "UPDATE appointments SET status = $2, updated_at = NOW() WHERE id = $1", id, status)
	if err != nil {
		return nil, err
	}
	if command_tag.RowsAffected() == 0 {
		return nil, model.AppointmentNotFoundError
	}

	var appointment model.Appointment
	err = r.db.QueryRow(ctx, "SELECT id, title, description, doctor_id, status, created_at, updated_at FROM appointments WHERE id = $1", id).Scan(
		&appointment.ID,
		&appointment.Title,
		&appointment.Description,
		&appointment.DoctorID,
		&appointment.Status,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &appointment, nil
}
