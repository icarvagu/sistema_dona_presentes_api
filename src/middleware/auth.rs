use axum::{
    body::Body,
    extract::State,
    http::Request,
    middleware::Next,
    response::{IntoResponse, Response},
};
use chrono::Utc;
use jsonwebtoken::{decode, DecodingKey, Validation};
use serde::{Deserialize, Serialize};

use crate::error::AppError;
use crate::AppState;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub user_id: i32,
    pub username: String,
    pub role: String,
    pub permissions: Vec<String>,
    pub exp: usize,
    pub iat: usize,
    pub sub: String,
}

impl Claims {
    pub fn new(user_id: i32, username: String, role: String, permissions: Vec<String>) -> Self {
        let now = Utc::now().timestamp() as usize;
        Self {
            user_id,
            username,
            role,
            permissions,
            exp: now + 900,
            iat: now,
            sub: "access".into(),
        }
    }
}

#[derive(Clone, Debug)]
pub struct UserContext {
    pub user_id: i32,
    pub username: String,
    pub role: String,
    pub permissions: Vec<String>,
}

pub async fn auth_middleware(
    State(state): State<AppState>,
    mut req: Request<Body>,
    next: Next,
) -> Response {
    let token = extract_token(&req);

    let Some(token) = token else {
        return AppError::unauthorized("Token de autenticação não fornecido").into_response();
    };

    let decoding_key = DecodingKey::from_secret(state.config.jwt_secret.as_bytes());
    let validation = Validation::default();

    let token_data = match decode::<Claims>(&token, &decoding_key, &validation) {
        Ok(data) => data,
        Err(_) => {
            return AppError::unauthorized("Token inválido ou expirado").into_response();
        }
    };

    let claims = token_data.claims;

    let user = UserContext {
        user_id: claims.user_id,
        username: claims.username,
        role: claims.role,
        permissions: claims.permissions,
    };

    req.extensions_mut().insert(user);
    next.run(req).await
}

pub fn require_user<B>(req: &Request<B>) -> Result<&UserContext, AppError> {
    req.extensions()
        .get::<UserContext>()
        .ok_or_else(|| AppError::unauthorized("Usuário não autenticado"))
}

pub fn require_user_id<B>(req: &Request<B>) -> Result<i32, AppError> {
    require_user(req).map(|u| u.user_id)
}

pub async fn admin_only_middleware(
    req: Request<Body>,
    next: Next,
) -> Response {
    let Some(user) = req.extensions().get::<UserContext>() else {
        return AppError::forbidden("Acesso negado").into_response();
    };

    if user.role != "admin" {
        return AppError::forbidden(
            "Acesso negado. Apenas administradores podem acessar este recurso",
        ).into_response();
    }

    next.run(req).await
}

pub fn has_permission<B>(req: &Request<B>, permission: &str) -> bool {
    req.extensions()
        .get::<UserContext>()
        .map(|u| u.role == "admin" || u.permissions.iter().any(|p| p == permission))
        .unwrap_or(false)
}

fn extract_token<B>(req: &Request<B>) -> Option<String> {
    if let Some(auth) = req.headers().get("Authorization") {
        if let Ok(auth) = auth.to_str() {
            if let Some(token) = auth.strip_prefix("Bearer ") {
                return Some(token.to_string());
            }
        }
    }
    None
}
