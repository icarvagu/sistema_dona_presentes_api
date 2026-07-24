use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct ProductItem {
    pub id: i32,
    pub product_parent_id: i32,
    pub product_id: i32,
    pub quantity: i32,
    #[sqlx(skip)]
    pub product: Option<super::Product>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
pub struct ProductItemInput {
    pub product_id: i32,
    pub quantity: i32,
}
