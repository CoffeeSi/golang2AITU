package repository

import (
	"context"
	"errors"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DoctorRepository struct {
	db *pgxpool.Pool
}

func NewDoctorRepository(db *pgxpool.Pool) DoctorRepository {
	return DoctorRepository{db: db}
}

func (r DoctorRepository) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	var id string
	err := r.db.QueryRow(ctx, "INSERT INTO doctors (id, full_name, specialization, email, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		doctor.ID,
		doctor.FullName,
		doctor.Specialization,
		doctor.Email,
		doctor.CreatedAt,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return model.DoctorAlreadyExistsError
			}
		}
		return err
	}
	doctor.ID = id
	return nil
}

func (r DoctorRepository) GetDoctorByID(ctx context.Context, id string) (*model.Doctor, error) {
	var doctor model.Doctor
	err := r.db.QueryRow(ctx, "SELECT id, full_name, specialization, email, created_at FROM doctors WHERE id = $1", id).Scan(
		&doctor.ID,
		&doctor.FullName,
		&doctor.Specialization,
		&doctor.Email,
		&doctor.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.DoctorNotFoundError
		}
		return nil, err
	}
	return &doctor, nil
}

func (r DoctorRepository) ListDoctors(ctx context.Context) ([]*model.Doctor, error) {
	var doctors []*model.Doctor
	rows, err := r.db.Query(ctx, "SELECT id, full_name, specialization, email, created_at FROM doctors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var doctor model.Doctor
		err := rows.Scan(
			&doctor.ID,
			&doctor.FullName,
			&doctor.Specialization,
			&doctor.Email,
			&doctor.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		doctors = append(doctors, &doctor)
	}
	return doctors, nil
}

func (r DoctorRepository) GetDoctorByEmail(ctx context.Context, email string) (*model.Doctor, error) {
	var doctor model.Doctor
	err := r.db.QueryRow(ctx, "SELECT id, full_name, specialization, email, created_at FROM doctors WHERE email = $1", email).Scan(
		&doctor.ID,
		&doctor.FullName,
		&doctor.Specialization,
		&doctor.Email,
		&doctor.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.DoctorNotFoundError
		}
		return nil, err
	}
	return &doctor, nil
}
