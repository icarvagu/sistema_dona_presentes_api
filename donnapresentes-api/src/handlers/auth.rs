use axum::{extract::State, Extension, Json};

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::*;
use crate::response::{ok_response, no_content};
use crate::services::auth::AuthService;
use crate::AppState;

pub async fn login(
    State(state): State<AppState>,
    Json(input): Json<LoginInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let response = AuthService::login(&state.db, &state.config.jwt_secret, &input).await?;
    Ok(ok_response(response))
}

pub async fn refresh(
    State(state): State<AppState>,
    Json(input): Json<RefreshRequest>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let response =
        AuthService::refresh_access_token(&state.db, &state.config.jwt_secret, &input.refresh_token)
            .await?;
    Ok(ok_response(response))
}

pub async fn logout(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Json(_input): Json<LogoutRequest>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    AuthService::logout(&state.db, user.user_id).await?;
    Ok(no_content())
}

pub async fn get_current_user(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let u = AuthService::get_user_by_id(&state.db, user.user_id).await?;
    Ok(ok_response(u))
}

pub async fn change_password(
    State(state): State<AppState>,
    Extension(user): Extension<UserContext>,
    Json(input): Json<ChangePasswordRequest>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    AuthService::change_password(&state.db, user.user_id, &input.new_password).await?;
    Ok(ok_response(serde_json::json!({"message": "Senha alterada com sucesso"})))
}

pub async fn forgot_password(
    State(state): State<AppState>,
    Json(input): Json<ForgotPasswordRequest>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    AuthService::initiate_password_reset(&state.db, &input.username).await?;
    Ok(ok_response(serde_json::json!({"message": "Se o usuário existir, um email será enviado"})))
}

pub async fn reset_password(
    State(state): State<AppState>,
    Json(input): Json<ResetPasswordRequest>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    AuthService::reset_password(&state.db, &input.token, &input.new_password).await?;
    Ok(ok_response(serde_json::json!({"message": "Senha redefinida com sucesso"})))
}
