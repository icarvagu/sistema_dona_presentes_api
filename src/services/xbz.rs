use std::time::Duration;

use reqwest::Client;
use serde::Deserialize;

const XBZ_BASE_URL: &str = "https://api.minhaxbz.com.br:5001";

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct XBZProduct {
    #[serde(rename = "IdProduto")]
    pub id_produto: i32,
    #[serde(rename = "CodigoXbz")]
    pub codigo_xbz: String,
    #[serde(rename = "CodigoAmigavel")]
    pub codigo_amigavel: Option<String>,
    #[serde(rename = "Nome")]
    pub nome: String,
    #[serde(rename = "Descricao")]
    pub descricao: Option<String>,
    #[serde(rename = "ImageLink")]
    pub image_link: Option<String>,
    #[serde(rename = "WebTipo")]
    pub web_tipo: Option<String>,
    #[serde(rename = "QuantidadeDisponivel")]
    pub quantidade_disponivel: Option<i32>,
    #[serde(rename = "Ncm")]
    pub ncm: Option<String>,
    #[serde(rename = "PrecoVenda")]
    pub preco_venda: Option<f64>,
    #[serde(rename = "CorWebPrincipal")]
    pub cor: Option<String>,
}

pub struct XBZService {
    client: Client,
    cnpj: String,
    token: String,
}

impl XBZService {
    pub fn new(cnpj: String, token: String) -> Self {
        Self {
            client: Client::builder()
                .timeout(Duration::from_secs(120))
                .build()
                .expect("Failed to create HTTP client"),
            cnpj,
            token,
        }
    }

    pub async fn get_products(&self) -> Result<Vec<XBZProduct>, crate::error::AppError> {
        let url = format!(
            "{}/api/clientes/GetListaDeProdutos?cnpj={}&token={}",
            XBZ_BASE_URL, self.cnpj, self.token
        );

        let resp = self
            .client
            .get(&url)
            .header("User-Agent", "DonaPresentes/2.0")
            .send()
            .await
            .map_err(|e| crate::error::AppError::internal(format!("XBZ API error: {e}")))?;

        let body = resp
            .text()
            .await
            .map_err(|e| crate::error::AppError::internal(e.to_string()))?;

        let products: Vec<XBZProduct> = serde_json::from_str(&body)
            .map_err(|e| crate::error::AppError::internal(format!("XBZ parse: {e}")))?;

        Ok(products)
    }
}
