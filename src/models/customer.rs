use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Customer {
    pub id: i32,
    pub customer_type: String,
    pub status: String,
    pub name: String,
    pub trade_name: Option<String>,
    pub company_name: Option<String>,
    pub state_registration: Option<String>,
    pub city_registration: Option<String>,
    pub responsible: Option<String>,
    pub contact_financial_name: Option<String>,
    pub contact_financial_email: Option<String>,
    pub contact_financial_phone: Option<String>,
    pub contact_nf_name: Option<String>,
    pub contact_nf_email: Option<String>,
    pub contact_nf_phone: Option<String>,
    pub contact_commercial_name: Option<String>,
    pub contact_commercial_email: Option<String>,
    pub contact_commercial_phone: Option<String>,
    pub cnpj: Option<String>,
    pub cpf: Option<String>,
    pub email: Option<String>,
    pub business_phone: Option<String>,
    pub mobile_phone: Option<String>,
    pub website: Option<String>,
    pub notes: Option<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[sqlx(skip)]
    pub addresses: Vec<Address>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    #[sqlx(skip)]
    pub additional_contacts: Vec<AdditionalContact>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Address {
    pub id: i32,
    pub customer_id: i32,
    pub address_type: String,
    #[sqlx(rename = "address")]
    pub address_line: String,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct AdditionalContact {
    pub id: i32,
    pub customer_id: i32,
    pub name: String,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
pub struct CustomerInput {
    pub customer_type: String,
    pub status: Option<String>,
    pub name: String,
    pub trade_name: Option<String>,
    pub company_name: Option<String>,
    pub state_registration: Option<String>,
    pub city_registration: Option<String>,
    pub responsible: Option<String>,
    pub contact_financial_name: Option<String>,
    pub contact_financial_email: Option<String>,
    pub contact_financial_phone: Option<String>,
    pub contact_nf_name: Option<String>,
    pub contact_nf_email: Option<String>,
    pub contact_nf_phone: Option<String>,
    pub contact_commercial_name: Option<String>,
    pub contact_commercial_email: Option<String>,
    pub contact_commercial_phone: Option<String>,
    pub cnpj: Option<String>,
    pub cpf: Option<String>,
    pub email: Option<String>,
    pub business_phone: Option<String>,
    pub mobile_phone: Option<String>,
    pub website: Option<String>,
    pub notes: Option<String>,
    #[serde(default)]
    pub addresses: Vec<AddressInput>,
    #[serde(default)]
    pub additional_contacts: Vec<AdditionalContactInput>,
}

#[derive(Debug, Deserialize)]
pub struct AddressInput {
    pub address_type: String,
    pub address: String,
}

#[derive(Debug, Deserialize)]
pub struct AdditionalContactInput {
    pub name: String,
    pub email: Option<String>,
    pub phone: Option<String>,
}
