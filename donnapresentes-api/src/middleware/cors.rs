use axum::{
    body::Body,
    http::{Method, Request},
    response::Response,
};
use tower_http::cors::CorsLayer;

pub fn cors_layer(allowed_origins: &str) -> CorsLayer {
    let mut origins = vec![
        "http://localhost:8081".parse().unwrap(),
        "http://localhost:19006".parse().unwrap(),
        "http://localhost:5173".parse().unwrap(),
        "http://localhost:5174".parse().unwrap(),
        "http://localhost:5175".parse().unwrap(),
        "http://localhost:5176".parse().unwrap(),
    ];

    for origin in allowed_origins.split(',').map(|s| s.trim()).filter(|s| !s.is_empty()) {
        if let Ok(o) = origin.parse() {
            origins.push(o);
        }
    }

    CorsLayer::new()
        .allow_origin(origins)
        .allow_credentials(true)
        .allow_methods([
            Method::GET,
            Method::POST,
            Method::PUT,
            Method::DELETE,
            Method::PATCH,
            Method::OPTIONS,
        ])
        .allow_headers([
            "Content-Type".parse().unwrap(),
            "Authorization".parse().unwrap(),
            "X-API-Key".parse().unwrap(),
            "X-Requested-With".parse().unwrap(),
            "Accept".parse().unwrap(),
            "Origin".parse().unwrap(),
            "X-CSRF-Token".parse().unwrap(),
        ])
        .expose_headers([
            "Content-Length".parse().unwrap(),
            "Content-Type".parse().unwrap(),
            "Location".parse().unwrap(),
        ])
        .max_age(std::time::Duration::from_secs(3600))
}

pub async fn cors_fallback(req: Request<Body>) -> Response {
    if req.method() == Method::OPTIONS {
        Response::builder()
            .status(204)
            .body(Body::empty())
            .unwrap()
    } else {
        Response::builder()
            .status(404)
            .body(Body::from("Not Found"))
            .unwrap()
    }
}

