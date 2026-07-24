use std::net::SocketAddr;
use tower_http::normalize_path::NormalizePathLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use donnapresentes::config::AppConfig;
use donnapresentes::db;
use donnapresentes::middleware::cors;
use donnapresentes::routes;
use donnapresentes::AppState;

#[tokio::main]
async fn main() {
    dotenvy::dotenv().ok();

    let log_format = std::env::var("LOG_FORMAT").unwrap_or_else(|_| "plain".into());
    if log_format == "json" {
        tracing_subscriber::registry()
            .with(tracing_subscriber::fmt::layer().json())
            .with(tracing_subscriber::EnvFilter::new(
                std::env::var("RUST_LOG").unwrap_or_else(|_| "info".into()),
            ))
            .init();
    } else {
        tracing_subscriber::registry()
            .with(tracing_subscriber::fmt::layer())
            .with(tracing_subscriber::EnvFilter::new(
                std::env::var("RUST_LOG").unwrap_or_else(|_| "info".into()),
            ))
            .init();
    }

    let config = AppConfig::from_env().expect("Failed to load configuration");
    let port = config.port.clone();

    let pool = db::init_pool(&config.database_url)
        .await
        .expect("Failed to connect to database");

    db::run_migrations(&pool)
        .await
        .expect("Failed to run migrations");

    tracing::info!("Database connected and migrated successfully");

    let app_state = AppState {
        db: pool,
        config: std::sync::Arc::new(config),
    };

    let cors_layer = cors::cors_layer(&app_state.config.cors_allowed_origins);

    let app = routes::create_router(app_state)
        .layer(cors_layer)
        .layer(NormalizePathLayer::trim_trailing_slash());

    let addr: SocketAddr = format!("0.0.0.0:{port}").parse().unwrap();
    tracing::info!("Server starting on {addr}");

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
