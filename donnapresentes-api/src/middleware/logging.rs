use axum::{
    body::Body,
    http::Request,
    middleware::Next,
    response::Response,
};
use tracing::error;


pub async fn recovery_middleware(req: Request<Body>, next: Next) -> Response {
    let response = next.run(req).await;

    let status = response.status();
    if status.is_server_error() {
        error!(
            status = status.as_u16(),
            "Request completed with error"
        );
    }

    response
}
