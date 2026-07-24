use axum::{
    extract::{Path, State},
    Extension, Json,
};
use bcrypt;

use crate::error::AppError;
use crate::middleware::auth::UserContext;
use crate::models::UserInput;
use crate::repositories;
use crate::response::{created_response, no_content, ok_response};
use crate::AppState;

pub async fn list(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let users = repositories::user::get_all(&state.db).await?;
    Ok(ok_response(users))
}

pub async fn get_by_id(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let user = repositories::user::get_by_id(&state.db, id).await?;
    Ok(ok_response(user))
}

pub async fn create(
    State(state): State<AppState>,
    _admin: Extension<UserContext>,
    Json(input): Json<UserInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    if input.username.trim().is_empty() {
        return Err(AppError::validation("Nome de usuário é obrigatório"));
    }
    if let Some(ref pwd) = input.password {
        crate::services::auth::validate_password_complexity(pwd)?;
        let hash = bcrypt::hash(pwd, bcrypt::DEFAULT_COST)
            .map_err(|e| AppError::internal(e.to_string()))?;
        let user = repositories::user::create(&state.db, &input, &hash).await?;
        Ok(created_response(user))
    } else {
        Err(AppError::validation("Senha é obrigatória"))
    }
}

pub async fn update(
    State(state): State<AppState>,
    _user: Extension<UserContext>,
    Path(id): Path<i32>,
    Json(input): Json<UserInput>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    let user = repositories::user::update(&state.db, id, &input).await?;
    Ok(ok_response(user))
}

pub async fn delete(
    State(state): State<AppState>,
    _admin: Extension<UserContext>,
    Path(id): Path<i32>,
) -> Result<impl axum::response::IntoResponse, AppError> {
    repositories::user::delete(&state.db, id).await?;
    Ok(no_content())
}
