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

        let products = parse_products(&body)?;

        Ok(products)
    }
}

fn parse_products(body: &str) -> Result<Vec<XBZProduct>, crate::error::AppError> {
    serde_json::from_str(body)
        .map_err(|e| crate::error::AppError::internal(format!("XBZ parse: {e}")))
}

#[cfg(test)]
mod tests {
    use super::parse_products;

    #[test]
    fn parse_products_reads_valid_payload() {
        let body = r#"[
          {
            "IdProduto": 10,
            "CodigoXbz": "DN100",
            "CodigoAmigavel": "ABC",
            "Nome": "Caneca",
            "Descricao": "Caneca branca",
            "ImageLink": "http://img",
            "WebTipo": "Brindes",
            "QuantidadeDisponivel": 15,
            "Ncm": "123",
            "PrecoVenda": 12.5,
            "CorWebPrincipal": "Branco"
          }
        ]"#;

        let products = parse_products(body).unwrap();
        assert_eq!(products.len(), 1);
        assert_eq!(products[0].id_produto, 10);
        assert_eq!(products[0].codigo_xbz, "DN100");
        assert_eq!(products[0].nome, "Caneca");
        assert_eq!(products[0].preco_venda, Some(12.5));
    }

    #[test]
    fn parse_products_rejects_invalid_json() {
        let err = parse_products("{ nope }").unwrap_err();
        assert!(err.message.contains("XBZ parse"));
    }
}
