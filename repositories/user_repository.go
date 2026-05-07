package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, created_at, updated_at
		 FROM users WHERE username=$1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	

	if rg.Valid {
		u.RG = &rg.String
	}
	if gender.Valid {
		u.Gender = &gender.String
	}
	if contactEmail.Valid {
		u.ContactEmail = &contactEmail.String
	}
	if fullAddress.Valid {
		u.FullAddress = &fullAddress.String
	}
	if contactPhone.Valid {
		u.ContactPhone = &contactPhone.String
	}
	if notes.Valid {
		u.Notes = &notes.String
	}
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	
	return &u, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, created_at, updated_at
		 FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	

	if rg.Valid {
		u.RG = &rg.String
	}
	if gender.Valid {
		u.Gender = &gender.String
	}
	if contactEmail.Valid {
		u.ContactEmail = &contactEmail.String
	}
	if fullAddress.Valid {
		u.FullAddress = &fullAddress.String
	}
	if contactPhone.Valid {
		u.ContactPhone = &contactPhone.String
	}
	if notes.Valid {
		u.Notes = &notes.String
	}
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	
	return &u, nil
}

func (r *UserRepository) Create(u *models.User) (*models.User, error) {
	if u.Permissions == nil {
		u.Permissions = []string{}
	}
	err := r.db.QueryRow(
		`INSERT INTO users (username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 RETURNING id, created_at, updated_at`,
		u.Username, u.PasswordHash, u.Role, pq.Array(u.Permissions), u.FullName, u.CPF, u.RG, u.BirthDate, u.Gender, u.Status,
		u.ContactEmail, u.FullAddress, u.ContactPhone, u.Notes,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return u, nil
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	rows, err := r.db.Query(
		`SELECT id, username, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, created_at, updated_at
		 FROM users ORDER BY id`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
		var birthDate sql.NullTime
		
		err := rows.Scan(&u.ID, &u.Username, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
			&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes,
			&u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if rg.Valid {
			u.RG = &rg.String
		}
		if gender.Valid {
			u.Gender = &gender.String
		}
		if contactEmail.Valid {
			u.ContactEmail = &contactEmail.String
		}
		if fullAddress.Valid {
			u.FullAddress = &fullAddress.String
		}
		if contactPhone.Valid {
			u.ContactPhone = &contactPhone.String
		}
		if notes.Valid {
			u.Notes = &notes.String
		}
		if birthDate.Valid {
			u.BirthDate = &birthDate.Time
		}
		if u.Permissions == nil {
			u.Permissions = []string{}
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Update(id int, u *models.User) (*models.User, error) {
	if u.Permissions == nil {
		u.Permissions = []string{}
	}
	err := r.db.QueryRow(
		`UPDATE users SET username=$1, role=$2, permissions=$3, full_name=$4, cpf=$5, rg=$6, birth_date=$7, gender=$8,
		 status=$9, contact_email=$10, full_address=$11, contact_phone=$12, notes=$13, updated_at=NOW()
		 WHERE id=$14 RETURNING updated_at`,
		u.Username, u.Role, pq.Array(u.Permissions), u.FullName, u.CPF, u.RG, u.BirthDate, u.Gender, u.Status,
		u.ContactEmail, u.FullAddress, u.ContactPhone, u.Notes, id,
	).Scan(&u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	u.ID = id
	return u, nil
}

func (r *UserRepository) UpdatePassword(id int, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, passwordHash, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) GetByCPF(cpf string) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, created_at, updated_at
		 FROM users WHERE cpf=$1`,
		cpf,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	

	if rg.Valid {
		u.RG = &rg.String
	}
	if gender.Valid {
		u.Gender = &gender.String
	}
	if contactEmail.Valid {
		u.ContactEmail = &contactEmail.String
	}
	if fullAddress.Valid {
		u.FullAddress = &fullAddress.String
	}
	if contactPhone.Valid {
		u.ContactPhone = &contactPhone.String
	}
	if notes.Valid {
		u.Notes = &notes.String
	}
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	
	return &u, nil
}

func (r *UserRepository) SearchByFilter(filter string) ([]models.User, error) {
	filterPattern := "%" + filter + "%"
	rows, err := r.db.Query(
		`SELECT id, username, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, created_at, updated_at
		 FROM users
		 WHERE LOWER(full_name) LIKE LOWER($1)
		    OR LOWER(username) LIKE LOWER($1)
		    OR cpf LIKE $1
		 ORDER BY id`,
		filterPattern)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
		var birthDate sql.NullTime

		err := rows.Scan(&u.ID, &u.Username, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
			&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes,
			&u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if rg.Valid {
			u.RG = &rg.String
		}
		if gender.Valid {
			u.Gender = &gender.String
		}
		if contactEmail.Valid {
			u.ContactEmail = &contactEmail.String
		}
		if fullAddress.Valid {
			u.FullAddress = &fullAddress.String
		}
		if contactPhone.Valid {
			u.ContactPhone = &contactPhone.String
		}
		if notes.Valid {
			u.Notes = &notes.String
		}
		if birthDate.Valid {
			u.BirthDate = &birthDate.Time
		}
		if u.Permissions == nil {
			u.Permissions = []string{}
		}
		users = append(users, u)
	}
	return users, nil
}
