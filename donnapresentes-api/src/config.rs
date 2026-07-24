use std::env;

pub struct AppConfig {
    pub database_url: String,
    pub jwt_secret: String,
    pub encryption_key: Option<String>,
    pub xbz_cnpj: Option<String>,
    pub xbz_token: Option<String>,
    pub smtp_host: Option<String>,
    pub smtp_port: u16,
    pub smtp_user: Option<String>,
    pub smtp_password: Option<String>,
    pub smtp_from: Option<String>,
    pub tls_cert_file: Option<String>,
    pub tls_key_file: Option<String>,
    pub cors_allowed_origins: String,
    pub port: String,
    pub log_format: String,
}

impl AppConfig {
    pub fn from_env() -> Result<Self, String> {
        Ok(Self {
            database_url: require_env("DATABASE_URL")?,
            jwt_secret: require_env("JWT_SECRET")?,
            encryption_key: env::var("ENCRYPTION_KEY").ok().filter(|s| !s.is_empty()),
            xbz_cnpj: env::var("XBZ_CNPJ").ok().filter(|s| !s.is_empty()),
            xbz_token: env::var("XBZ_TOKEN").ok().filter(|s| !s.is_empty()),
            smtp_host: env::var("SMTP_HOST").ok().filter(|s| !s.is_empty()),
            smtp_port: env::var("SMTP_PORT").ok().and_then(|p| p.parse().ok()).unwrap_or(587),
            smtp_user: env::var("SMTP_USER").ok().filter(|s| !s.is_empty()),
            smtp_password: env::var("SMTP_PASSWORD").ok().filter(|s| !s.is_empty()),
            smtp_from: env::var("SMTP_FROM").ok().filter(|s| !s.is_empty()),
            tls_cert_file: env::var("TLS_CERT_FILE").ok().filter(|s| !s.is_empty()),
            tls_key_file: env::var("TLS_KEY_FILE").ok().filter(|s| !s.is_empty()),
            cors_allowed_origins: env::var("CORS_ALLOWED_ORIGINS").unwrap_or_default(),
            port: env::var("PORT").unwrap_or_else(|_| "8080".into()),
            log_format: env::var("LOG_FORMAT").unwrap_or_else(|_| "plain".into()),
        })
    }
}

fn require_env(key: &str) -> Result<String, String> {
    env::var(key).map_err(|_| format!("{key} environment variable is required"))
}
