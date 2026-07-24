use axum::{
    body::Body,
    body::to_bytes,
    http::{Method, Request, StatusCode},
    response::IntoResponse,
    Router,
};
use serde_json::json;
use tower::ServiceExt;

use donnapresentes::error::AppError;
use donnapresentes::routes;
use donnapresentes::AppState;

fn test_app(db_url: &str) -> Router {
    let config = donnapresentes::config::AppConfig {
        database_url: db_url.into(),
        jwt_secret: "test-secret-key-for-integration-tests".into(),
        encryption_key: None,
        xbz_cnpj: None,
        xbz_token: None,
        smtp_host: None,
        smtp_port: 587,
        smtp_user: None,
        smtp_password: None,
        smtp_from: None,
        tls_cert_file: None,
        tls_key_file: None,
        cors_allowed_origins: String::new(),
        port: "0".into(),
        log_format: "plain".into(),
    };

    let pool = donnapresentes::db::init_pool_lazy(db_url);

    let state = AppState {
        db: pool,
        config: std::sync::Arc::new(config),
    };

    routes::create_router(state)
}

fn auth_header(role: &str, user_id: i32) -> String {
    use jsonwebtoken::{encode, EncodingKey, Header};

    let claims = donnapresentes::middleware::auth::Claims {
        user_id,
        username: format!("{role}_user"),
        role: role.into(),
        permissions: if role == "admin" {
            vec!["admin".into()]
        } else {
            vec![]
        },
        exp: (chrono::Utc::now().timestamp() + 3600) as usize,
        iat: chrono::Utc::now().timestamp() as usize,
        sub: "access".into(),
    };

    let token = encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(b"test-secret-key-for-integration-tests"),
    )
    .unwrap();

    format!("Bearer {token}")
}

// --- Route + Middleware Tests ---

#[tokio::test]
async fn test_health_check() {
    let app = test_app("postgres:///nonexistent");

    let response = app
        .oneshot(
            Request::builder()
                .uri("/health")
                .method(Method::GET)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::OK);
}

#[tokio::test]
async fn test_unauthorized_without_token() {
    let app = test_app("postgres:///nonexistent");

    let response = app
        .oneshot(
            Request::builder()
                .uri("/users")
                .method(Method::GET)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn test_admin_endpoint_blocked_for_standard() {
    let app = test_app("postgres:///nonexistent");
    let token = auth_header("standard", 2);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/products-xbz/sync")
                .method(Method::POST)
                .header("Authorization", token)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::FORBIDDEN);
}

#[tokio::test]
async fn test_admin_endpoint_allows_admin_token() {
    let app = test_app("postgres:///nonexistent");
    let token = auth_header("admin", 1);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/products-xbz/sync")
                .method(Method::POST)
                .header("Authorization", token)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_ne!(response.status(), StatusCode::FORBIDDEN);
    assert_ne!(response.status(), StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn test_invalid_token_returns_unauthorized() {
    let app = test_app("postgres:///nonexistent");

    let response = app
        .oneshot(
            Request::builder()
                .uri("/users")
                .method(Method::GET)
                .header("Authorization", "Bearer invalid-token")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn test_auth_login_missing_fields() {
    let app = test_app("postgres:///nonexistent");

    let response = app
        .oneshot(
            Request::builder()
                .uri("/auth/login")
                .method(Method::POST)
                .header("Content-Type", "application/json")
                .body(Body::from(json!({"username": "test"}).to_string()))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
}

#[tokio::test]
async fn test_products_pending_alias_requires_auth_not_404() {
    let app = test_app("postgres:///nonexistent");

    let response = app
        .oneshot(
            Request::builder()
                .uri("/products/pending")
                .method(Method::GET)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn test_dashboard_seller_route_exists() {
    let app = test_app("postgres:///nonexistent");
    let token = auth_header("admin", 1);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/dashboard/seller/1")
                .method(Method::GET)
                .header("Authorization", token)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_ne!(response.status(), StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_dashboard_root_route_exists() {
    let app = test_app("postgres:///nonexistent");
    let token = auth_header("admin", 1);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/dashboard")
                .method(Method::GET)
                .header("Authorization", token)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_ne!(response.status(), StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_products_search_route_exists() {
    let app = test_app("postgres:///nonexistent");
    let token = auth_header("admin", 1);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/products/search?q=abc")
                .method(Method::GET)
                .header("Authorization", token)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_ne!(response.status(), StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_products_by_group_route_exists() {
    let app = test_app("postgres:///nonexistent");
    let token = auth_header("admin", 1);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/products/by-group/Brindes")
                .method(Method::GET)
                .header("Authorization", token)
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_ne!(response.status(), StatusCode::NOT_FOUND);
}

// --- Unit Tests ---

#[tokio::test]
async fn test_password_complexity_valid() {
    let result = donnapresentes::services::auth::validate_password_complexity("Abc123!@");
    assert!(result.is_ok());
}

#[tokio::test]
async fn test_password_complexity_short() {
    let result = donnapresentes::services::auth::validate_password_complexity("Ab1!");
    assert!(result.is_err());
}

#[tokio::test]
async fn test_password_complexity_no_upper() {
    let result = donnapresentes::services::auth::validate_password_complexity("abc123!@");
    assert!(result.is_err());
}

#[tokio::test]
async fn test_password_complexity_no_special() {
    let result = donnapresentes::services::auth::validate_password_complexity("Abc12345");
    assert!(result.is_err());
}

#[tokio::test]
async fn test_jwt_claims() {
    let claims = donnapresentes::middleware::auth::Claims::new(
        42,
        "testuser".into(),
        "admin".into(),
        vec!["read".into(), "write".into()],
    );

    assert_eq!(claims.user_id, 42);
    assert_eq!(claims.username, "testuser");
    assert_eq!(claims.role, "admin");
    assert_eq!(claims.permissions.len(), 2);
    assert_eq!(claims.sub, "access");
}

#[tokio::test]
async fn test_app_error_not_found() {
    let err = AppError::not_found("Produto");
    assert_eq!(err.message, "Produto não encontrado");
    assert_eq!(err.code, StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_app_error_unauthorized() {
    let err = AppError::unauthorized("Token inválido");
    assert_eq!(err.code, StatusCode::UNAUTHORIZED);
}

#[tokio::test]
async fn test_app_error_bad_request() {
    let err = AppError::bad_request("Campo obrigatório");
    assert_eq!(err.code, StatusCode::BAD_REQUEST);
}

#[tokio::test]
async fn test_app_error_conflict() {
    let err = AppError::conflict("Duplicado");
    assert_eq!(err.code, StatusCode::CONFLICT);
}

#[tokio::test]
async fn test_app_error_into_response_contains_details() {
    let response = AppError::bad_request("Campo obrigatório")
        .with_details("username")
        .into_response();

    assert_eq!(response.status(), StatusCode::BAD_REQUEST);

    let body = to_bytes(response.into_body(), usize::MAX).await.unwrap();
    let payload: serde_json::Value = serde_json::from_slice(&body).unwrap();

    assert_eq!(payload["error"], "Campo obrigatório");
    assert_eq!(payload["details"], "username");
    assert_eq!(payload["code"], 400);
}

#[tokio::test]
async fn test_rate_limiter_allows() {
    let limiter = donnapresentes::middleware::rate_limiter::RateLimiter::new();

    for _ in 0..10 {
        assert!(
            limiter
                .check("key", 10, std::time::Duration::from_secs(60))
                .await
        );
    }
}

#[tokio::test]
async fn test_rate_limiter_blocks_excess() {
    let limiter = donnapresentes::middleware::rate_limiter::RateLimiter::new();

    for _ in 0..10 {
        limiter
            .check("key2", 10, std::time::Duration::from_secs(60))
            .await;
    }

    assert!(
        !limiter
            .check("key2", 10, std::time::Duration::from_secs(60))
            .await
    );
}

#[tokio::test]
async fn test_sanitizer_trims() {
    let result = donnapresentes::services::sanitizer::sanitize_string("  hello world  ");
    assert_eq!(result, "hello world");
}

#[tokio::test]
async fn test_sanitizer_newlines() {
    let result = donnapresentes::services::sanitizer::sanitize_string("a\nb\tc");
    assert_eq!(result, "a\nb\tc");
}

#[tokio::test]
async fn test_sanitizer_empty() {
    assert_eq!(
        donnapresentes::services::sanitizer::sanitize_string(""),
        ""
    );
}

#[tokio::test]
async fn test_claims_encode_decode() {
    use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};

    let claims = donnapresentes::middleware::auth::Claims::new(
        99, "seller".into(), "standard".into(),
        vec!["sales".into()],
    );

    let token = encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(b"my-secret"),
    )
    .unwrap();

    let decoded = decode::<donnapresentes::middleware::auth::Claims>(
        &token,
        &DecodingKey::from_secret(b"my-secret"),
        &Validation::default(),
    )
    .unwrap();

    assert_eq!(decoded.claims.user_id, 99);
    assert_eq!(decoded.claims.role, "standard");
    assert_eq!(decoded.claims.permissions, vec!["sales"]);
}

#[tokio::test]
async fn test_claims_exp_is_in_future() {
    let now = chrono::Utc::now().timestamp() as usize;
    let claims = donnapresentes::middleware::auth::Claims::new(
        7,
        "futureuser".into(),
        "admin".into(),
        vec![],
    );

    assert!(claims.exp > now);
    assert!(claims.exp - claims.iat <= 900);
}

#[tokio::test]
async fn test_sale_input_serialization() {
    let input = donnapresentes::models::SaleInput {
        seller_id: 1,
        customer_id: 2,
        payment_method: Some("PIX".into()),
        installments: Some(3),
        payment_term_days: Some(30),
        first_installment_start: None,
        installment_dates: None,
        status: Some("Pendente".into()),
        is_event: Some(false),
        delivery_address: Some("Rua Teste".into()),
        delivery_date: None,
        departure_date: None,
        arrival_date: None,
        priority: Some("normal".into()),
        care_of: None,
        invoice_email: None,
        financial_email: None,
        purchase_order: None,
        external_notes: None,
        internal_notes: None,
        layout_urls: vec![],
        items: vec![],
        carrier_ids: vec![],
    };

    assert_eq!(input.seller_id, 1);
    assert_eq!(input.payment_method, Some("PIX".into()));
}

#[tokio::test]
async fn test_config_from_env() {
    std::env::set_var("DATABASE_URL", "postgres://test/db");
    std::env::set_var("JWT_SECRET", "secret123");
    let config = donnapresentes::config::AppConfig::from_env();
    std::env::remove_var("DATABASE_URL");
    std::env::remove_var("JWT_SECRET");

    assert!(config.is_ok());
    let c = config.unwrap();
    assert_eq!(c.database_url, "postgres://test/db");
    assert_eq!(c.jwt_secret, "secret123");
    assert_eq!(c.port, "8080");
}
