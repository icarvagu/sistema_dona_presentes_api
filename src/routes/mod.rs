use axum::{middleware, Router};

use crate::AppState;

mod auth;
mod users;
mod customers;
mod suppliers;
mod carriers;
mod products;
mod quotes;
mod sales;
mod purchases;
mod production;
mod art_final;
mod dashboard;
mod sales_workflow;
mod admin;
mod item_layouts;

use auth::{public_router as auth_public, protected_router as auth_protected};

pub fn create_router(state: AppState) -> Router {
    let auth_layer = middleware::from_fn_with_state(
        state.clone(),
        crate::middleware::auth::auth_middleware,
    );

    let admin_layer = middleware::from_fn(crate::middleware::auth::admin_only_middleware);

    let admin = Router::new()
        .merge(admin::router())
        .route_layer(admin_layer);

    let protected = Router::new()
        .merge(auth_protected())
        .merge(users::router())
        .merge(customers::router())
        .merge(suppliers::router())
        .merge(carriers::router())
        .merge(products::router())
        .merge(quotes::router())
        .merge(sales::router())
        .merge(purchases::router())
        .merge(production::router())
        .merge(art_final::router())
        .merge(dashboard::router())
        .merge(sales_workflow::router())
        .merge(item_layouts::router())
        .merge(admin)
        .route_layer(auth_layer);

    Router::new()
        .merge(auth_public())
        .merge(protected)
        .with_state(state)
}
