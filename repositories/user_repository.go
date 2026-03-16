package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername busca um usuário pelo username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime
	
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, full_name, cpf, rg, birth_date, gender, status, 
		 contact_email, full_address, contact_phone, notes, created_at, updated_at 
		 FROM users WHERE username=$1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.FullName, &u.CPF, &rg, &birthDate, 
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	
	// Converter campos nullable
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

// GetByID busca um usuário pelo ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime
	
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, full_name, cpf, rg, birth_date, gender, status, 
		 contact_email, full_address, contact_phone, notes, created_at, updated_at 
		 FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.FullName, &u.CPF, &rg, &birthDate, 
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	
	// Converter campos nullable
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

// Create cria um novo usuário
func (r *UserRepository) Create(u *models.User) (*models.User, error) {
	err := r.db.QueryRow(
		`INSERT INTO users (username, password_hash, role, full_name, cpf, rg, birth_date, gender, status, 
		 contact_email, full_address, contact_phone, notes) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		 RETURNING id, created_at, updated_at`,
		u.Username, u.PasswordHash, u.Role, u.FullName, u.CPF, u.RG, u.BirthDate, u.Gender, u.Status,
		u.ContactEmail, u.FullAddress, u.ContactPhone, u.Notes,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return u, nil
}

// GetAll retorna todos os usuários (sem senha)
func (r *UserRepository) GetAll() ([]models.User, error) {
	rows, err := r.db.Query(
		`SELECT id, username, role, full_name, cpf, rg, birth_date, gender, status, 
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
		
		err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.FullName, &u.CPF, &rg, &birthDate, 
			&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, 
			&u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		// Converter campos nullable
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
		
		users = append(users, u)
	}
	return users, nil
}

// Update atualiza um usuário existente
func (r *UserRepository) Update(id int, u *models.User) (*models.User, error) {
	err := r.db.QueryRow(
		`UPDATE users SET username=$1, role=$2, full_name=$3, cpf=$4, rg=$5, birth_date=$6, gender=$7, 
		 status=$8, contact_email=$9, full_address=$10, contact_phone=$11, notes=$12, updated_at=NOW() 
		 WHERE id=$13 RETURNING updated_at`,
		u.Username, u.Role, u.FullName, u.CPF, u.RG, u.BirthDate, u.Gender, u.Status,
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

// UpdatePassword atualiza apenas a senha do usuário
func (r *UserRepository) UpdatePassword(id int, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, passwordHash, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Delete remove um usuário
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

// GetByCPF busca um usuário pelo CPF
func (r *UserRepository) GetByCPF(cpf string) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime
	
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, full_name, cpf, rg, birth_date, gender, status, 
		 contact_email, full_address, contact_phone, notes, created_at, updated_at 
		 FROM users WHERE cpf=$1`,
		cpf,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.FullName, &u.CPF, &rg, &birthDate, 
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	
	// Converter campos nullable
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

// SearchByFilter busca usuários por filtro (nome, CPF, username)
func (r *UserRepository) SearchByFilter(filter string) ([]models.User, error) {
	filterPattern := "%" + filter + "%"
	rows, err := r.db.Query(
		`SELECT id, username, role, full_name, cpf, rg, birth_date, gender, status, 
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
		
		err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.FullName, &u.CPF, &rg, &birthDate, 
			&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, 
			&u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		// Converter campos nullable
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
		
		users = append(users, u)
	}
	return users, nil
}
