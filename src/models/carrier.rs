use chrono::NaiveDateTime;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Carrier {
    pub id: i32,
    pub name: String,
    pub cnpj: Option<String>,
    pub carrier_type: String,
    pub email: Option<String>,
    pub landline_phone: Option<String>,
    pub mobile_phone: Option<String>,
    pub full_address: Option<String>,
    pub contact_name: Option<String>,
    pub contact_phone: Option<String>,
    pub website: Option<String>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
}

#[derive(Debug, Deserialize)]
pub struct CarrierInput {
    pub name: String,
    pub cnpj: Option<String>,
    pub carrier_type: String,
    pub email: Option<String>,
    pub landline_phone: Option<String>,
    pub mobile_phone: Option<String>,
    pub full_address: Option<String>,
    pub contact_name: Option<String>,
    pub contact_phone: Option<String>,
    pub website: Option<String>,
}
