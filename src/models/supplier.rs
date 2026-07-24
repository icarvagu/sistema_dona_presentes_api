use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Supplier {
    pub id: i32,
    pub name: String,
    pub cnpj: Option<String>,
    pub state_registration: Option<String>,
    pub contact_person: Option<String>,
    pub email: Option<String>,
    pub landline_phone: Option<String>,
    pub mobile_phone: Option<String>,
    pub responsible_email: Option<String>,
    pub commercial_address: Option<String>,
    pub postal_code: Option<String>,
    pub website: Option<String>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
pub struct SupplierInput {
    pub name: String,
    pub cnpj: Option<String>,
    pub state_registration: Option<String>,
    pub contact_person: Option<String>,
    pub email: Option<String>,
    pub landline_phone: Option<String>,
    pub mobile_phone: Option<String>,
    pub responsible_email: Option<String>,
    pub commercial_address: Option<String>,
    pub postal_code: Option<String>,
    pub website: Option<String>,
}
