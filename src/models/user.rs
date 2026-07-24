use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct User {
    pub id: i32,
    pub username: String,
    #[serde(skip_serializing)]
    pub password_hash: String,
    pub role: String,
    #[sqlx(default)]
    #[serde(default)]
    pub permissions: Vec<String>,
    pub full_name: Option<String>,
    pub cpf: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpf_hash: Option<String>,
    pub rg: Option<String>,
    pub birth_date: Option<chrono::NaiveDate>,
    pub gender: Option<String>,
    pub status: String,
    pub contact_email: Option<String>,
    pub full_address: Option<String>,
    pub contact_phone: Option<String>,
    pub notes: Option<String>,
    #[serde(skip_serializing)]
    pub failed_login_attempts: i32,
    pub locked_until: Option<chrono::DateTime<chrono::Utc>>,
    pub must_change_password: bool,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
pub struct UserInput {
    pub username: String,
    pub password: Option<String>,
    pub role: Option<String>,
    #[serde(default)]
    pub permissions: Vec<String>,
    pub full_name: String,
    pub cpf: String,
    pub rg: Option<String>,
    pub birth_date: Option<chrono::NaiveDate>,
    pub gender: Option<String>,
    pub status: Option<String>,
    pub contact_email: Option<String>,
    pub full_address: Option<String>,
    pub contact_phone: Option<String>,
    pub notes: Option<String>,
    #[serde(default)]
    pub must_change_password: bool,
}

#[derive(Debug, Deserialize)]
pub struct LoginInput {
    pub username: String,
    pub password: String,
}

#[derive(Debug, Serialize)]
pub struct LoginResponse {
    pub token: String,
    pub refresh_token: String,
    pub user: User,
}

#[derive(Debug, Deserialize)]
pub struct RefreshRequest {
    pub refresh_token: String,
}

#[derive(Debug, Deserialize)]
pub struct ForgotPasswordRequest {
    pub username: String,
}

#[derive(Debug, Deserialize)]
pub struct ResetPasswordRequest {
    pub token: String,
    pub new_password: String,
}

#[derive(Debug, Deserialize)]
pub struct LogoutRequest {
    pub refresh_token: String,
}

#[derive(Debug, Deserialize)]
pub struct ChangePasswordRequest {
    pub new_password: String,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct RefreshToken {
    pub id: i32,
    pub user_id: i32,
    pub token_hash: String,
    pub revoked: bool,
    pub expires_at: chrono::DateTime<chrono::Utc>,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct PasswordResetToken {
    pub id: i32,
    pub user_id: i32,
    pub token_hash: String,
    pub used: bool,
    pub expires_at: chrono::DateTime<chrono::Utc>,
    pub created_at: chrono::DateTime<chrono::Utc>,
}
