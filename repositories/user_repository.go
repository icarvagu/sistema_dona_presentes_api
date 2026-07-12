package repositories

import (
	"database/sql"
	"time"

	apperrors "donapresentes/errors"
	"donapresentes/models"

	"github.com/lib/pq"
)

// PasswordResetToken represents a row from the password_reset_tokens table.
type PasswordResetToken struct {
	ID        int
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// UserRepository handles all database operations for the users and password_reset_tokens tables.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository with the given database connection.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername retrieves a user by their username from the database.
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate, lockedUntil sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, failed_login_attempts, locked_until, must_change_password, created_at, updated_at
		 FROM users WHERE username=$1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.FailedLoginAttempts, &lockedUntil, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
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
	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}
	
	return &u, nil
}

// GetByID retrieves a user by their primary key from the database.
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate, lockedUntil sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, failed_login_attempts, locked_until, must_change_password, created_at, updated_at
		 FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.FailedLoginAttempts, &lockedUntil, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
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
	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}
	
	return &u, nil
}

// Create inserts a new user record and returns the created user with its ID and timestamps.
func (r *UserRepository) Create(u *models.User) (*models.User, error) {
	if u.Permissions == nil {
		u.Permissions = []string{}
	}
	err := r.db.QueryRow(
		`INSERT INTO users (username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, must_change_password)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING id, created_at, updated_at`,
		u.Username, u.PasswordHash, u.Role, pq.Array(u.Permissions), u.FullName, u.CPF, u.RG, u.BirthDate, u.Gender, u.Status,
		u.ContactEmail, u.FullAddress, u.ContactPhone, u.Notes, u.MustChangePassword,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return u, nil
}

// GetAll returns all users from the database ordered by ID.
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

// Update modifies an existing user record identified by id with the provided data.
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

// UpdatePassword updates the password hash for a user and clears the must_change_password flag.
func (r *UserRepository) UpdatePassword(id int, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash=$1, must_change_password=FALSE, updated_at=NOW() WHERE id=$2`, passwordHash, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Delete removes a user record by its primary key.
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

// GetByCPF retrieves a user by their CPF number from the database.
func (r *UserRepository) GetByCPF(cpf string) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate, lockedUntil sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, permissions, full_name, cpf, rg, birth_date, gender, status,
		 contact_email, full_address, contact_phone, notes, failed_login_attempts, locked_until, must_change_password, created_at, updated_at
		 FROM users WHERE cpf=$1`,
		cpf,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, pq.Array(&u.Permissions), &u.FullName, &u.CPF, &rg, &birthDate,
		&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &u.FailedLoginAttempts, &lockedUntil, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
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
	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}
	
	return &u, nil
}

// SearchByFilter searches users by full_name, username, or CPF using a LIKE pattern.
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

// UpdateFailedAttempts updates the failed login attempts counter and lockout time for a user.
func (r *UserRepository) UpdateFailedAttempts(id int, attempts int, lockedUntil *time.Time) error {
	_, err := r.db.Exec(
		`UPDATE users SET failed_login_attempts=$1, locked_until=$2, updated_at=NOW() WHERE id=$3`,
		attempts, lockedUntil, id,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// ResetFailedAttempts resets the failed login attempts counter and clears the lockout for a user.
func (r *UserRepository) ResetFailedAttempts(id int) error {
	_, err := r.db.Exec(
		`UPDATE users SET failed_login_attempts=0, locked_until=NULL, updated_at=NOW() WHERE id=$1`,
		id,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// UpdateCPFHash stores the CPF hash for a user record.
func (r *UserRepository) UpdateCPFHash(id int, cpfHash string) error {
	_, err := r.db.Exec(`UPDATE users SET cpf_hash=$1 WHERE id=$2`, cpfHash, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// GetByCPFHash retrieves a user by their CPF hash from the database.
func (r *UserRepository) GetByCPFHash(cpfHash string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(
		`SELECT id, username, role, full_name FROM users WHERE cpf_hash=$1`,
		cpfHash,
	).Scan(&u.ID, &u.Username, &u.Role, &u.FullName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &u, nil
}

// CreatePasswordResetToken inserts a new password reset token record into the database.
func (r *UserRepository) CreatePasswordResetToken(userID int, tokenHash string, expiresAt time.Time) (*PasswordResetToken, error) {
	var prt PasswordResetToken
	err := r.db.QueryRow(
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at`,
		userID, tokenHash, expiresAt,
	).Scan(&prt.ID, &prt.CreatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	prt.UserID = userID
	prt.TokenHash = tokenHash
	prt.ExpiresAt = expiresAt
	return &prt, nil
}

// FindPasswordResetToken retrieves a password reset token by its hash from the database.
func (r *UserRepository) FindPasswordResetToken(tokenHash string) (*PasswordResetToken, error) {
	var prt PasswordResetToken
	err := r.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, used, created_at FROM password_reset_tokens WHERE token_hash=$1`,
		tokenHash,
	).Scan(&prt.ID, &prt.UserID, &prt.TokenHash, &prt.ExpiresAt, &prt.Used, &prt.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Token não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &prt, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used in the database.
func (r *UserRepository) MarkPasswordResetTokenUsed(id int) error {
	_, err := r.db.Exec(`UPDATE password_reset_tokens SET used=TRUE WHERE id=$1`, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}
