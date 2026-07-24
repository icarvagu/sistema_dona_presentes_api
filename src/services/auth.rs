use chrono::Utc;
use jsonwebtoken::{encode, EncodingKey, Header};
use rand::Rng;
use sha2::{Digest, Sha256};
use sqlx::PgPool;

use crate::error::AppError;
use crate::middleware::auth::Claims;
use crate::models::{LoginInput, LoginResponse, User};

pub struct AuthService;

impl AuthService {
    pub async fn login(
        pool: &PgPool,
        jwt_secret: &str,
        input: &LoginInput,
    ) -> Result<LoginResponse, AppError> {
        let username = sanitize_username(&input.username);
        if username.is_empty() {
            return Err(AppError::validation("Usuário ou senha inválidos"));
        }

        let user = sqlx::query_as::<_, User>(
            "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                    rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                    notes, failed_login_attempts, locked_until, must_change_password,
                    created_at, updated_at
             FROM users WHERE username = $1",
        )
        .bind(&username)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()).with_log(format!("DB error: {e}")))?;

        let user = user.ok_or_else(|| AppError::validation("Usuário ou senha inválidos"))?;

        if let Some(locked_until) = user.locked_until {
            if Utc::now() < locked_until {
                return Err(AppError::validation(
                    "Conta temporariamente bloqueada por muitas tentativas. Tente novamente mais tarde",
                ));
            }
        }

        if !bcrypt::verify(&input.password, &user.password_hash).unwrap_or(false) {
            let new_attempts = user.failed_login_attempts + 1;
            let lock_until = if new_attempts >= 5 {
                Some(Utc::now() + chrono::Duration::minutes(15))
            } else {
                None
            };

            let _ = sqlx::query(
                "UPDATE users SET failed_login_attempts = $1, locked_until = $2 WHERE id = $3",
            )
            .bind(new_attempts)
            .bind(lock_until)
            .bind(user.id)
            .execute(pool)
            .await;

            return Err(AppError::validation("Usuário ou senha inválidos"));
        }

        let _ = sqlx::query("UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE id = $1")
            .bind(user.id)
            .execute(pool)
            .await;

        let access_token = Self::generate_access_token(&user, jwt_secret)?;
        let refresh_token = Self::generate_and_store_refresh_token(pool, user.id).await?;

        Ok(LoginResponse {
            token: access_token,
            refresh_token,
            user,
        })
    }

    pub fn generate_access_token(user: &User, jwt_secret: &str) -> Result<String, AppError> {
        let permissions = if user.permissions.is_empty() {
            vec![]
        } else {
            user.permissions.clone()
        };

        let claims = Claims::new(user.id, user.username.clone(), user.role.clone(), permissions);

        encode(
            &Header::default(),
            &claims,
            &EncodingKey::from_secret(jwt_secret.as_bytes()),
        )
        .map_err(|e| AppError::internal(e.to_string()))
    }

    pub async fn generate_and_store_refresh_token(
        pool: &PgPool,
        user_id: i32,
    ) -> Result<String, AppError> {
        let random_bytes: [u8; 32] = rand::thread_rng().gen();
        let token = hex::encode(random_bytes);

        let token_hash = hex::encode(Sha256::digest(token.as_bytes()));
        let expires_at = Utc::now() + chrono::Duration::days(7);

        sqlx::query(
            "INSERT INTO refresh_tokens (user_id, token_hash, revoked, expires_at)
             VALUES ($1, $2, false, $3)",
        )
        .bind(user_id)
        .bind(&token_hash)
        .bind(expires_at)
        .execute(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;

        Ok(token)
    }

    pub async fn refresh_access_token(
        pool: &PgPool,
        jwt_secret: &str,
        refresh_token: &str,
    ) -> Result<LoginResponse, AppError> {
        let token_hash = hex::encode(Sha256::digest(refresh_token.as_bytes()));

        let rt = sqlx::query_as::<_, crate::models::RefreshToken>(
            "SELECT id, user_id, token_hash, revoked, expires_at, created_at
             FROM refresh_tokens WHERE token_hash = $1",
        )
        .bind(&token_hash)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::unauthorized("Refresh token inválido ou expirado"))?;

        if rt.revoked || Utc::now() > rt.expires_at {
            return Err(AppError::unauthorized("Refresh token inválido ou expirado"));
        }

        sqlx::query("UPDATE refresh_tokens SET revoked = true WHERE id = $1")
            .bind(rt.id)
            .execute(pool)
            .await
            .ok();

        let user = sqlx::query_as::<_, User>(
            "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                    rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                    notes, failed_login_attempts, locked_until, must_change_password,
                    created_at, updated_at
             FROM users WHERE id = $1",
        )
        .bind(rt.user_id)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::unauthorized("Usuário não encontrado"))?;

        if let Some(locked) = user.locked_until {
            if Utc::now() < locked {
                return Err(AppError::validation("Conta temporariamente bloqueada"));
            }
        }

        let access_token = Self::generate_access_token(&user, jwt_secret)?;
        let new_refresh_token = Self::generate_and_store_refresh_token(pool, user.id).await?;

        Ok(LoginResponse {
            token: access_token,
            refresh_token: new_refresh_token,
            user,
        })
    }

    pub async fn logout(pool: &PgPool, user_id: i32) -> Result<(), AppError> {
        sqlx::query("UPDATE refresh_tokens SET revoked = true WHERE user_id = $1")
            .bind(user_id)
            .execute(pool)
            .await
            .map_err(|e| AppError::internal(e.to_string()))?;
        Ok(())
    }

    pub async fn change_password(
        pool: &PgPool,
        user_id: i32,
        new_password: &str,
    ) -> Result<(), AppError> {
        let password = sanitize_password(new_password);
        validate_password_complexity(&password)?;

        let hash = bcrypt::hash(&password, bcrypt::DEFAULT_COST)
            .map_err(|e| AppError::internal(e.to_string()))?;

        sqlx::query("UPDATE users SET password_hash = $1, must_change_password = false WHERE id = $2")
            .bind(&hash)
            .bind(user_id)
            .execute(pool)
            .await
            .map_err(|e| AppError::internal(e.to_string()))?;

        sqlx::query("UPDATE refresh_tokens SET revoked = true WHERE user_id = $1")
            .bind(user_id)
            .execute(pool)
            .await
            .ok();

        Ok(())
    }

    pub async fn initiate_password_reset(
        pool: &PgPool,
        username: &str,
    ) -> Result<(), AppError> {
        let username = sanitize_username(username);
        let user = sqlx::query_as::<_, User>(
            "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                    rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                    notes, failed_login_attempts, locked_until, must_change_password,
                    created_at, updated_at
             FROM users WHERE username = $1",
        )
        .bind(&username)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;

        let Some(user) = user else {
            return Ok(());
        };

        let random_bytes: [u8; 32] = rand::thread_rng().gen();
        let token = hex::encode(random_bytes);
        let token_hash = hex::encode(Sha256::digest(token.as_bytes()));
        let expires_at = Utc::now() + chrono::Duration::hours(1);

        sqlx::query(
            "INSERT INTO password_reset_tokens (user_id, token_hash, used, expires_at)
             VALUES ($1, $2, false, $3)",
        )
        .bind(user.id)
        .bind(&token_hash)
        .bind(expires_at)
        .execute(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?;

        Ok(())
    }

    pub async fn reset_password(
        pool: &PgPool,
        token: &str,
        new_password: &str,
    ) -> Result<(), AppError> {
        let password = sanitize_password(new_password);
        validate_password_complexity(&password)?;

        let token_hash = hex::encode(Sha256::digest(token.as_bytes()));

        let prt = sqlx::query_as::<_, crate::models::PasswordResetToken>(
            "SELECT id, user_id, token_hash, used, expires_at, created_at
             FROM password_reset_tokens WHERE token_hash = $1",
        )
        .bind(&token_hash)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::validation("Token de reset inválido ou expirado"))?;

        if prt.used || Utc::now() > prt.expires_at {
            return Err(AppError::validation("Token de reset inválido ou expirado"));
        }

        let hash = bcrypt::hash(&password, bcrypt::DEFAULT_COST)
            .map_err(|e| AppError::internal(e.to_string()))?;

        sqlx::query("UPDATE users SET password_hash = $1 WHERE id = $2")
            .bind(&hash)
            .bind(prt.user_id)
            .execute(pool)
            .await
            .map_err(|e| AppError::internal(e.to_string()))?;

        sqlx::query("UPDATE password_reset_tokens SET used = true WHERE id = $1")
            .bind(prt.id)
            .execute(pool)
            .await
            .ok();

        sqlx::query("UPDATE refresh_tokens SET revoked = true WHERE user_id = $1")
            .bind(prt.user_id)
            .execute(pool)
            .await
            .ok();

        Ok(())
    }

    pub async fn validate_token(
        jwt_secret: &str,
        token: &str,
    ) -> Result<Claims, AppError> {
        use jsonwebtoken::{decode, DecodingKey, Validation};

        decode::<Claims>(
            token,
            &DecodingKey::from_secret(jwt_secret.as_bytes()),
            &Validation::default(),
        )
        .map(|data| data.claims)
        .map_err(|_| AppError::unauthorized("Token inválido ou expirado"))
    }

    pub async fn get_user_by_id(pool: &PgPool, id: i32) -> Result<User, AppError> {
        sqlx::query_as::<_, User>(
            "SELECT id, username, password_hash, role, permissions, full_name, cpf, cpf_hash,
                    rg, birth_date, gender, status, contact_email, full_address, contact_phone,
                    notes, failed_login_attempts, locked_until, must_change_password,
                    created_at, updated_at
             FROM users WHERE id = $1",
        )
        .bind(id)
        .fetch_optional(pool)
        .await
        .map_err(|e| AppError::internal(e.to_string()))?
        .ok_or_else(|| AppError::not_found("Usuário"))
    }
}

pub fn validate_password_complexity(password: &str) -> Result<(), AppError> {
    if password.len() < 8 {
        return Err(AppError::validation("Senha deve ter pelo menos 8 caracteres"));
    }
    if !password.chars().any(|c| c.is_uppercase()) {
        return Err(AppError::validation("Senha deve conter pelo menos uma letra maiúscula"));
    }
    if !password.chars().any(|c| c.is_lowercase()) {
        return Err(AppError::validation("Senha deve conter pelo menos uma letra minúscula"));
    }
    if !password.chars().any(|c| c.is_ascii_digit()) {
        return Err(AppError::validation("Senha deve conter pelo menos um dígito"));
    }
    if !password.chars().any(|c| !c.is_alphanumeric()) {
        return Err(AppError::validation("Senha deve conter pelo menos um caractere especial"));
    }
    Ok(())
}

fn sanitize_username(s: &str) -> String {
    s.trim().to_lowercase()
}

fn sanitize_password(s: &str) -> String {
    s.trim().to_string()
}
